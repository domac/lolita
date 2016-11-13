package amqp

import (
	"github.com/bitly/go-hostpool"
	"github.com/streadway/amqp"
)

const ModuleName = "amqp"

type AmqpOutputHandler struct {
	Exchange       string
	ExchangeType   string
	Retries        int
	ReconnectDelay int
	hostPool       hostpool.HostPool
	amqpClients    map[string]amqpClient
}

type amqpClient struct {
	client    *amqp.Channel
	reconnect chan hostpool.HostPoolResponse
}

type amqpConn struct {
	Channel    *amqp.Channel
	Connection *amqp.Connection
}

func InitHandler(opt map[string]interface{}) *AmqpOutputHandler {
	return &AmqpOutputHandler{
		amqpClients: map[string]amqpClient{},
	}
}

func (handler *AmqpOutputHandler) Event(packets [][]byte) error {
	return nil
}
