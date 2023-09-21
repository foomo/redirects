package redirection

import (
	"context"

	keelmongo "github.com/foomo/keel/persistence/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

type RedirectsStore struct {
	l          *zap.Logger
	persistor  *keelmongo.Persistor
	collection *keelmongo.Collection
}

func NewRedirectsStore(l *zap.Logger, persistor *keelmongo.Persistor) (rs *RedirectsStore, err error) {
	collection, cErr := persistor.Collection(
		"redirects",
		keelmongo.CollectionWithIndexes(
			mongo.IndexModel{
				Keys: bson.M{
					"id": 1,
				},
				Options: options.Index().SetUnique(true),
			},
		),
	)

	if cErr != nil {
		return nil, cErr
	}
	return &RedirectsStore{
		l:          l,
		persistor:  persistor,
		collection: collection}, nil
}

func (rs RedirectsStore) Find(ctx context.Context, id string) (*RedirectDefinition, error) {
	// TODO: Implement
	return nil, nil
}

func (rs RedirectsStore) Insert(ctx context.Context, def *RedirectDefinition) error {
	// TODO: Implement
	return nil
}

func (rs RedirectsStore) Update(ctx context.Context, def *RedirectDefinition) error {
	// TODO: Implement
	return nil
}

func (rs RedirectsStore) Delete(ctx context.Context, id string) error {
	// TODO: Implement
	return nil
}
