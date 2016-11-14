package lolid

import (
	"log"
	"os"
)

//配置选项
type Options struct {
	Verbose     bool   `flag:"verbose"`
	HTTPAddress string `flag:"http-address"`
	TCPAddress  string `flag:"tcp-address"`
	OpenTasks   bool   `flag:"open-tasks"`

	MaxWriteChannelSize int `flag:"max-write-channel-size"`
	MaxWriteBulkSize    int `flag:"max-write-bulk-size"`

	SendType     int    `flag:"send-type"`
	EtcdEndpoint string `flag:"etcd-endpoint"`
	ServiceName  string `flag:"service-name"`

	RmqAddress  string `flag:"rmq-address"`
	RmqQueueKey string `flag:"rmq-key"`

	Logger Logger
}

func NewOptions() *Options {
	return &Options{
		HTTPAddress:         "0.0.0.0:13360",
		TCPAddress:          "0.0.0.0:13361",
		EtcdEndpoint:        "0.0.0.0:2379",
		ServiceName:         "localhost",
		MaxWriteChannelSize: 4096,
		MaxWriteBulkSize:    200,
		SendType:            0,
		Logger:              log.New(os.Stderr, "[LOLID] ", log.Ldate|log.Ltime|log.Lmicroseconds),
	}
}
