package webhook

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

type RetryKey struct {
	UUID      string    `json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
}

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

func (f *Firestore) LoadRetryKey(ctx context.Context, collectionID string, uuid string) (*RetryKey, error) {
	q := f.fsClient.
		Collection(collectionID).
		Where("UUID", "==", uuid).
		Where("CreatedAt", ">", time.Now().Add(-1*time.Hour)).
		OrderBy("CreatedAt", firestore.Desc).
		Limit(1)
	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}
	subject := &RetryKey{}
	if err := docs[0].DataTo(subject); err != nil {
		return nil, err
	}
	return subject, nil
}

func (f *Firestore) SaveRetryKey(ctx context.Context, collectionID string, uuid string) error {
	col := f.fsClient.Collection(collectionID)
	doc := col.NewDoc()
	subject := &RetryKey{
		UUID:      uuid,
		CreatedAt: time.Now(),
	}
	if _, err := doc.Create(ctx, subject); err != nil {
		return err
	}
	return nil
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
