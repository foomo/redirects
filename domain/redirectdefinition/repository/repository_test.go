package redirectrepository_test

import (
	"context"
	"fmt"
	"testing"

	keelmongo "github.com/foomo/keel/persistence/mongo"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.uber.org/zap"
)

func TestGetAllRedirects(t *testing.T) {
	l := zap.L()
	mongoURI := "mongodb://localhost:27017/local"
	remotePersistor, err := keelmongo.New(
		context.Background(),
		mongoURI,
		keelmongo.WithOtelOptions(
			otelmongo.WithCommandAttributeDisabled(true),
		),
	)
	assert.NoError(t, err)
	// create repository
	repo, err := redirectrepository.NewRedirectsDefinitionRepository(l, remotePersistor)
	if err != nil {
		fmt.Print(err)
	}
	redirectDefinitions, err := repo.FindAll(context.Background())
	assert.Equal(t, 0, len(*redirectDefinitions))
	assert.NoError(t, err)

}
