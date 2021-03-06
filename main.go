package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/BurntSushi/toml"
	appconfig "github.com/domac/lolita/config"
	app "github.com/domac/lolita/lolid"
	"github.com/domac/lolita/version"
	"github.com/judwhite/go-svc/svc"
	"github.com/mreiferson/go-options"
)

var (
	flagSet     = flag.NewFlagSet("lolita", flag.ExitOnError)
	showVersion = flagSet.Bool("version", false, "print version string") //版本
	config      = flagSet.String("config", "", "path to config file")
	verbose     = flagSet.Bool("verbose", false, "enable verbose logging")                                       //配置文件
	httpAddress = flagSet.String("http-address", "0.0.0.0:13360", "<addr>:<port> to listen on for HTTP clients") //http定义地址
	tcpAddress  = flagSet.String("tcp-address", "0.0.0.0:13361", "<addr>:<port> to listen on for HTTP clients")  //tcp定义地址
	openTasks   = flagSet.Bool("open-tasks", false, "if opened, lolid will execute tasks soon")

	MaxWriteChannelSize = flagSet.Int("max-write-channel-size", 4096, "max writeChannel size")
	MaxWriteBulkSize    = flagSet.Int("max-write-bulk-size", 4096, "max writeBulk size")

	sendType     = flagSet.Int("send-type", 0, "message send type: 0-output 1-dump")
	etcdEndpoint = flagSet.String("etcd-endpoint", "0.0.0.0:2379", "ectd service discovery address")
	AgentId      = flagSet.String("agent-id", "localhost", "the service name which ectd can find it")
	AgentGroup   = flagSet.String("agent-group", "devops", "the service group which agent work on")

	rmq_address = flagSet.String("rmq-address", "", "rabbitmq address")
	rmq_key     = flagSet.String("rmq-key", "", "rabbitmq queue key")
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

	appconfig.SetConfigInfo(cfg)

	//后台进程创建
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
