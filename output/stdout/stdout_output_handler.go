package stdout

const ModuleName = "stdout"

type StdoutOutputHandler struct {
}

func InitHandler(opt map[string]interface{}) *StdoutOutputHandler {
	return &StdoutOutputHandler{}
}

func (handler *StdoutOutputHandler) Event(packets [][]byte) error {
	println("----------stdout")
	return nil
}
