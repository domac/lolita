package lolid

import (
	"fmt"
	"github.com/domac/lolita/config"
	"math"
	"time"
)

//默认任务间隔
const DEFAULT_TASK_INTERVAL = 1000 * time.Millisecond

//TODO: 主动发现Etcd的配置信息
func (l *Lolid) lookupEtcd() {

}

//主动发现需要去做的服务
func (l *Lolid) lookupInputTasks() {
	//调度定时器
	ticker := time.Tick(DEFAULT_TASK_INTERVAL)
	for {
		select {
		case <-ticker:
			//执行采集输入
			l.runInputs()
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
	for i := 0; i < 100; i++ {
		go func(a int) {
			l.Put([]byte(fmt.Sprintf("%d", a)))
		}(i)
	}
	return nil
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
func (l *Lolid) lookupOnputTasks() {

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
