package redirectrepository

import (
	"context"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"

	keelmongo "github.com/foomo/keel/persistence/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type (
	RSI interface {
		Find(ctx context.Context, id string) (*redirectstore.RedirectDefinition, error)
	}
	BaseRedirectsDefinition struct {
		l          *zap.Logger
		collection *keelmongo.Collection
	}
)

func NewRedirectsStore(l *zap.Logger, persistor *keelmongo.Persistor) (rs *RedirectsDefinitionRepository, err error) {
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
	return &RedirectsDefinitionRepository{
		l:          l,
		persistor:  persistor,
		collection: collection}, nil
}

func (rs RedirectsDefinitionRepository) Find(ctx context.Context, id string) (*RedirectDefinition, error) {
	var result RedirectDefinition
	findErr := rs.collection.FindOne(ctx, bson.M{"id": id}, &result)
	if findErr != nil {
		return nil, findErr
	}
	return &result, nil
}

func (rs RedirectsDefinitionRepository) Insert(ctx context.Context, def *RedirectDefinition) error {
	_, err := rs.collection.Col().InsertOne(ctx, def)
	return err
}

func (rs RedirectsDefinitionRepository) Update(ctx context.Context, def *RedirectDefinition) error {
	filter := bson.D{{Key: "id", Value: def.ID}}
	update := bson.D{{Key: "$set", Value: def}}

	_, err := rs.collection.Col().UpdateOne(ctx, filter, update)
	return err

}

// maybe will be needed for migrating manual redirections?
func (rs RedirectsDefinitionRepository) UpsertMany(ctx context.Context, defs []*RedirectDefinition) error {

	var operations []mongo.WriteModel

	for _, def := range defs {
		operation := mongo.NewUpdateOneModel()
		operation.SetFilter(bson.M{
			"id": def.ID,
		})
		operation.SetUpdate(bson.D{{Key: "$set", Value: def}})
		operation.SetUpsert(true)
		operations = append(operations, operation)
	}
	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(false)

	_, err := rs.collection.Col().BulkWrite(ctx, operations, &bulkOption)
	if err != nil {
		return err
	}

	return err
}

func (rs RedirectsDefinitionRepository) Delete(ctx context.Context, id string) error {
	filter := bson.D{{Key: "id", Value: id}}

	_, err := rs.collection.Col().DeleteOne(ctx, filter)
	return err
}
