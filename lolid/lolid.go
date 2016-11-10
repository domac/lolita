package lolid

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/domac/lolita/util"
	"github.com/domac/lolita/version"
)

type Lolid struct {
	sync.RWMutex
	opts         *Options
	httpListener net.Listener
	exitChan     chan int
	waitGroup    util.WaitGroupWrapper
}

func New(opts *Options) *Lolid {
	l := &Lolid{
		opts:     opts,
		exitChan: make(chan int),
	}
	l.logf(version.String("LOLID"))
	return l
}

func (l *Lolid) logf(f string, args ...interface{}) {
	if l.opts.Logger == nil {
		return
	}
	l.opts.Logger.Output(2, fmt.Sprintf(f, args...))
}

//主程序入口
func (l *Lolid) Main() {
	ctx := &Context{l}
	httpListener, err := net.Listen("tcp", l.opts.HTTPAddress)
	if err != nil {
		l.logf("FATAL: listen (%s) failed - %s", l.opts.HTTPAddress, err)
		os.Exit(1)
	}
	l.Lock()
	l.httpListener = httpListener
	l.Unlock()
	httpServer := newHTTPServer(ctx)
	//开启对外提供的http服务
	l.waitGroup.Wrap(func() {
		Serve(httpListener, httpServer, "HTTP", l.opts.Logger)
	})
	//开启执行任务
	if l.opts.OpenTasks {
		l.waitGroup.Wrap(func() { l.lookupTasks() })
	}

}

//后台程序退出
func (l *Lolid) Exit() {
	if l.httpListener != nil {
		l.httpListener.Close()
	}
	close(l.exitChan)
	l.waitGroup.Wait()
}
