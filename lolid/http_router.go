package lolid

import (
	"net/http"

	"github.com/domac/lolita/version"
	"github.com/julienschmidt/httprouter"
)

type httpServer struct {
	ctx    *Context
	router http.Handler
}

//HTTP 服务
func newHTTPServer(ctx *Context) *httpServer {

	log := Log(ctx.lolid.opts.Logger)

	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.PanicHandler = LogPanicHandler(ctx.lolid.opts.Logger)
	router.NotFound = LogNotFoundHandler(ctx.lolid.opts.Logger)
	router.MethodNotAllowed = LogMethodNotAllowedHandler(ctx.lolid.opts.Logger)

	s := &httpServer{
		ctx:    ctx,
		router: router,
	}

	//在这里注册路由服务
	router.Handle("GET", "/version", Decorate(s.versionHandler, log, Default))
	router.Handle("GET", "/ping", Decorate(s.pingHandler, log, PlainText))
	router.Handle("POST", "/pong", Decorate(s.pongHandler, log, Default))
	return s
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

func (s *httpServer) versionHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	version := version.String("LOLID")
	res := NewResult(RESULT_CODE_FAIL, true, "", version)
	return res, nil
}

//Ping
func (s *httpServer) pingHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	return "OK", nil
}

//Pong
func (s *httpServer) pongHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	pp := &PostParams{req}
	param, err := pp.Get("pong")
	if err != nil {
		return "no param", err
	}
	s.ctx.lolid.opts.Logger.Output(2, param)
	res := NewResult(RESULT_CODE_FAIL, true, "PONG!", param)
	return res, nil
}
