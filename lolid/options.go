package lolid

//配置选项
type Options struct {
	Verbose bool `flag:"verbose"`
}

func NewOptions() *Options {
	return &Options{}
}
