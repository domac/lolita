package amqp

import (
	"errors"
	"fmt"
	"github.com/bitly/go-hostpool"
	"github.com/streadway/amqp"
	"log"
	"time"
)

const ModuleName = "amqp"

type AmqpOutputHandler struct {
	URLs               []string
	Exchange           string
	ExchangeType       string
	ExchangeDurable    bool
	ExchangeAutoDelete bool
	Retries            int
	ReconnectDelay     int
	hostPool           hostpool.HostPool
	amqpClients        map[string]amqpClient
	isCheck            bool
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
	handler.ExchangeDurable = false
	handler.ExchangeAutoDelete = true

	if err := handler.InitAmqpClients(); err != nil {
		fmt.Println(err.Error())
		handler.isCheck = false
	}

	return handler
}

func NewAmpqHandler(urls []string, exchange string, exchange_type string) *AmqpOutputHandler {
	handler := &AmqpOutputHandler{
		URLs:               urls,
		Exchange:           exchange,
		ExchangeType:       exchange_type,
		ExchangeDurable:    false,
		ExchangeAutoDelete: true,
		amqpClients:        map[string]amqpClient{},
	}
	return handler
}

func (self *AmqpOutputHandler) InitAmqpClients() error {
	var hosts []string
	for _, url := range self.URLs {
		if conn, err := self.getConnection(url); err == nil {
			if ch, err := conn.Channel(); err == nil {
				err := ch.ExchangeDeclare(
					self.Exchange,
					self.ExchangeType,
					self.ExchangeDurable,
					self.ExchangeAutoDelete,
					false,
					false,
					nil,
				)
				if err != nil {
					return err
				}
				self.amqpClients[url] = amqpClient{
					client:    ch,
					reconnect: make(chan hostpool.HostPoolResponse, 1),
				}
				//重连处理
				go self.reconnect(url)
				hosts = append(hosts, url)
			}
		}

	}
	if len(hosts) == 0 {
		return errors.New("FAIL TO CONNECT AMQP SERVERS")
	}

	return nil
}

func (self *AmqpOutputHandler) Event(packets [][]byte) error {
	fmt.Printf("amqp is sending %d envent \n", len(packets))
	return nil
}

func (self *AmqpOutputHandler) Write(packets [][]byte) error {
	fmt.Printf("amqp is writing %d msg \n", len(packets))
	return nil
}

func (self *AmqpOutputHandler) getConnection(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	return conn, err
}

//重连机制
func (self *AmqpOutputHandler) reconnect(url string) {
	for {
		select {
		case poolResponse := <-self.amqpClients[url].reconnect:
			for {
				time.Sleep(time.Duration(self.ReconnectDelay) * time.Second)

				if conn, err := self.getConnection(poolResponse.Host()); err == nil {
					if ch, err := conn.Channel(); err == nil {
						if err := ch.ExchangeDeclare(
							self.Exchange,
							self.ExchangeType,
							self.ExchangeDurable,
							self.ExchangeAutoDelete,
							false,
							false,
							nil,
						); err == nil {
							self.amqpClients[poolResponse.Host()] = amqpClient{
								client:    ch,
								reconnect: make(chan hostpool.HostPoolResponse, 1),
							}
							poolResponse.Mark(nil)
							break
						}
					}
				}
				log.Println("Failed to reconnect to ", url, ". Waiting ", self.ReconnectDelay, " seconds...")
			}
		}
	}
}

func (self *AmqpOutputHandler) Check() bool {
	return self.isCheck
}
