package lolid

import (
	"errors"
	"github.com/domac/lolita/config"
	"time"
)

//默认任务间隔
const DEFAULT_TASK_INTERVAL = 1000 * time.Millisecond

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
	return nil
}

//扫描发送通道,并对采集的结果进行发送
func (l *Lolid) lookupOnputTasks() {
	//获取输出器
	outputs, err := l.getOutputs()
	if err != nil {
		panic(err)
	}
	for {
		select {
		case data := <-l.outchan:
			l.runOutputs(outputs, data)
		case <-l.exitChan:
			goto exit
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
exit:
	l.logf("LOOKUP: closing")
}

func (l *Lolid) runOutputs(outputs []config.TypeOutputConfig, data []byte) error {
	if data == nil {
		return errors.New("data null")
	}

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
