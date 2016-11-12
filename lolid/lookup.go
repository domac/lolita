package lolid

import (
	"errors"
	"fmt"
	"github.com/domac/lolita/config"
	"math"
	"time"
)

//默认任务间隔
const DEFAULT_TASK_INTERVAL = 3000 * time.Millisecond

//主动发现Etcd的配置信息
func (l *Lolid) lookupEtcd() {
}

//主动发现需要去做的服务
func (l *Lolid) lookupInputTasks() {
	ticker := time.Tick(DEFAULT_TASK_INTERVAL)
	for {
		select {
		case <-ticker:
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
			time.Sleep(300 * time.Millisecond)
			l.outchan <- []byte(fmt.Sprintf("%d", a))
		}(i)
	}
	return nil
}

//扫描发送通道,并对采集的结果进行发送
func (l *Lolid) lookupOnputTasks() {
	//获取输出器
	outputs, err := l.getOutputs()
	if err != nil {
		panic(err)
	}

	maxWirteBulkSize := l.opts.MaxWriteBulkSize

	//批量bulk
	packets := make([][]byte, 0, maxWirteBulkSize)

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
				l.runOutputs(outputs, packets)
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

func (l *Lolid) runOutputs(outputs []config.TypeOutputConfig, packets [][]byte) error {
	if packets == nil {
		return errors.New("data null")
	}
	fmt.Printf("==== %d\n", len(packets))
	return nil
}

func (l *Lolid) getOutputs() (outputs []config.TypeOutputConfig, err error) {
	rawConfig, err := l.GetConfigs()
	if err != nil {
		return nil, errors.New("no config")
	}
	configOutputs := rawConfig["outputs"].([]map[string]interface{})
	if configOutputs == nil {
		return nil, errors.New("no outputs config")
	}
	return nil, nil
}
