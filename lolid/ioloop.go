package lolid

import (
	"errors"
	"fmt"
	"github.com/domac/lolita/clients/etcd"
	httpclient "github.com/domac/lolita/clients/http"
	"github.com/domac/lolita/config"
	"github.com/domac/lolita/util"
	"math"
	"os"
	"strings"
	"time"
)

//默认任务间隔
const DEFAULT_TASK_INTERVAL = 1000 * time.Millisecond
const HRARTBEAT_INTERVAL = 1000 * time.Millisecond
const LUA_READ_TIMEOUT = 500 * time.Millisecond

//TODO: 主动发现Etcd的配置信息
func (l *Lolid) lookupEtcd() {

	etcdEndpointAddress := l.opts.EtcdEndpoint

	endpoints := strings.Split(etcdEndpointAddress, ",")

	//etcd环境初始化
	err := etcd.Init(endpoints)
	if err != nil {
		l.logf("counld not connections etcd %v \n:", endpoints)
		l.Exit()
		os.Exit(2)
		return
	}

	//获取etcd客户端实例
	etcdClient := etcd.GetClient()

	//etcd 上下文
	ctx := etcd.GetContext()

	//agent识别信息
	agentId := l.opts.AgentId
	agentGroup := l.opts.AgentGroup

	//作业目录
	jobsFile := fmt.Sprintf("/apus/agent-groups/%s/jobs", agentGroup)
	if !etcdClient.IsFileExist(jobsFile) {
		//若没有,初始化一个
		etcdClient.Set(jobsFile, "")
	}

	//Leader目录
	leaderFile := fmt.Sprintf("/apus/agent-groups/%s/leader", agentGroup)
	if !etcdClient.IsFileExist(leaderFile) {
		//如果所在组的Leader文件不存在抢占成为leader
		etcdClient.Set(leaderFile, agentId)
	} else {
		currentLeader, _ := etcdClient.Get(leaderFile)
		if currentLeader != agentId {
			l.paused = true
		}
	}

	//成员目录
	memberAgentDir := fmt.Sprintf("/apus/agent-groups/%s/members/%s", agentGroup, agentId)
	if !etcdClient.IsDirExist(memberAgentDir) {
		//若没有,创建一个目录
		etcdClient.MakeDir(memberAgentDir)
	}

	//心跳目录
	heartbeatFile := fmt.Sprintf("/apus/agent-groups/%s/members/%s/heartbeat", agentGroup, agentId)
	if !etcdClient.IsFileExist(heartbeatFile) {
		//若没有,初始化一个(这个过程真正完成服务注册会向etcd发送一个Create事件表示服务需要被发现)
		//HA模块会监听Create事件,用来确认有新的agent进来了
		etcdClient.CreateDir(heartbeatFile)
	}

	jobInfo, err := etcdClient.Get(jobsFile)
	if err != nil {
		l.logf("counld not get jobs from etcd %s:%s \n:", etcdEndpointAddress, jobsFile)
	}

	//初始化当前作业
	l.RefleshJobs(jobInfo)

	localip := ""

	//获取本地IP
	localIps, err := util.IntranetIP()
	if err == nil {
		localip = localIps[0]
	}

	//心跳间隔
	hbInterval := time.Tick(HRARTBEAT_INTERVAL)

	//异步发送心跳到etcd
	go func() {
		for {
			select {
			case <-hbInterval:
				etcdClient.Set(heartbeatFile, TouchHeart(localip))
			case <-l.exitChan:
				break
			}
		}
	}()

	//监听leader的配置
	leaderWorker, err := etcdClient.CreateWatcher(leaderFile)

	if err != nil {
		l.logf("etcd leader watch fail ...")
		return
	}

	go func() {
		for {
			resp, err := leaderWorker.Next(ctx)
			if err != nil {
				continue
			}
			switch resp.Action {
			case "set", "update": //新增,修改
				currentLeader := resp.Node.Value
				if currentLeader != agentId {
					l.logf("leader is changed !")
					l.paused = true
				} else {
					l.paused = false
				}
			case "expire", "delete": //过期,删除
				l.paused = false
			default:
			}
		}
	}()

	//监听jobs的目录
	worker, err := etcdClient.CreateWatcher(jobsFile)
	if err != nil {
		l.logf("etcd job watch fail ....")
		return
	}

	//jobs目录监听
	go func() {
		for {
			resp, err := worker.Next(ctx)
			if err != nil {
				continue
			}
			switch resp.Action {
			case "set", "update": //新增,修改
				l.RefleshJobs(resp.Node.Value)
				break
			case "expire", "delete": //过期,删除
				l.RefleshJobs(resp.Node.Value)
				break
			default:
			}
		}
	}()

	//监听自身的注册目录
	agentworker, err := etcdClient.CreateWatcher(memberAgentDir)
	if err != nil {
		l.logf("etcd self watch fail ....")
		return
	}
	go func() {
		for {
			resp, err := agentworker.Next(ctx)
			if err != nil {
				continue
			}
			switch resp.Action {
			case "set", "update", "expire": //新增,修改
				break
			case "delete": //过期,删除
				//若删除,agent就退出运行吧
				l.logf("etcd send away signal.... \n")
				l.RefleshJobs("")
				l.Exit()
				os.Exit(2)
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
