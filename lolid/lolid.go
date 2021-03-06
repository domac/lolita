package lolid

import (
	"errors"
	"fmt"
	"github.com/domac/lolita/util"
	"github.com/domac/lolita/version"
	"github.com/pquerna/ffjson/ffjson"
	"net"
	"os"
	"strings"
	"sync"
)

type Lolid struct {
	sync.RWMutex
	opts *Options

	tcpListener  net.Listener
	httpListener net.Listener
	InstanceMap  map[string][]string

	waitGroup                 util.WaitGroupWrapper
	messageCollectStartedChan chan int
	outchan                   chan []byte //数据输出通道
	exitChan                  chan int

	isExit bool
	paused bool
}

func New(opts *Options) *Lolid {
	l := &Lolid{
		opts:                      opts,
		exitChan:                  make(chan int),
		outchan:                   make(chan []byte, opts.MaxWriteChannelSize),
		messageCollectStartedChan: make(chan int),
		InstanceMap:               make(map[string][]string),
		paused:                    false,
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

//清空作业信息
func (l *Lolid) ClearLocalJobs() error {
	l.RLock()
	l.InstanceMap = make(map[string][]string, 0)
	l.RUnlock()
	return nil
}

//刷新当前工作信息
func (l *Lolid) RefleshJobs(jobJson string) error {
	if jobJson == "" {
		l.logf("reflesh job is null")
		l.ClearLocalJobs()
		return errors.New("reflesh job is null")
	}
	l.logf("Agent Jobs [%s] info reflesh !!! \n", jobJson)
	jsonMap := make(map[string]string)

	//校验工作信息的值格式是否正确
	err := ffjson.Unmarshal([]byte(jobJson), &jsonMap)
	if err != nil {
		l.logf("reflesh jobs map error")
		l.ClearLocalJobs()
		return err
	}
	refleshMap := make(map[string][]string, len(jsonMap))
	for key, value := range jsonMap {
		proxys := strings.Split(value, ",")
		refleshMap[key] = proxys
	}
	l.RLock()
	l.InstanceMap = refleshMap
	l.RUnlock()
	return nil
}

func (l *Lolid) Empty() error {
	l.Lock()
	defer l.Unlock()

	for {
		select {
		case <-l.outchan:
		default:
			goto finish
		}
	}
finish:
	return nil
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

		if l.opts.SendType == 0 {
			l.waitGroup.Wrap(func() { l.loopOnputTasks() })
		} else {
			l.waitGroup.Wrap(func() { l.messagesDump() })
		}

		// messageCollectStartedCha用于同步输出与输入的流程
		// 这样可以保证输出器的初始化工作完成后,才进行数据采集的工作
		// 可以避免因为输出器因为某些原因无法工作,导致数据不断采集而无消费
		// 这样容易导致内存消息堆积,引起无法控制的情况
		<-l.messageCollectStartedChan
		l.logf("start doing jobs")
		l.waitGroup.Wrap(func() { l.lookupEtcd() })
		l.waitGroup.Wrap(func() { l.loopInputTasks() })
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
	l.isExit = true
	l.waitGroup.Wait()
}
