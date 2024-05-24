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
	RedirectsDefinitionRepository interface {
		FindOne(ctx context.Context, id, source string) (*redirectstore.RedirectDefinition, error)
		FindMany(ctx context.Context, id, source, dimension string) (*redirectstore.RedirectDefinitions, error)
		FindAll(ctx context.Context) (defs *redirectstore.RedirectDefinitions, err error)
		Insert(ctx context.Context, def *redirectstore.RedirectDefinition) error
		Update(ctx context.Context, def *redirectstore.RedirectDefinition) error
		UpsertMany(ctx context.Context, defs *redirectstore.RedirectDefinitions) error
		Delete(ctx context.Context, source, dimension string) error
		DeleteMany(ctx context.Context, sources []redirectstore.RedirectSource, dimension string) error
	}
	BaseRedirectsDefinitionRepository struct {
		l          *zap.Logger
		collection *keelmongo.Collection
	}
)

func NewRedirectsDefinitionRepository(l *zap.Logger, collection *keelmongo.Collection) *BaseRedirectsDefinitionRepository {
	return &BaseRedirectsDefinitionRepository{
		l:          l,
		collection: collection,
	}
}

func NewBaseRedirectsDefinitionRepository(l *zap.Logger, persistor *keelmongo.Persistor) (rs *BaseRedirectsDefinitionRepository, err error) {
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
	return NewRedirectsDefinitionRepository(l, collection), nil
}

func (rs BaseRedirectsDefinitionRepository) FindOne(ctx context.Context, id, source string) (*redirectstore.RedirectDefinition, error) {
	var result redirectstore.RedirectDefinition
	findErr := rs.collection.FindOne(ctx, bson.M{"id": id, "source": source}, &result)
	if findErr != nil {
		return nil, findErr
	}
	return &result, nil
}

// TODO: DraganaB check if we need to search by id
func (rs BaseRedirectsDefinitionRepository) FindMany(ctx context.Context, id, source, dimension string) (*redirectstore.RedirectDefinitions, error) {
	var result redirectstore.RedirectDefinitions

	// Create a regex pattern for fuzzy match
	pattern := primitive.Regex{Pattern: source, Options: "i"} // "i" for case-insensitive match

	// Create a filter with the regex pattern
	filter := bson.M{"source": primitive.Regex{Pattern: pattern.Pattern, Options: pattern.Options}, "dimension": dimension}

	findErr := rs.collection.FindOne(ctx, filter, &result)
	if findErr != nil {
		return nil, findErr
	}
	return &result, nil
}

func (rs BaseRedirectsDefinitionRepository) FindAll(ctx context.Context) (*redirectstore.RedirectDefinitions, error) {
	var result []redirectstore.RedirectDefinition
	err := rs.collection.Find(ctx, bson.M{}, &result)
	if err != nil {
		return nil, err
	}
	var retResult = make(redirectstore.RedirectDefinitions)
	for _, res := range result {
		retResult[res.Source] = make(map[redirectstore.Dimension]*redirectstore.RedirectDefinition)
		retResult[res.Source][res.Dimension] = &res
	}
	return &retResult, nil
}

func (rs BaseRedirectsDefinitionRepository) Insert(ctx context.Context, def *redirectstore.RedirectDefinition) error {
	_, err := rs.collection.Col().InsertOne(ctx, def)
	return err
}

func (rs BaseRedirectsDefinitionRepository) Update(ctx context.Context, def *redirectstore.RedirectDefinition) error {
	filter := bson.D{{Key: "source", Value: def.Source}, {Key: "dimension", Value: def.Dimension}}
	update := bson.D{{Key: "$set", Value: def}}

	_, err := rs.collection.Col().UpdateOne(ctx, filter, update)
	return err

}

// maybe will be needed for migrating manual redirections?
func (rs BaseRedirectsDefinitionRepository) UpsertMany(ctx context.Context, defs *redirectstore.RedirectDefinitions) error {

	var operations []mongo.WriteModel

	for source, defByDimension := range *defs {
		for _, def := range defByDimension {
			operation := mongo.NewUpdateOneModel()
			operation.SetFilter(bson.M{
				"source":    source,
				"dimension": def.Dimension,
			})
			operation.SetUpdate(bson.D{{Key: "$set", Value: def}})
			operation.SetUpsert(true)
			operations = append(operations, operation)
		}

	}
	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(false)

	_, err := rs.collection.Col().BulkWrite(ctx, operations, &bulkOption)
	if err != nil {
		return err
	}

	return err
}

func (rs BaseRedirectsDefinitionRepository) Delete(ctx context.Context, source, dimension string) error {
	filter := bson.D{{Key: "source", Value: source}, {Key: "dimension", Value: dimension}}
	_, err := rs.collection.Col().DeleteOne(ctx, filter)
	return err
}

func (rs BaseRedirectsDefinitionRepository) DeleteMany(ctx context.Context, sources []redirectstore.RedirectSource, dimension string) error {
	filter := bson.M{"source": bson.M{"$in": sources}, "dimension": dimension}
	_, err := rs.collection.Col().DeleteMany(ctx, filter)
	return err
}
