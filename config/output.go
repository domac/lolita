package config

import (
	"errors"
	"fmt"
	"github.com/domac/lolita/output/amqp"
	"github.com/domac/lolita/output/stdout"
)

var (
	ERROR_NO_OUTPUT_CONFIG = "no output config found in conf file"
)

//初始化函数,用于注册输出器
func Init() {

	RegistOutputHandler(stdout.ModuleName, func(opt map[string]interface{}) OutputHandler {
		return stdout.InitHandler(opt)
	})

	RegistOutputHandler(amqp.ModuleName, func(opt map[string]interface{}) OutputHandler {
		return amqp.InitHandler(opt)
	})
}

//输出配置结构
type Output struct {
	Opts OutPutConfig
}

func NewOutput() *Output {
	outputs, err := GetOutputs()
	if err != nil {
		panic(err)
	}
	return &Output{
		Opts: outputs,
	}
}

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

//获取输出配置
func GetOutputs() (OutPutConfig, error) {
	fmt.Println("get output config...")
	if configRaw == nil {
		return nil, errors.New(ERROR_NO_OUTPUT_CONFIG)
	}

	if configRaw["outputs"] == nil {
		return nil, errors.New(ERROR_NO_OUTPUT_CONFIG)
	}

	configOutputs := configRaw["outputs"].([]map[string]interface{})
	if configOutputs == nil {
		return nil, errors.New(ERROR_NO_OUTPUT_CONFIG)
	}

	outputs := make(OutPutConfig, 0, len(configOutputs))

	for _, outMap := range configOutputs {
		handlerName := outMap["type"].(string)
		if _, ok := mapOutputHandler[handlerName]; !ok {
			continue
		}
		//只加入已经提交注册的输出器
		getHandler := mapOutputHandler[handlerName]
		if getHandler == nil {
			continue
		}
		outputs = append(outputs, outMap)
	}

	return outputs, nil
}

//执行输出
func (o *Output) Pop(packets [][]byte) error {
	//获取输出器
	outputs := o.Opts
	if len(outputs) == 0 {
		panic("no output available, please check the config file")
	}
	if packets == nil {
		return errors.New("the packets received was null !")
	}
	for _, outMap := range outputs {
		handlerName := outMap["type"].(string)
		getHandler := mapOutputHandler[handlerName]
		handler := getHandler(outMap)
		//异步触发输出工作
		go handler.Event(packets)
	}
	return nil
}
