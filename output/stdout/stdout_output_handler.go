package stdout

import (
	"fmt"
)

const ModuleName = "stdout"

type StdoutOutputHandler struct {
}

func InitHandler(opt map[string]interface{}) *StdoutOutputHandler {
	return &StdoutOutputHandler{}
}

func (handler *StdoutOutputHandler) Event(packets [][]byte) error {
	temp := ""
	for i := 0; i < len(packets); i++ {
		temp += string(packets[i]) + ","
	}
	fmt.Printf("%d >>>> : %s \n\n", len(packets), temp)
	return nil
}

func (handler *StdoutOutputHandler) Check() bool {
	return true
}
