package redirectstore

import (
	"github.com/google/uuid"
)

// EntityID base type
type EntityID string

// EntityIDGeneratorFn type
type EntityIDGeneratorFn func() EntityID

// EntityIDGenerator holds the used generator function
var EntityIDGenerator EntityIDGeneratorFn = DefaultEntityIDGenerator

// DefaultEntityIDGenerator represents the default entity generator function
func DefaultEntityIDGenerator() EntityID {
	return EntityID(uuid.New().String())
}

// NewEntityID returns a new id from the EntityGenerator
func NewEntityID() EntityID {
	return EntityIDGenerator()
}
