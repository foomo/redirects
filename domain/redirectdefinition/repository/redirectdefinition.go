package redirectrepository

import (
	"context"

	keelmongo "github.com/foomo/keel/persistence/mongo"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type (
	RedirectsDefinitionRepository interface {
		FindOne(ctx context.Context, id, source string) (*redirectstore.RedirectDefinition, error)
		FindMany(ctx context.Context, source, dimension string, onlyActive bool) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error)
		FindAll(ctx context.Context, onlyActive bool) (defs map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, err error)
		Insert(ctx context.Context, def *redirectstore.RedirectDefinition) error
		Update(ctx context.Context, def *redirectstore.RedirectDefinition) error
		UpsertMany(ctx context.Context, defs []*redirectstore.RedirectDefinition) error
		Delete(ctx context.Context, id redirectstore.EntityID) error
		DeleteMany(ctx context.Context, ids []redirectstore.EntityID) error
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

func NewBaseRedirectsDefinitionRepository(l *zap.Logger, persistor *keelmongo.Persistor) (*BaseRedirectsDefinitionRepository, error) {
	collection, cErr := persistor.Collection(
		"redirects",
		keelmongo.CollectionWithIndexes(
			mongo.IndexModel{
				Keys: bson.D{
					{Key: "source", Value: 1},
					{Key: "dimension", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
		),
	)

	_, _ = collection.Col().Indexes().DropOne(context.TODO(), "source_1")

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

func (rs BaseRedirectsDefinitionRepository) FindMany(ctx context.Context, source, dimension string, onlyActive bool) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
	var result []*redirectstore.RedirectDefinition
	filter := bson.M{}

	if source != "" {
		// Create a regex pattern for fuzzy match
		pattern := primitive.Regex{Pattern: source, Options: "i"} // "i" for case-insensitive match
		filter["source"] = primitive.Regex{Pattern: pattern.Pattern, Options: pattern.Options}
	}

	if dimension != "" {
		filter["dimension"] = dimension
	}

	if onlyActive {
		filter["stale"] = false
	}

	findErr := rs.collection.Find(ctx, filter, &result)
	var retResult = make(map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition)
	for _, red := range result {
		if _, ok := retResult[red.Source]; !ok {
			retResult[red.Source] = red
		}
	}

	if findErr != nil {
		return nil, findErr
	}
	return retResult, nil
}

func (rs BaseRedirectsDefinitionRepository) FindAll(ctx context.Context, onlyActive bool) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
	var result []redirectstore.RedirectDefinition
	filter := bson.M{}
	if onlyActive {
		filter["stale"] = false
	}

	err := rs.collection.Find(ctx, bson.M{}, &result)
	if err != nil {
		return nil, err
	}
	var retResult = make(map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition)
	for _, res := range result {
		resCopy := res
		if _, ok := retResult[res.Dimension]; !ok {
			retResult[res.Dimension] = make(map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition)
		}
		retResult[res.Dimension][res.Source] = &resCopy
	}
	return retResult, nil
}

func (rs BaseRedirectsDefinitionRepository) Insert(ctx context.Context, def *redirectstore.RedirectDefinition) error {
	if def.ID == "" {
		def.ID = redirectstore.NewEntityID()
	}
	_, err := rs.collection.Col().InsertOne(ctx, def)
	return err
}

func (rs BaseRedirectsDefinitionRepository) Update(ctx context.Context, def *redirectstore.RedirectDefinition) error {
	filter := bson.D{{Key: "id", Value: def.ID}}
	update := bson.D{{Key: "$set", Value: def}}

	_, err := rs.collection.Col().UpdateOne(ctx, filter, update)
	return err
}

// maybe will be needed for migrating manual redirections?
func (rs BaseRedirectsDefinitionRepository) UpsertMany(ctx context.Context, defs []*redirectstore.RedirectDefinition) error {
	operations := make([]mongo.WriteModel, 0, len(defs))

	for _, def := range defs {
		if def.ID == "" {
			def.ID = redirectstore.NewEntityID()
		}
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

func (rs BaseRedirectsDefinitionRepository) Delete(ctx context.Context, id redirectstore.EntityID) error {
	filter := bson.D{{Key: "id", Value: id}}
	_, err := rs.collection.Col().DeleteOne(ctx, filter)
	return err
}

func (rs BaseRedirectsDefinitionRepository) DeleteMany(ctx context.Context, ids []redirectstore.EntityID) error {
	filter := bson.M{"id": bson.M{"$in": ids}}
	_, err := rs.collection.Col().DeleteMany(ctx, filter)
	return err
}
