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
	Logger      Logger
}

func NewOptions() *Options {
	return &Options{
		HTTPAddress: "0.0.0.0:13360",
		TCPAddress:  "0.0.0.0:13361",
		Logger:      log.New(os.Stderr, "[LOLID] ", log.Ldate|log.Ltime|log.Lmicroseconds),
	}
}
