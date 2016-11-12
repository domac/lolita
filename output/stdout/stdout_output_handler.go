package stdout

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
	println(temp)
	println("===============================")
	return nil
}
