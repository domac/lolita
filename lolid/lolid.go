package lolid

import (
	"fmt"
	"github.com/domac/lolita/util"
	"github.com/domac/lolita/version"
	"net"
	"os"
	"sync"
)

type Lolid struct {
	sync.RWMutex
	opts         *Options
	httpListener net.Listener
	waitGroup    util.WaitGroupWrapper
}

func New(opts *Options) *Lolid {
	l := &Lolid{
		opts: opts,
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

func (l *Lolid) Main() {
	ctx := &Context{l}
	fmt.Println("lolita love you: ", l.opts.HTTPAddress)
	httpListener, err := net.Listen("tcp", l.opts.HTTPAddress)
	if err != nil {
		l.logf("FATAL: listen (%s) failed - %s", l.opts.HTTPAddress, err)
		os.Exit(1)
	}
	l.Lock()
	l.httpListener = httpListener
	l.Unlock()
	httpServer := newHTTPServer(ctx)
	l.waitGroup.Wrap(func() {
		Serve(httpListener, httpServer, "HTTP", l.opts.Logger)
	})
}

//后台程序退出
func (l *Lolid) Exit() {
	if l.httpListener != nil {
		l.httpListener.Close()
	}
	l.waitGroup.Wait()
}
