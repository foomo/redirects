package redirectrepository_test

// "context"
// _ "embed"
// "encoding/json"
// "fmt"
// "testing"

// "github.com/foomo/contentserver/content"
// keelmongo "github.com/foomo/keel/persistence/mongo"
// redirectapi "github.com/foomo/redirects/v2/domain/redirectdefinition"
// redirectcommand "github.com/foomo/redirects/v2/domain/redirectdefinition/command"
// redirectrepository "github.com/foomo/redirects/v2/domain/redirectdefinition/repository"
// "github.com/stretchr/testify/assert"
// "go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
// "go.uber.org/zap"

// //go:embed content.json
// var contentNodes []byte

// //go:embed content-changed.json
// var contentNodesChanged []byte

// func TestGetAllRedirects(t *testing.T) {
// 	l := zap.L()
// 	mongoURI := "mongodb://localhost:27017/local"
// 	remotePersistor, err := keelmongo.New(
// 		context.Background(),
// 		mongoURI,
// 		keelmongo.WithOtelOptions(
// 			otelmongo.WithCommandAttributeDisabled(true),
// 		),
// 	)
// 	assert.NoError(t, err)
// 	// create repository
// 	repo, err := redirectrepository.NewBaseRedirectsDefinitionRepository(l, remotePersistor)
// 	if err != nil {
// 		fmt.Print(err)
// 	}
// 	redirectDefinitions, err := repo.FindAll(context.Background(), false)
// 	assert.Equal(t, 1, len(redirectDefinitions))
// 	assert.NoError(t, err)
// }
// func TestGenerateAutoRedirects(t *testing.T) {
// 	l := zap.L()
// 	mongoURI := "mongodb://localhost:27017/local"
// 	remotePersistor, err := keelmongo.New(
// 		context.Background(),
// 		mongoURI,
// 		keelmongo.WithOtelOptions(
// 			otelmongo.WithCommandAttributeDisabled(true),
// 		),
// 	)
// 	assert.NoError(t, err)
// 	// create repository
// 	repo, err := redirectrepository.NewBaseRedirectsDefinitionRepository(l, remotePersistor)
// 	if err != nil {
// 		fmt.Print(err)
// 	}

// 	api, err := redirectapi.NewAPI(l, *repo, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	p := map[string]*content.RepoNode{}
// 	err = json.Unmarshal(contentNodes, &p)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	pChanged := map[string]*content.RepoNode{}
// 	err = json.Unmarshal(contentNodesChanged, &pChanged)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	err = api.CreateRedirects(context.Background(), redirectcommand.CreateRedirects{
// 		OldState: p, NewState: pChanged,
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	redirectDefinitions, err := repo.FindAll(context.Background(), false)
// 	assert.Equal(t, 116, len(redirectDefinitions["de"]))
// 	assert.NoError(t, err)
// }
