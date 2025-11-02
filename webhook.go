package webhook

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/joho/godotenv"
)

var bot *LineWorksBot

func init() {
	_ = godotenv.Load()

	pk, err := GetSecret(
		os.Getenv("GCP_PROJECT"),
		os.Getenv("PRIVATE_KEY_SECRET_NAME"),
	)
	if err != nil {
		log.Fatal(err)
	}

	bot = NewLineWorksBot(
		os.Getenv("BOT_ID"),
		os.Getenv("CLIENT_ID"),
		os.Getenv("CLIENT_SECRET"),
		os.Getenv("SERVICE_ACCOUNT"),
		pk,
	)

	functions.HTTP("Notify", notify)
}

func getOrCreateToken(ctx context.Context) (string, error) {
	col := os.Getenv("COLLECTION_ID")

	fsClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		return "", err
	}
	defer fsClient.Close()

	fs := NewFirestore(fsClient)
	token, err := fs.LoadAccessToken(ctx, col)
	if err != nil {
		return "", err
	}
	if token != "" {
		fmt.Println("Token Exists")
		return token, nil
	}

	token, err = bot.GenerateAccessToken()
	if err != nil {
		return "", err
	}
	fmt.Println("Token Created")

	if err := fs.SaveAccessToken(ctx, col, token); err != nil {
		return "", err
	}
	fmt.Println("Token Saved")

	return token, nil
}

func notify(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	retryKey := r.FormValue("retry_key")
	if retryKey != "" {
		fsClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
		if err != nil {
			return
		}
		defer fsClient.Close()
		fs := NewFirestore(fsClient)
		retryKeyCol := "RetryKey"
		key, err := fs.LoadRetryKey(ctx, retryKeyCol, retryKey)
		if err == nil && key != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("skip %s", key.CreatedAt.Format(time.RFC3339))))
			return
		}
		fs.SaveRetryKey(ctx, retryKeyCol, retryKey)
	}
	channelID := r.FormValue("channel_id")
	text := r.FormValue("text")
	token, err := getOrCreateToken(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error: %v", err)))
		return
	}

	if err := bot.SendTextMessage(token, channelID, text); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error: %v", err)))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}
}
