package tcp

import (
	"net"
)

//连接结构
type Conn struct {
	conn *net.TCPConn
}

//创建连接
func NewConn() *Conn {

	return &Conn{}
}
