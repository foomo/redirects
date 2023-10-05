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
	RedirectsDefinitionRepository struct {
		l          *zap.Logger
		collection *keelmongo.Collection
	}
)

func NewRedirectsDefinitionRepository(l *zap.Logger, persistor *keelmongo.Persistor) (rs *RedirectsDefinitionRepository, err error) {
	collection, cErr := persistor.Collection(
		"redirects",
		keelmongo.CollectionWithIndexes(
			mongo.IndexModel{
				Keys: bson.M{
					"source": 1,
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
		collection: collection}, nil
}

func (rs RedirectsDefinitionRepository) Find(ctx context.Context, source string) (*redirectstore.RedirectDefinition, error) {
	var result redirectstore.RedirectDefinition
	findErr := rs.collection.FindOne(ctx, bson.M{"source": source}, &result)
	if findErr != nil {
		return nil, findErr
	}
	return &result, nil
}

func (rs RedirectsDefinitionRepository) Insert(ctx context.Context, def *redirectstore.RedirectDefinition) error {
	_, err := rs.collection.Col().InsertOne(ctx, def)
	return err
}

func (rs RedirectsDefinitionRepository) Update(ctx context.Context, def *redirectstore.RedirectDefinition) error {
	filter := bson.D{{Key: "source", Value: def.Source}}
	update := bson.D{{Key: "$set", Value: def}}

	_, err := rs.collection.Col().UpdateOne(ctx, filter, update)
	return err

}

// maybe will be needed for migrating manual redirections?
func (rs RedirectsDefinitionRepository) UpsertMany(ctx context.Context, defs []*redirectstore.RedirectDefinition) error {

	var operations []mongo.WriteModel

	for _, def := range defs {
		operation := mongo.NewUpdateOneModel()
		operation.SetFilter(bson.M{
			"source": def.Source,
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

func (rs RedirectsDefinitionRepository) Delete(ctx context.Context, source string) error {
	filter := bson.D{{Key: "source", Value: source}}

	_, err := rs.collection.Col().DeleteOne(ctx, filter)
	return err
}
