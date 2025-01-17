package redirectrepository

import (
	"context"
	"fmt"
	"time"

	keelmongo "github.com/foomo/keel/persistence/mongo"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type SortField string

const (
	SortFieldSource        SortField = "source"
	SortFieldUpdated       SortField = "updated"
	SortFieldLastUpdatedBy SortField = "lastUpdatedBy"
)

type Direction string

const (
	DirectionAscending  Direction = "ascending"
	DirectionDescending Direction = "descending"
)

type Sort struct {
	Field     SortField `json:"field"`
	Direction Direction `json:"direction"`
}

func (d Direction) GetSortValue() int {
	switch d {
	case DirectionAscending:
		return 1
	case DirectionDescending:
		return -1
	default:
		return 1
	}
}

type PaginatedResult struct {
	Results  []*redirectstore.RedirectDefinition `json:"results"`
	Total    int                                 `json:"total"`
	Page     int                                 `json:"page"`
	PageSize int                                 `json:"pageSize"`
}

type (
	RedirectsDefinitionRepository interface {
		FindOne(ctx context.Context, id, source string) (*redirectstore.RedirectDefinition, error)
		FindMany(ctx context.Context, source, dimension, redirectType string, onlyActive bool, pagination Pagination, sort Sort) (*PaginatedResult, error)
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
			mongo.IndexModel{
				Keys: bson.D{
					{Key: string(SortFieldUpdated), Value: 1},
				},
			},
			mongo.IndexModel{
				Keys: bson.D{
					{Key: string(SortFieldLastUpdatedBy), Value: 1},
				},
			},
			// Index for 'source' field (optional for search optimization)
			mongo.IndexModel{
				Keys: bson.D{
					{Key: string(SortFieldSource), Value: 1},
				},
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

func (rs BaseRedirectsDefinitionRepository) FindMany(
	ctx context.Context,
	source, dimension, redirectType string,
	onlyActive bool,
	pagination Pagination,
	sort Sort,
) (*PaginatedResult, error) {
	// Validate pagination
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 20 // Default page size
	}

	var result []*redirectstore.RedirectDefinition
	filter := bson.M{}

	// Apply filters
	if source != "" {
		pattern := primitive.Regex{Pattern: source, Options: "i"} // Case-insensitive regex
		filter["source"] = pattern
	}
	if dimension != "" {
		filter["dimension"] = dimension
	}
	if redirectType != "" {
		filter["redirectType"] = redirectType
	}
	if onlyActive {
		filter["stale"] = false
	}

	// Pagination settings
	skip := (pagination.Page - 1) * pagination.PageSize
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pagination.PageSize))

	// Sorting settings
	sortField := sort.Field
	if sortField == "" {
		sortField = SortFieldSource // Default sort field
	}
	opts.SetSort(bson.D{
		{Key: string(sortField), Value: sort.Direction.GetSortValue()},
		{Key: "_id", Value: 1}, // Tie-breaker for consistent results
	})

	// Query MongoDB
	cursor, err := rs.collection.Col().Find(ctx, filter, opts)
	if err != nil {
		return &PaginatedResult{}, err
	}
	defer cursor.Close(ctx)

	// Decode results
	for cursor.Next(ctx) {
		var red redirectstore.RedirectDefinition
		if err := cursor.Decode(&red); err != nil {
			return &PaginatedResult{}, err
		}
		result = append(result, &red)
	}

	total, err := rs.collection.Col().CountDocuments(ctx, filter)
	if err != nil {
		return &PaginatedResult{}, err
	}

	return &PaginatedResult{
		Results:  result,
		Total:    int(total),
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}, nil
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

func (rs BaseRedirectsDefinitionRepository) UpsertMany(ctx context.Context, defs []*redirectstore.RedirectDefinition) error {
	chunkSize := 1000
	retries := 3

	for i := 0; i < len(defs); i += chunkSize {
		end := i + chunkSize
		if end > len(defs) {
			end = len(defs)
		}

		err := rs.upsertChunkWithRetry(ctx, defs[i:end], retries)
		if err != nil {
			rs.l.Error("failed to upsert chunk",
				zap.Int("start", i),
				zap.Int("end", end),
				zap.Error(err))
			return err
		}
	}
	return nil
}

func (rs BaseRedirectsDefinitionRepository) upsertChunkWithRetry(ctx context.Context, defs []*redirectstore.RedirectDefinition, retries int) error {
	for i := 0; i < retries; i++ {
		err := rs.upsertChunk(ctx, defs)
		if err == nil {
			return nil // Success
		}

		// Log the retry and wait briefly before retrying
		rs.l.Info("Retrying chunk upsert...", zap.Int("retry no", i+1), zap.Error(err))
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("failed to upsert chunk after %d retries", retries)
}

func (rs BaseRedirectsDefinitionRepository) upsertChunk(ctx context.Context, defs []*redirectstore.RedirectDefinition) error {
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

	result, err := rs.collection.Col().BulkWrite(ctx, operations, &bulkOption)
	if err != nil {
		rs.l.Error("Bulk write error", zap.Error(err))
		return err
	}

	// Log results
	rs.l.Info("Bulk write result",
		zap.Int("Inserted", int(result.InsertedCount)),
		zap.Int("Matched", int(result.MatchedCount)),
		zap.Int("Modified", int(result.ModifiedCount)),
		zap.Int("Upserted", len(result.UpsertedIDs)))
	return nil
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
