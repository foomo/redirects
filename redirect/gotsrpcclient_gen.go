// Code generated by gotsrpc https://github.com/foomo/gotsrpc/v2  - DO NOT EDIT.

package redirect

import (
	go_context "context"
	go_net_http "net/http"

	gotsrpc "github.com/foomo/gotsrpc/v2"
	pkg_errors "github.com/pkg/errors"
)

type RedirectGoTSRPCClient interface {
	Create(ctx go_context.Context) (clientErr error)
}

type HTTPRedirectGoTSRPCClient struct {
	URL      string
	EndPoint string
	Client   gotsrpc.Client
}

func NewDefaultRedirectGoTSRPCClient(url string) *HTTPRedirectGoTSRPCClient {
	return NewRedirectGoTSRPCClient(url, "/services/redirects/redirect")
}

func NewRedirectGoTSRPCClient(url string, endpoint string) *HTTPRedirectGoTSRPCClient {
	return NewRedirectGoTSRPCClientWithClient(url, endpoint, nil)
}

func NewRedirectGoTSRPCClientWithClient(url string, endpoint string, client *go_net_http.Client) *HTTPRedirectGoTSRPCClient {
	return &HTTPRedirectGoTSRPCClient{
		URL:      url,
		EndPoint: endpoint,
		Client:   gotsrpc.NewClientWithHttpClient(client),
	}
}
func (tsc *HTTPRedirectGoTSRPCClient) Create(ctx go_context.Context) (clientErr error) {
	args := []interface{}{}
	reply := []interface{}{}
	clientErr = tsc.Client.Call(ctx, tsc.URL, tsc.EndPoint, "Create", args, reply)
	if clientErr != nil {
		clientErr = pkg_errors.WithMessage(clientErr, "failed to call redirect.RedirectGoTSRPCProxy Create")
	}
	return
}
