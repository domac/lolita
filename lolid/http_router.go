package lolid

import (
	"net/http"

	"bytes"
	"github.com/domac/lolita/util"
	"github.com/domac/lolita/version"
	"github.com/julienschmidt/httprouter"
	"net/http/pprof"
)

type httpServer struct {
	ctx    *context
	router http.Handler
}

//HTTP 服务
func newHTTPServer(ctx *context) *httpServer {

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

	//内置监控
	router.GET("/debug/pprof/*pprof", innerPprofHandler)

	//在这里注册路由服务
	router.Handle("GET", "/version", Decorate(s.versionHandler, log, Default))
	router.Handle("GET", "/debug", Decorate(s.pprofHandler, log, PlainText))
	router.Handle("GET", "/ping", Decorate(s.pingHandler, log, PlainText))
	router.Handle("GET", "/empty", Decorate(s.emptyHandler, log, PlainText))
	router.Handle("POST", "/pong", Decorate(s.pongHandler, log, Default))

	return s
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

//调用内置的pprof
func innerPprofHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	switch p.ByName("pprof") {
	case "/cmdline":
		pprof.Cmdline(w, r)
	case "/profile":
		pprof.Profile(w, r)
	case "/symbol":
		pprof.Symbol(w, r)
	default:
		pprof.Index(w, r)
	}
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

//输出性能信息
func (s *httpServer) pprofHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	paramReq, err := NewReqParams(req)
	if err != nil {
		return nil, err
	}
	cmd, _ := paramReq.Get("cmd")
	buf := bytes.Buffer{}
	util.ProcessInput(cmd, &buf)
	return buf.String(), nil
}

//通道数据清空
func (s *httpServer) emptyHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) (interface{}, error) {
	s.ctx.lolid.Empty()
	return "empty is finish", nil
}
