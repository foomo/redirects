package redirectdefinition

import (
	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
)

type Commands struct {
	CreateRedirects redirectcommand.CreateRedirectsHandlerFn
	CreateRedirect  redirectcommand.CreateRedirectHandlerFn
	UpdateRedirect  redirectcommand.UpdateRedirectHandlerFn
	DeleteRedirect  redirectcommand.DeleteRedirectHandlerFn
}
