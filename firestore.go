package webhook

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

type LineWorksAccessToken struct {
	AccessToken string    `json:"access_token"`
	CreatedAt   time.Time `json:"created_at"`
}

type Firestore struct {
	fsClient *firestore.Client
}

func NewFirestore(fsClient *firestore.Client) *Firestore {
	return &Firestore{
		fsClient: fsClient,
	}
}

func (f *Firestore) LoadAccessToken(ctx context.Context, collectionID string) (string, error) {
	q := f.fsClient.
		Collection(collectionID).
		Where("CreatedAt", ">", time.Now().Add(-20*time.Hour)).
		Limit(1)
	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return "", err
	}
	if len(docs) == 0 {
		return "", nil
	}
	token := &LineWorksAccessToken{}
	if err := docs[0].DataTo(token); err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (f *Firestore) SaveAccessToken(ctx context.Context, collectionID string, token string) error {
	col := f.fsClient.Collection(collectionID)
	doc := col.NewDoc()
	subject := &LineWorksAccessToken{
		AccessToken: token,
		CreatedAt:   time.Now(),
	}
	if _, err := doc.Create(ctx, subject); err != nil {
		return err
	}
	return nil
}
