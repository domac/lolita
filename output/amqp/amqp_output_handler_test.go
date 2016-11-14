package amqp

import (
	"testing"
)

func TestAMQPConnection(t *testing.T) {
	urls := []string{"amqp://admin:admin@127.0.0.1:5672"}
	exchange := ""
	exchange_type := ""
	handler := NewAmpqHandler(urls, "", exchange, exchange_type)
	err := handler.InitAmqpClients()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("connect success")
	}

	packets := [][]byte{}
	packets = append(packets, []byte{90})
	packets = append(packets, []byte{90})
	packets = append(packets, []byte{90})
	err = handler.WriteToMQ(packets)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("write success")
	}
}
