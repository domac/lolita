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
	opts *Options

	tcpListener  net.Listener
	httpListener net.Listener
	instanceMap  map[string]*EtcdInstance

	waitGroup util.WaitGroupWrapper
	outchan   chan []byte //数据输出通道
	exitChan  chan int
}

func New(opts *Options) *Lolid {
	l := &Lolid{
		opts:     opts,
		exitChan: make(chan int),
		outchan:  make(chan []byte, opts.MaxWriteChannelSize),
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

func (l *Lolid) RealHTTPAddr() *net.TCPAddr {
	l.RLock()
	defer l.RUnlock()
	return l.httpListener.Addr().(*net.TCPAddr)
}

//主程序入口
func (l *Lolid) Main() {
	ctx := &context{l}
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
		Serve(l.httpListener, httpServer, "HTTP", l.opts.Logger)
	})

	//开启执行调度任务(如果不开启,本程序只可提供基本HTTP api功能)
	if l.opts.OpenTasks {
		l.logf("open tasks")
		l.waitGroup.Wrap(func() { l.lookupOnputTasks() })
		l.waitGroup.Wrap(func() { l.lookupEtcd() })
		l.waitGroup.Wrap(func() { l.lookupInputTasks() })
	}

}

//后台程序退出
func (l *Lolid) Exit() {
	if l.httpListener != nil {
		l.httpListener.Close()
	}

	if l.tcpListener != nil {
		l.tcpListener.Close()
	}
	close(l.outchan)
	close(l.exitChan)
	l.waitGroup.Wait()
}
