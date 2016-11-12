package config

import (
	"errors"
	"github.com/domac/lolita/output/stdout"
)

var configRaw map[string]interface{}

type OutPutConfig []map[string]interface{}

type OutputHandler interface {
	Event(packets [][]byte) error
}

var mapOutputHandler = make(map[string]func(opt map[string]interface{}) OutputHandler)

//输出器配置
func RegistOutputHandler(name string, handler func(opt map[string]interface{}) OutputHandler) {
	mapOutputHandler[name] = handler
}

func SetConfigInfo(config map[string]interface{}) {
	configRaw = config
}

//初始化函数
func Init() {
	RegistOutputHandler(stdout.ModuleName, func(opt map[string]interface{}) OutputHandler {
		return stdout.InitHandler(opt)
	})
}

//获取输出配置
func GetOutputs() (OutPutConfig, error) {
	if configRaw == nil {
		return nil, errors.New("no config")
	}
	configOutputs := configRaw["outputs"].([]map[string]interface{})
	if configOutputs == nil {
		return nil, errors.New("no outputs config")
	}
	return configOutputs, nil
}

//执行输出
func RunOutputs(packets [][]byte) error {
	//获取输出器
	outputs, err := GetOutputs()
	if err != nil {
		panic(err)
	}

	if packets == nil {
		return errors.New("data null")
	}
	for _, outMap := range outputs {
		handlerName := outMap["type"].(string)
		if _, ok := mapOutputHandler[handlerName]; !ok {
			continue
		}
		getHandler := mapOutputHandler[handlerName]
		if getHandler == nil {
			continue
		}
		handler := getHandler(outMap)
		handler.Event(packets)
	}
	return nil
}
