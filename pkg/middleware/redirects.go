package redirectmiddleware

import (
	"net/http"

	keellog "github.com/foomo/keel/log"
	"github.com/foomo/keel/net/http/middleware"
	redirectprovider "github.com/foomo/redirects/domain/redirectdefinition/provider"
	"go.uber.org/zap"
)

// Redirects middleware
func Redirects(provider *redirectprovider.RedirectsProvider) middleware.Middleware {
	return func(l *zap.Logger, name string, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// only get request will ever be in need of redirects
			if r.Method == http.MethodGet {
				redirect, err := provider.Process(r)
				if err != nil {
					// just log the error and continue with the rest of the middlewares
					// redirect problems should never impair the gateway
					keellog.WithError(l, err).Info("error occurred during redirect processing", keellog.FValue(r.URL.RequestURI()))
				}
				if redirect != nil {
					l.Debug("performing redirect", keellog.FValue(redirect.Response), keellog.FValue(redirect.Code))
					http.Redirect(w, r, string(redirect.Response), int(redirect.Code))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
