// Code generated by gotsrpc https://github.com/foomo/gotsrpc/v2  - DO NOT EDIT.

package service

import (
	go_context "context"
	go_net_http "net/http"

	github_com_foomo_contentserver_content "github.com/foomo/contentserver/content"
	gotsrpc "github.com/foomo/gotsrpc/v2"
	github_com_foomo_redirects_domain_redirectdefinition_store "github.com/foomo/redirects/domain/redirectdefinition/store"
	pkg_errors "github.com/pkg/errors"
)

type InternalServiceGoTSRPCClient interface {
	CreateRedirectsFromContentserverexport(ctx go_context.Context, old map[string]*github_com_foomo_contentserver_content.RepoNode, new map[string]*github_com_foomo_contentserver_content.RepoNode) (retCreateRedirectsFromContentserverexport_0 error, clientErr error)
	GetRedirects(ctx go_context.Context) (retGetRedirects_0 map[github_com_foomo_redirects_domain_redirectdefinition_store.Dimension]map[github_com_foomo_redirects_domain_redirectdefinition_store.RedirectSource]*github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition, retGetRedirects_1 error, clientErr error)
}

type HTTPInternalServiceGoTSRPCClient struct {
	URL      string
	EndPoint string
	Client   gotsrpc.Client
}

func NewDefaultInternalServiceGoTSRPCClient(url string) *HTTPInternalServiceGoTSRPCClient {
	return NewInternalServiceGoTSRPCClient(url, "/services/redirectdefinition/internal")
}

func NewInternalServiceGoTSRPCClient(url string, endpoint string) *HTTPInternalServiceGoTSRPCClient {
	return NewInternalServiceGoTSRPCClientWithClient(url, endpoint, nil)
}

func NewInternalServiceGoTSRPCClientWithClient(url string, endpoint string, client *go_net_http.Client) *HTTPInternalServiceGoTSRPCClient {
	return &HTTPInternalServiceGoTSRPCClient{
		URL:      url,
		EndPoint: endpoint,
		Client:   gotsrpc.NewClientWithHttpClient(client),
	}
}
func (tsc *HTTPInternalServiceGoTSRPCClient) CreateRedirectsFromContentserverexport(ctx go_context.Context, old map[string]*github_com_foomo_contentserver_content.RepoNode, new map[string]*github_com_foomo_contentserver_content.RepoNode) (retCreateRedirectsFromContentserverexport_0 error, clientErr error) {
	args := []interface{}{old, new}
	reply := []interface{}{&retCreateRedirectsFromContentserverexport_0}
	clientErr = tsc.Client.Call(ctx, tsc.URL, tsc.EndPoint, "CreateRedirectsFromContentserverexport", args, reply)
	if clientErr != nil {
		clientErr = pkg_errors.WithMessage(clientErr, "failed to call service.InternalServiceGoTSRPCProxy CreateRedirectsFromContentserverexport")
	}
	return
}

func (tsc *HTTPInternalServiceGoTSRPCClient) GetRedirects(ctx go_context.Context) (retGetRedirects_0 map[github_com_foomo_redirects_domain_redirectdefinition_store.Dimension]map[github_com_foomo_redirects_domain_redirectdefinition_store.RedirectSource]*github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition, retGetRedirects_1 error, clientErr error) {
	args := []interface{}{}
	reply := []interface{}{&retGetRedirects_0, &retGetRedirects_1}
	clientErr = tsc.Client.Call(ctx, tsc.URL, tsc.EndPoint, "GetRedirects", args, reply)
	if clientErr != nil {
		clientErr = pkg_errors.WithMessage(clientErr, "failed to call service.InternalServiceGoTSRPCProxy GetRedirects")
	}
	return
}

type AdminServiceGoTSRPCClient interface {
	Create(ctx go_context.Context, def *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition, locale string) (retCreate_0 github_com_foomo_redirects_domain_redirectdefinition_store.RedirectID, retCreate_1 *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinitionError, clientErr error)
	Delete(ctx go_context.Context, path string, dimension string) (retDelete_0 *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinitionError, clientErr error)
	Search(ctx go_context.Context, locale string, path string) (retSearch_0 map[github_com_foomo_redirects_domain_redirectdefinition_store.RedirectSource]*github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition, retSearch_1 *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinitionError, clientErr error)
	Update(ctx go_context.Context, def *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition) (retUpdate_0 *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinitionError, clientErr error)
}

type HTTPAdminServiceGoTSRPCClient struct {
	URL      string
	EndPoint string
	Client   gotsrpc.Client
}

func NewDefaultAdminServiceGoTSRPCClient(url string) *HTTPAdminServiceGoTSRPCClient {
	return NewAdminServiceGoTSRPCClient(url, "/services/redirectdefinition/admin")
}

func NewAdminServiceGoTSRPCClient(url string, endpoint string) *HTTPAdminServiceGoTSRPCClient {
	return NewAdminServiceGoTSRPCClientWithClient(url, endpoint, nil)
}

func NewAdminServiceGoTSRPCClientWithClient(url string, endpoint string, client *go_net_http.Client) *HTTPAdminServiceGoTSRPCClient {
	return &HTTPAdminServiceGoTSRPCClient{
		URL:      url,
		EndPoint: endpoint,
		Client:   gotsrpc.NewClientWithHttpClient(client),
	}
}
func (tsc *HTTPAdminServiceGoTSRPCClient) Create(ctx go_context.Context, def *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition, locale string) (retCreate_0 github_com_foomo_redirects_domain_redirectdefinition_store.RedirectID, retCreate_1 *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinitionError, clientErr error) {
	args := []interface{}{def, locale}
	reply := []interface{}{&retCreate_0, &retCreate_1}
	clientErr = tsc.Client.Call(ctx, tsc.URL, tsc.EndPoint, "Create", args, reply)
	if clientErr != nil {
		clientErr = pkg_errors.WithMessage(clientErr, "failed to call service.AdminServiceGoTSRPCProxy Create")
	}
	return
}

func (tsc *HTTPAdminServiceGoTSRPCClient) Delete(ctx go_context.Context, path string, dimension string) (retDelete_0 *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinitionError, clientErr error) {
	args := []interface{}{path, dimension}
	reply := []interface{}{&retDelete_0}
	clientErr = tsc.Client.Call(ctx, tsc.URL, tsc.EndPoint, "Delete", args, reply)
	if clientErr != nil {
		clientErr = pkg_errors.WithMessage(clientErr, "failed to call service.AdminServiceGoTSRPCProxy Delete")
	}
	return
}

func (tsc *HTTPAdminServiceGoTSRPCClient) Search(ctx go_context.Context, locale string, path string) (retSearch_0 map[github_com_foomo_redirects_domain_redirectdefinition_store.RedirectSource]*github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition, retSearch_1 *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinitionError, clientErr error) {
	args := []interface{}{locale, path}
	reply := []interface{}{&retSearch_0, &retSearch_1}
	clientErr = tsc.Client.Call(ctx, tsc.URL, tsc.EndPoint, "Search", args, reply)
	if clientErr != nil {
		clientErr = pkg_errors.WithMessage(clientErr, "failed to call service.AdminServiceGoTSRPCProxy Search")
	}
	return
}

func (tsc *HTTPAdminServiceGoTSRPCClient) Update(ctx go_context.Context, def *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition) (retUpdate_0 *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinitionError, clientErr error) {
	args := []interface{}{def}
	reply := []interface{}{&retUpdate_0}
	clientErr = tsc.Client.Call(ctx, tsc.URL, tsc.EndPoint, "Update", args, reply)
	if clientErr != nil {
		clientErr = pkg_errors.WithMessage(clientErr, "failed to call service.AdminServiceGoTSRPCProxy Update")
	}
	return
}
