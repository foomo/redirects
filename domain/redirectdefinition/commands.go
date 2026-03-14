package redirectdefinition

import (
	commandx "github.com/foomo/redirects/v2/domain/redirectdefinition/command"
)

type Commands struct {
	CreateRedirects      commandx.CreateRedirectsHandlerFn
	CreateRedirect       commandx.CreateRedirectHandlerFn
	UpdateRedirect       commandx.UpdateRedirectHandlerFn
	UpdateRedirectsState commandx.UpdateRedirectsStateHandlerFn
	DeleteRedirect       commandx.DeleteRedirectHandlerFn
}
