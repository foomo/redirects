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
	InternalServiceGoTSRPCProxyCreateRedirectsFromContentserverexport = "CreateRedirectsFromContentserverexport"
	InternalServiceGoTSRPCProxyGetRedirects                           = "GetRedirects"
)

type InternalServiceGoTSRPCProxy struct {
	EndPoint string
	service  InternalService
}

func NewDefaultInternalServiceGoTSRPCProxy(service InternalService) *InternalServiceGoTSRPCProxy {
	return NewInternalServiceGoTSRPCProxy(service, "/services/redirectdefinition/internal")
}

func NewInternalServiceGoTSRPCProxy(service InternalService, endpoint string) *InternalServiceGoTSRPCProxy {
	return &InternalServiceGoTSRPCProxy{
		EndPoint: endpoint,
		service:  service,
	}
}

// ServeHTTP exposes your service
func (p *InternalServiceGoTSRPCProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	callStats.Service = "InternalService"
	switch funcName {
	case InternalServiceGoTSRPCProxyCreateRedirectsFromContentserverexport:
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
		rw := gotsrpc.ResponseWriter{ResponseWriter: w}
		createRedirectsFromContentserverexportRet := p.service.CreateRedirectsFromContentserverexport(&rw, r, arg_old, arg_new)
		callStats.Execution = time.Since(executionStart)
		if rw.Status() == http.StatusOK {
			rets = []interface{}{createRedirectsFromContentserverexportRet}
			if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
				gotsrpc.ErrorCouldNotReply(w)
				return
			}
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	case InternalServiceGoTSRPCProxyGetRedirects:
		var (
			args []interface{}
			rets []interface{}
		)
		executionStart := time.Now()
		rw := gotsrpc.ResponseWriter{ResponseWriter: w}
		getRedirectsRet, getRedirectsRet_1 := p.service.GetRedirects(&rw, r)
		callStats.Execution = time.Since(executionStart)
		if rw.Status() == http.StatusOK {
			rets = []interface{}{getRedirectsRet, getRedirectsRet_1}
			if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
				gotsrpc.ErrorCouldNotReply(w)
				return
			}
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	default:
		gotsrpc.ClearStats(r)
		gotsrpc.ErrorFuncNotFound(w)
	}
}

const (
	AdminServiceGoTSRPCProxyCreate = "Create"
	AdminServiceGoTSRPCProxyDelete = "Delete"
	AdminServiceGoTSRPCProxySearch = "Search"
	AdminServiceGoTSRPCProxyUpdate = "Update"
)

type AdminServiceGoTSRPCProxy struct {
	EndPoint string
	service  AdminService
}

func NewDefaultAdminServiceGoTSRPCProxy(service AdminService) *AdminServiceGoTSRPCProxy {
	return NewAdminServiceGoTSRPCProxy(service, "/services/redirectdefinition/admin")
}

func NewAdminServiceGoTSRPCProxy(service AdminService, endpoint string) *AdminServiceGoTSRPCProxy {
	return &AdminServiceGoTSRPCProxy{
		EndPoint: endpoint,
		service:  service,
	}
}

// ServeHTTP exposes your service
func (p *AdminServiceGoTSRPCProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	callStats.Service = "AdminService"
	switch funcName {
	case AdminServiceGoTSRPCProxyCreate:
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
		rw := gotsrpc.ResponseWriter{ResponseWriter: w}
		createRet := p.service.Create(&rw, r, arg_def)
		callStats.Execution = time.Since(executionStart)
		if rw.Status() == http.StatusOK {
			rets = []interface{}{createRet}
			if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
				gotsrpc.ErrorCouldNotReply(w)
				return
			}
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	case AdminServiceGoTSRPCProxyDelete:
		var (
			args []interface{}
			rets []interface{}
		)
		var (
			arg_path string
		)
		args = []interface{}{&arg_path}
		if err := gotsrpc.LoadArgs(&args, callStats, r); err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		rw := gotsrpc.ResponseWriter{ResponseWriter: w}
		deleteRet := p.service.Delete(&rw, r, arg_path)
		callStats.Execution = time.Since(executionStart)
		if rw.Status() == http.StatusOK {
			rets = []interface{}{deleteRet}
			if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
				gotsrpc.ErrorCouldNotReply(w)
				return
			}
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	case AdminServiceGoTSRPCProxySearch:
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
		rw := gotsrpc.ResponseWriter{ResponseWriter: w}
		searchRet, searchRet_1 := p.service.Search(&rw, r, arg_dimension, arg_id, arg_path)
		callStats.Execution = time.Since(executionStart)
		if rw.Status() == http.StatusOK {
			rets = []interface{}{searchRet, searchRet_1}
			if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
				gotsrpc.ErrorCouldNotReply(w)
				return
			}
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	case AdminServiceGoTSRPCProxyUpdate:
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
		rw := gotsrpc.ResponseWriter{ResponseWriter: w}
		updateRet := p.service.Update(&rw, r, arg_def)
		callStats.Execution = time.Since(executionStart)
		if rw.Status() == http.StatusOK {
			rets = []interface{}{updateRet}
			if err := gotsrpc.Reply(rets, callStats, r, w); err != nil {
				gotsrpc.ErrorCouldNotReply(w)
				return
			}
		}
		gotsrpc.Monitor(w, r, args, rets, callStats)
		return
	default:
		gotsrpc.ClearStats(r)
		gotsrpc.ErrorFuncNotFound(w)
	}
}
