package main

import (
	"flag"
	"fmt"
	app "github.com/domac/lolita/lolid"
	"github.com/domac/lolita/version"
	"github.com/judwhite/go-svc/svc"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

var (
	flagSet     = flag.NewFlagSet("lolita", flag.ExitOnError)
	port        = flagSet.String("port", "", "the port which server run") //端口
	showVersion = flagSet.Bool("version", false, "print version string")  //版本
)

//程序封装
type program struct {
	lolid *app.Lolid
}

func (p *program) Init(env svc.Environment) error {
	if env.IsWindowsService() {
		//切换工作目录
		dir := filepath.Dir(os.Args[0])
		return os.Chdir(dir)
	}
	return nil
}

//程序启动
func (p *program) Start() error {
	flagSet.Parse(os.Args[1:])
	if *showVersion {
		fmt.Println(version.String("lolita"))
		os.Exit(0)
	}
	return nil
}

//程序停止
func (p *program) Stop() error {
	if p.lolid != nil {
		p.lolid.Exit()
	}
	return nil
}

//引导程序
func main() {
	prg := &program{}
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}
