// Code generated by gotsrpc https://github.com/foomo/gotsrpc/v2  - DO NOT EDIT.

package service

import (
	io "io"
	http "net/http"
	time "time"

	github_com_foomo_contentserver_content "github.com/foomo/contentserver/content"
	gotsrpc "github.com/foomo/gotsrpc/v2"
	github_com_foomo_redirects_domain_redirectdefinition_store "github.com/foomo/redirects/domain/redirectdefinition/store"
)

const (
	RedirectDefinitionServiceGoTSRPCProxyCreate                                 = "Create"
	RedirectDefinitionServiceGoTSRPCProxyCreateRedirectsFromContentserverexport = "CreateRedirectsFromContentserverexport"
	RedirectDefinitionServiceGoTSRPCProxyDelete                                 = "Delete"
	RedirectDefinitionServiceGoTSRPCProxyGetRedirects                           = "GetRedirects"
	RedirectDefinitionServiceGoTSRPCProxySearch                                 = "Search"
	RedirectDefinitionServiceGoTSRPCProxyUpdate                                 = "Update"
)

type RedirectDefinitionServiceGoTSRPCProxy struct {
	EndPoint string
	service  RedirectDefinitionService
}

func NewDefaultRedirectDefinitionServiceGoTSRPCProxy(service RedirectDefinitionService) *RedirectDefinitionServiceGoTSRPCProxy {
	return NewRedirectDefinitionServiceGoTSRPCProxy(service, "/services/redirects/redirectdefinition")
}

func NewRedirectDefinitionServiceGoTSRPCProxy(service RedirectDefinitionService, endpoint string) *RedirectDefinitionServiceGoTSRPCProxy {
	return &RedirectDefinitionServiceGoTSRPCProxy{
		EndPoint: endpoint,
		service:  service,
	}
}

// ServeHTTP exposes your service
func (p *RedirectDefinitionServiceGoTSRPCProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	} else if r.Method != http.MethodPost {
		gotsrpc.ErrorMethodNotAllowed(w)
		return
	}
	defer io.Copy(io.Discard, r.Body) // Drain Request Body

	funcName := gotsrpc.GetCalledFunc(r, p.EndPoint)
	callStats, _ := gotsrpc.GetStatsForRequest(r)
	callStats.Func = funcName
	callStats.Package = "github.com/foomo/redirects/domain/redirectdefinition/service"
	callStats.Service = "RedirectDefinitionService"
	switch funcName {
	case RedirectDefinitionServiceGoTSRPCProxyCreate:
		var (
			args []interface{}
			rets []interface{}
		)
		var (
			arg_def *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition
		)
		args = []interface{}{&arg_def}
		if err := gotsrpc.LoadArgs(&args, callStats, r); err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		createRet := p.service.Create(arg_def)
		callStats.Execution = time.Since(executionStart)
		rets = []interface{}{createRet}
		if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
			gotsrpc.ErrorCouldNotReply(w)
			return
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	case RedirectDefinitionServiceGoTSRPCProxyCreateRedirectsFromContentserverexport:
		var (
			args []interface{}
			rets []interface{}
		)
		var (
			arg_old map[string]*github_com_foomo_contentserver_content.RepoNode
			arg_new map[string]*github_com_foomo_contentserver_content.RepoNode
		)
		args = []interface{}{&arg_old, &arg_new}
		if err := gotsrpc.LoadArgs(&args, callStats, r); err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		createRedirectsFromContentserverexportRet := p.service.CreateRedirectsFromContentserverexport(arg_old, arg_new)
		callStats.Execution = time.Since(executionStart)
		rets = []interface{}{createRedirectsFromContentserverexportRet}
		if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
			gotsrpc.ErrorCouldNotReply(w)
			return
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	case RedirectDefinitionServiceGoTSRPCProxyDelete:
		var (
			args []interface{}
			rets []interface{}
		)
		var (
			arg_id string
		)
		args = []interface{}{&arg_id}
		if err := gotsrpc.LoadArgs(&args, callStats, r); err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		deleteRet := p.service.Delete(arg_id)
		callStats.Execution = time.Since(executionStart)
		rets = []interface{}{deleteRet}
		if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
			gotsrpc.ErrorCouldNotReply(w)
			return
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	case RedirectDefinitionServiceGoTSRPCProxyGetRedirects:
		var (
			args []interface{}
			rets []interface{}
		)
		executionStart := time.Now()
		getRedirectsRet, getRedirectsRet_1 := p.service.GetRedirects()
		callStats.Execution = time.Since(executionStart)
		rets = []interface{}{getRedirectsRet, getRedirectsRet_1}
		if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
			gotsrpc.ErrorCouldNotReply(w)
			return
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	case RedirectDefinitionServiceGoTSRPCProxySearch:
		var (
			args []interface{}
			rets []interface{}
		)
		var (
			arg_dimension string
			arg_id        string
			arg_path      string
		)
		args = []interface{}{&arg_dimension, &arg_id, &arg_path}
		if err := gotsrpc.LoadArgs(&args, callStats, r); err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		searchRet, searchRet_1 := p.service.Search(arg_dimension, arg_id, arg_path)
		callStats.Execution = time.Since(executionStart)
		rets = []interface{}{searchRet, searchRet_1}
		if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
			gotsrpc.ErrorCouldNotReply(w)
			return
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	case RedirectDefinitionServiceGoTSRPCProxyUpdate:
		var (
			args []interface{}
			rets []interface{}
		)
		var (
			arg_def *github_com_foomo_redirects_domain_redirectdefinition_store.RedirectDefinition
		)
		args = []interface{}{&arg_def}
		if err := gotsrpc.LoadArgs(&args, callStats, r); err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		updateRet := p.service.Update(arg_def)
		callStats.Execution = time.Since(executionStart)
		rets = []interface{}{updateRet}
		if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
			gotsrpc.ErrorCouldNotReply(w)
			return
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	default:
		gotsrpc.ClearStats(r)
		gotsrpc.ErrorFuncNotFound(w)
	}
}
