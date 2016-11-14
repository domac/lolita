package lolid

import (
	"errors"
	"fmt"
	"github.com/domac/lolita/clients/etcd"
	httpclient "github.com/domac/lolita/clients/http"
	"github.com/domac/lolita/config"
	"math"
	"time"
)

//默认任务间隔
const DEFAULT_TASK_INTERVAL = 1000 * time.Millisecond

//TODO: 主动发现Etcd的配置信息
func (l *Lolid) lookupEtcd() {

	etcdEndpointAddress := l.opts.EtcdEndpoint
	etcd.Init([]string{etcdEndpointAddress})

	serviceName := l.opts.ServiceName
	domainkey := fmt.Sprintf("/agent/domain/%s", serviceName)
	proxykey := fmt.Sprintf("/agent/proxy/%s", serviceName)
	etcdClient := etcd.GetClient()

	//判断etcd是否已经存在目录文件
	if !etcdClient.IsFileExist(domainkey) {
		//若没有,初始化一个
		etcdClient.Set(domainkey, "")
	}

	//获取当前域值
	domainValue, err := etcdClient.Get(domainkey)

	if err != nil {
		l.logf("counld not get domain value from etcd %s \n:", etcdEndpointAddress)
	}

	//判断etcd是否已经存在目录文件
	if !etcdClient.IsFileExist(proxykey) {
		//若没有,初始化一个
		etcdClient.Set(proxykey, "")
	}

	//获取当前proxy值
	proxyValue, err := etcdClient.Get(proxykey)

	if err != nil {
		l.logf("counld not get proxy value from etcd %s \n:", etcdEndpointAddress)
	}
	l.RefleshInstances(domainValue, proxyValue)

	//只监听proxy的目录
	worker, err := etcdClient.CreateWatcher(proxykey)
	if err != nil {
		l.logf("etcd watch fail ....")
		return
	}
	ctx := etcd.GetContext()
	//proxy目录监听
	go func() {
		for {
			resp, err := worker.Next(ctx)
			if err != nil {
				continue
			}
			switch resp.Action {
			case "set", "update": //新增,修改
				l.RefleshInstances(domainValue, resp.Node.Value)
				break
			case "expire", "delete": //过期,删除
				l.RefleshInstances(domainValue, resp.Node.Value)
				break
			default:
			}
		}
	}()
}

//主动发现需要去做的服务
func (l *Lolid) loopInputTasks() {
	//调度定时器
	ticker := time.Tick(DEFAULT_TASK_INTERVAL)
	for {
		select {
		case <-ticker:
			//执行采集输入
			err := l.runInputs()
			if err != nil {
				l.logf(err.Error())
			}
		case <-l.exitChan:
			goto exit
		}
	}
exit:
	l.logf("LOOKUP: closing")
}

//数据并发收集
func (l *Lolid) runInputs() error {
	//模拟采集
	// for i := 0; i < 30; i++ {
	// 	go func(a int) {
	// 		l.Put([]byte(fmt.Sprintf("%d", a)))
	// 	}(i)
	// }
	//fmt.Printf("本次处理任务信息: %v \n", l.InstanceMap)

	if len(l.InstanceMap) == 0 {
		return errors.New("Etcd Instance Map is null, Plesase check connection or ectd dir")
	}

	for k, v := range l.InstanceMap {
		if k == "" || v == nil || len(v) == 0 {
			continue
		}
		for _, p := range v {
			if p == "" {
				continue
			}
			go func(domain, proxy string) {
				//远程获取数据
				data := get_remote_data(domain, proxy)
				if data != nil {
					l.Put(data)
				}

			}(k, p)
		}
	}
	return nil
}

func get_remote_data(url, proxy string) []byte {
	demoClient := httpclient.NewHttpClient()
	demoClient.WithOptions(httpclient.Map{
		"opt_timeout_ms":        500,
		"opt_connecttimeout_ms": 500,
		"opt_proxy":             proxy,
	})
	resp, err := demoClient.Get(url, nil)
	if err != nil {
		return nil
	}
	data, err := resp.ReadAll()
	resp.Body.Close()
	if err != nil {
		return nil
	}
	return data
}

func (l *Lolid) Put(data []byte) error {
	select {
	case l.outchan <- data:
	default:
		data = data[:0]
	}
	return nil
}

//扫描发送通道,并对采集的结果进行发送
func (l *Lolid) loopOnputTasks() {

	config.Init()

	output := config.NewOutput()

	maxWirteBulkSize := l.opts.MaxWriteBulkSize

	//批量bulk
	packets := make([][]byte, 0, maxWirteBulkSize)

	//关闭messageCollectStartedChan, 宣告输出器的初始化工作已经完成
	//其它工作组件可以往下走
	close(l.messageCollectStartedChan)

	for {
		select {
		case data := <-l.outchan:

			if nil != data {
				packets = append(packets, data)
			}

			chanlen := int(math.Min(float64(len(l.outchan)), float64(maxWirteBulkSize)))

			//如果channel的长度还有数据, 批量最多读取maxWirteBulkSize条数据,再合并写出
			//减少系统调用
			//减少网络传输, 提高资源利用率
			for i := 0; i < chanlen; i++ {
				p := <-l.outchan
				if nil != data {
					packets = append(packets, p)
				}
			}

			if len(packets) > 0 {
				//执行输出
				output.Pop(packets)
				//回收包裹空间
				packets = packets[:0]
			}
		case <-l.exitChan:
			goto exit
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
exit:
	l.logf("LOOKUP: closing")
}

//消息集中处理
func (l *Lolid) messagesDump() {
	pipeline := NewPipeline(l)
	pipeline.Dump()
}
