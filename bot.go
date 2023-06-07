package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/xerrors"
)

const TokenURL = "https://auth.worksmobile.com/oauth2/v2.0/token"
const MessageURLPattern = "https://www.worksapis.com/v1.0/bots/%s/channels/%s/messages"

type LineWorksBot struct {
	BotID          string
	ClientID       string
	ClientSecret   string
	ServiceAccount string
	PrivateKey     []byte
}

type TokenResponse struct {
	ErrorMessage string `json:"message"`
	ErrorCode    string `json:"code"`
	ErrorDetail  string `json:"detail"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
}

type TextMessage struct {
	Content TextContent `json:"content"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewLineWorksBot(botID, clientID, clientSecret, serviceAccount string, privateKey []byte) *LineWorksBot {
	return &LineWorksBot{
		BotID:          botID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		ServiceAccount: serviceAccount,
		PrivateKey:     privateKey,
	}
}

func (bot *LineWorksBot) GenerateAccessToken() (string, error) {
	tok, err := jwt.NewBuilder().
		Subject(bot.ServiceAccount).
		Issuer(bot.ClientID).
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(time.Hour)).
		Build()
	if err != nil {
		return "", err
	}

	privkey, err := jwk.ParseKey(bot.PrivateKey, jwk.WithPEM(true))
	if err != nil {
		return "", err
	}

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, privkey))
	if err != nil {
		return "", err
	}

	v := url.Values{}
	v.Add("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	v.Add("assertion", string(signed))
	v.Add("client_id", bot.ClientID)
	v.Add("client_secret", bot.ClientSecret)
	v.Add("scope", "bot")

	req, err := http.NewRequest(http.MethodPost, TokenURL, strings.NewReader(v.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	client.Timeout = time.Second * 30
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 200 {
		b, _ := httputil.DumpResponse(resp, true)
		return "", xerrors.Errorf("failure get token: %s", string(b))
	}

	t := &TokenResponse{}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return "", err
	}

	if t.ErrorCode != "" {
		return "", xerrors.Errorf("failure get token: %+v", t)
	}

	return t.AccessToken, nil
}

func (bot *LineWorksBot) SendTextMessage(token, channelID, text string) error {
	url := fmt.Sprintf(MessageURLPattern, bot.BotID, channelID)
	txtMsg := &TextMessage{
		Content: TextContent{
			Type: "text",
			Text: text,
		},
	}
	b, err := json.Marshal(txtMsg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b1, _ := httputil.DumpRequest(req, true)
		b2, _ := httputil.DumpResponse(resp, true)
		return xerrors.Errorf("fail SendTextMessage %s\n\n%s", string(b1), string(b2))
	}

	return nil
}
