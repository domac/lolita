package amqp

import (
	"github.com/bitly/go-hostpool"
	"github.com/streadway/amqp"
)

const ModuleName = "amqp"

type AmqpOutputHandler struct {
	URLs           []string
	Exchange       string
	ExchangeType   string
	Retries        int
	ReconnectDelay int
	hostPool       hostpool.HostPool
	amqpClients    map[string]amqpClient
	isCheck        bool
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

	handler := &AmqpOutputHandler{
		amqpClients: map[string]amqpClient{},
	}
	if _, ok := opt["urls"]; ok {
		p_urls := opt["urls"].([]interface{})
		urls := []string{}
		for i := 0; i < len(p_urls); i++ {
			urls = append(urls, p_urls[i].(string))
		}
		handler.URLs = urls
	}

	if _, ok := opt["exchange"]; ok {
		handler.Exchange = opt["exchange"].(string)
	}

	if _, ok := opt["exchange_type"]; ok {
		handler.ExchangeType = opt["exchange_type"].(string)
	}

	if _, ok := opt["retries"]; ok {
		retries := opt["retries"].(int64)
		handler.Retries = int(retries)
	}
	handler.isCheck = true
	return handler
}

func (self *AmqpOutputHandler) initAmqpClients() error {

	return nil
}

func (self *AmqpOutputHandler) Event(packets [][]byte) error {
	return nil
}

func (self *AmqpOutputHandler) getConnection(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	return conn, err
}

func (self *AmqpOutputHandler) Check() bool {
	return self.isCheck
}
