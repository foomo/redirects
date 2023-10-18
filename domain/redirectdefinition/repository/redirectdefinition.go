package redirectrepository

import (
	"context"

	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"

	keelmongo "github.com/foomo/keel/persistence/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (rs RedirectsDefinitionRepository) Find(ctx context.Context, id, source string) (*redirectstore.RedirectDefinition, error) {
	var result redirectstore.RedirectDefinition
	findErr := rs.collection.FindOne(ctx, bson.M{"id": id, "source": source}, &result)
	if findErr != nil {
		return nil, findErr
	}
	return &result, nil
}

// TODO: DraganaB check if we need to search by id
func (rs RedirectsDefinitionRepository) FindMany(ctx context.Context, id, source string) (*redirectstore.RedirectDefinitions, error) {
	var result redirectstore.RedirectDefinitions

	// Create a regex pattern for fuzzy match
	pattern := primitive.Regex{Pattern: source, Options: "i"} // "i" for case-insensitive match

	// Create a filter with the regex pattern
	filter := bson.M{"source": primitive.Regex{Pattern: pattern.Pattern, Options: pattern.Options}}

	findErr := rs.collection.FindOne(ctx, filter, &result)
	if findErr != nil {
		return nil, findErr
	}
	return &result, nil
}

func (rs RedirectsDefinitionRepository) FindAll(ctx context.Context) (defs *redirectstore.RedirectDefinitions, err error) {
	err = rs.collection.Find(ctx, bson.M{}, &defs)
	if err != nil {
		return nil, err
	}
	return defs, nil
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
func (rs RedirectsDefinitionRepository) UpsertMany(ctx context.Context, defs *redirectstore.RedirectDefinitions) error {

	var operations []mongo.WriteModel

	for source, def := range *defs {
		operation := mongo.NewUpdateOneModel()
		operation.SetFilter(bson.M{
			"source": source,
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

func (rs RedirectsDefinitionRepository) DeleteMany(ctx context.Context, sources []redirectstore.RedirectSource) error {
	filter := bson.M{"source": bson.M{"$in": sources}}

	_, err := rs.collection.Col().DeleteMany(ctx, filter)
	return err
}
