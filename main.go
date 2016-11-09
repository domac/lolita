package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	app "github.com/domac/lolita/lolid"
	"github.com/domac/lolita/version"
	"github.com/judwhite/go-svc/svc"
	"github.com/mreiferson/go-options"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

var (
	flagSet     = flag.NewFlagSet("lolita", flag.ExitOnError)
	showVersion = flagSet.Bool("version", false, "print version string") //版本
	config      = flagSet.String("config", "", "path to config file")
	verbose     = flagSet.Bool("verbose", false, "enable verbose logging")                                      //配置文件
	httpAddress = flagSet.String("http-address", "0.0.0.0:6060", "<addr>:<port> to listen on for HTTP clients") //http定义地址
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

	var cfg map[string]interface{}
	if *config != "" {
		_, err := toml.DecodeFile(*config, &cfg)
		if err != nil {
			log.Fatalf("ERROR: failed to load config file %s - %s", *config, err.Error())
		}
	}

	opts := app.NewOptions()
	options.Resolve(opts, flagSet, cfg)

	daemon := app.New(opts)
	daemon.Main()
	p.lolid = daemon
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
