package lolid

//日志器
type Logger interface {
	Output(maxdepth int, s string) error
}
