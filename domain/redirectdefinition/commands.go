package redirectdefinition

import (
	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
)

type Commands struct {
	CreateRedirects      redirectcommand.CreateRedirectsHandlerFn
	CreateRedirect       redirectcommand.CreateRedirectHandlerFn
	UpdateRedirect       redirectcommand.UpdateRedirectHandlerFn
	UpdateRedirectsState redirectcommand.UpdateRedirectsStateHandlerFn
	DeleteRedirect       redirectcommand.DeleteRedirectHandlerFn
}
