package redirectrepository

import (
	"context"
	"fmt"
	"time"

	keelmongo "github.com/foomo/keel/persistence/mongo"
	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type (
	RedirectsDefinitionRepository interface {
		FindOne(ctx context.Context, id, source string) (*storex.RedirectDefinition, error)
		FindMany(ctx context.Context, source, dimension string, redirectType storex.RedirectionType, activeState storex.ActiveStateType, pagination storex.Pagination, sort storex.Sort) (*storex.PaginatedResult, error)
		FindAll(ctx context.Context, onlyActive bool) (map[storex.Dimension]map[storex.RedirectSource]*storex.RedirectDefinition, error)
		FindAllByDimension(ctx context.Context, dimension storex.Dimension, onlyActive bool) (map[storex.RedirectSource]*storex.RedirectDefinition, error)
		Insert(ctx context.Context, def *storex.RedirectDefinition) error
		Update(ctx context.Context, def *storex.RedirectDefinition) error
		UpsertMany(ctx context.Context, defs []*storex.RedirectDefinition) error
		FindByIDs(ctx context.Context, ids []*storex.EntityID) ([]*storex.RedirectDefinition, error)
		Delete(ctx context.Context, id storex.EntityID) error
		DeleteMany(ctx context.Context, ids []storex.EntityID) error
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
					{Key: string(storex.SortFieldUpdated), Value: 1},
				},
			},
			mongo.IndexModel{
				Keys: bson.D{
					{Key: string(storex.SortFieldLastUpdatedBy), Value: 1},
				},
			},
			// Index for 'source' field (optional for search optimization)
			mongo.IndexModel{
				Keys: bson.D{
					{Key: string(storex.SortFieldSource), Value: 1},
				},
			},
		),
	)
	if cErr != nil {
		return nil, cErr
	}

	return NewRedirectsDefinitionRepository(l, collection), nil
}

func (rs *BaseRedirectsDefinitionRepository) FindOne(ctx context.Context, id, source string) (*storex.RedirectDefinition, error) {
	var result storex.RedirectDefinition

	findErr := rs.collection.FindOne(ctx, bson.M{"id": id, "source": source}, &result)
	if findErr != nil {
		return nil, findErr
	}

	return &result, nil
}

func (rs *BaseRedirectsDefinitionRepository) FindMany(
	ctx context.Context,
	source, dimension string,
	redirectType storex.RedirectionType,
	activeState storex.ActiveStateType,
	pagination storex.Pagination,
	sort storex.Sort,
) (*storex.PaginatedResult, error) {
	// Validate pagination
	if pagination.Page < 1 {
		pagination.Page = 1
	}

	if pagination.PageSize < 1 {
		pagination.PageSize = 20 // Default page size
	}

	var result []*storex.RedirectDefinition

	filter := bson.M{}

	// Apply filters
	if source != "" {
		pattern := bson.Regex{Pattern: source, Options: "i"} // Case-insensitive regex
		filter["source"] = pattern
	}

	if dimension != "" {
		filter["dimension"] = dimension
	}

	// Apply redirect type filter
	if redirectValue, apply := redirectType.ToFilter(); apply {
		filter["redirectType"] = redirectValue
	}

	// Apply active state filter
	if stateValue, apply := activeState.ToFilter(); apply {
		filter["stale"] = stateValue
	}

	// Pagination settings
	skip := (pagination.Page - 1) * pagination.PageSize
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pagination.PageSize))

	// Sorting settings
	sortField := sort.Field
	if sortField == "" {
		sortField = storex.SortFieldSource // Default sort field
	}

	opts.SetSort(bson.D{
		{Key: string(sortField), Value: sort.Direction.GetSortValue()},
		{Key: "_id", Value: 1}, // Tie-breaker for consistent results
	})

	// Query MongoDB
	cursor, err := rs.collection.Col().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode results
	for cursor.Next(ctx) {
		var red storex.RedirectDefinition
		if err := cursor.Decode(&red); err != nil {
			return nil, err
		}

		result = append(result, &red)
	}

	total, err := rs.collection.Col().CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &storex.PaginatedResult{
		Results:  result,
		Total:    int(total),
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}, nil
}

func (rs *BaseRedirectsDefinitionRepository) FindAll(ctx context.Context, onlyActive bool) (map[storex.Dimension]map[storex.RedirectSource]*storex.RedirectDefinition, error) {
	var results []storex.RedirectDefinition

	filter := bson.M{}

	if onlyActive {
		filter["stale"] = false
	}

	cursor, err := rs.collection.Col().Find(ctx, filter)
	if err != nil {
		rs.l.Error("Failed to fetch redirects", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode all documents into results
	if err := cursor.All(ctx, &results); err != nil {
		rs.l.Error("Failed to decode redirect results", zap.Error(err))
		return nil, err
	}

	// Convert results into the expected map format
	retResult := make(map[storex.Dimension]map[storex.RedirectSource]*storex.RedirectDefinition)

	for _, res := range results {
		resCopy := res // Create a copy to avoid pointer issues

		if _, exists := retResult[res.Dimension]; !exists {
			retResult[res.Dimension] = make(map[storex.RedirectSource]*storex.RedirectDefinition)
		}

		retResult[res.Dimension][res.Source] = &resCopy
	}

	return retResult, nil
}

func (rs *BaseRedirectsDefinitionRepository) FindAllByDimension(ctx context.Context, dimension storex.Dimension, onlyActive bool) (map[storex.RedirectSource]*storex.RedirectDefinition, error) {
	var results []storex.RedirectDefinition

	filter := bson.M{"dimension": dimension}

	// If onlyActive is true, fetch only non-stale (active) redirects
	if onlyActive {
		filter["stale"] = false
	}

	err := rs.collection.Find(ctx, filter, &results)
	if err != nil {
		return nil, err
	}

	defs := make(map[storex.RedirectSource]*storex.RedirectDefinition)

	for _, def := range results {
		defCopy := def
		defs[def.Source] = &defCopy
	}

	return defs, nil
}

func (rs *BaseRedirectsDefinitionRepository) Insert(ctx context.Context, def *storex.RedirectDefinition) error {
	if def.ID == "" {
		def.ID = storex.NewEntityID()
	}

	_, err := rs.collection.Col().InsertOne(ctx, def)

	return err
}

func (rs *BaseRedirectsDefinitionRepository) Update(ctx context.Context, def *storex.RedirectDefinition) error {
	filter := bson.D{{Key: "id", Value: def.ID}}
	update := bson.D{{Key: "$set", Value: def}}

	_, err := rs.collection.Col().UpdateOne(ctx, filter, update)

	return err
}

func (rs *BaseRedirectsDefinitionRepository) UpsertMany(ctx context.Context, defs []*storex.RedirectDefinition) error {
	chunkSize := 1000
	retries := 3

	for i := 0; i < len(defs); i += chunkSize {
		end := min(i+chunkSize, len(defs))

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

func (rs *BaseRedirectsDefinitionRepository) Delete(ctx context.Context, id storex.EntityID) error {
	filter := bson.D{{Key: "id", Value: id}}
	_, err := rs.collection.Col().DeleteOne(ctx, filter)

	return err
}

func (rs *BaseRedirectsDefinitionRepository) DeleteMany(ctx context.Context, ids []storex.EntityID) error {
	filter := bson.M{"id": bson.M{"$in": ids}}
	_, err := rs.collection.Col().DeleteMany(ctx, filter)

	return err
}

func (rs *BaseRedirectsDefinitionRepository) FindByIDs(ctx context.Context, ids []*storex.EntityID) ([]*storex.RedirectDefinition, error) {
	var results []*storex.RedirectDefinition

	// Query all redirects matching the given IDs
	err := rs.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}}, &results)
	if err != nil {
		rs.l.Error("Failed to fetch redirects by IDs", zap.Error(err))
		return nil, err
	}

	return results, nil
}

func (rs *BaseRedirectsDefinitionRepository) upsertChunkWithRetry(ctx context.Context, defs []*storex.RedirectDefinition, retries int) error {
	for i := range retries {
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

func (rs *BaseRedirectsDefinitionRepository) upsertChunk(ctx context.Context, defs []*storex.RedirectDefinition) error {
	operations := make([]mongo.WriteModel, 0, len(defs))

	for _, def := range defs {
		if def.ID == "" {
			def.ID = storex.NewEntityID()
		}

		operation := mongo.NewUpdateOneModel()
		operation.SetFilter(bson.M{
			"id": def.ID,
		})
		operation.SetUpdate(bson.D{{Key: "$set", Value: def}})
		operation.SetUpsert(true)
		operations = append(operations, operation)
	}

	bulkOption := options.BulkWrite().SetOrdered(false)

	result, err := rs.collection.Col().BulkWrite(ctx, operations, bulkOption)
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
