package lolid

import (
	"sync"
)

//etcd实例结构
type EtcdInstance struct {
	messageCounr  uint64
	ip            string
	memoryMsgChan chan []byte
	sync.RWMutex
	ctx *context

	paused int32
}
