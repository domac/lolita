# lolita

## 如何使用

### 1. 获取代码

```
$ go get -u github.com/domac/lolita
```


### 2. 生成部署文件

在lolita下,执行构建命令

```
$ make build
```

执行成功后,会在lolita同级目录下生成builds文件夹, 文件夹里面有两个文件:

- lolita : 执行的二进制文件
- lolita.conf : 执行所依赖的配置文件


### 3. 执行lolita

```
$ ./lolita -conf=/your/config/file/dir/lolita.conf
```
如果后台没报错,代表lolita执行成功,它会开始执行采集调度的服务

> 注意: 因为lolita调度需要依赖 `ectd` 和 `rabbitmq` 所以,lolita.conf中把环境的信息配置好,否则,lolita是
不会正常工作的.

### 4 Lolita 对外提供的 API

  > 可以通过浏览器操作下面的API

- 版本信息: 
  ```
  http://127.0.0.1:13360/version
  ```

- 堆信息
  ```
  http://127.0.0.1:13360/debug?cmd=heap
  ```

- 垃圾回收信息
  ```
  http://127.0.0.1:13360/debug?cmd=gc
  ```

- goroutinue信息
  ```
  http://127.0.0.1:13360/debug?cmd=go
  ```


## 开发事项

本项目使用godep把所需要的依赖包已经打包进去了, go 1.7 或以上的版本直接编译或开发就可以了

`go 1.7` 以下的,建议打开 `GO15VENDOREXPERIMENT`

```
export GO15VENDOREXPERIMENT=1
```

## 附: 命令参数

```
Usage of lolita:

  -config string
    	path to config file
  -etcd-endpoint string
    	ectd service discovery address (default "0.0.0.0:2379")
  -http-address string
    	<addr>:<port> to listen on for HTTP clients (default "0.0.0.0:13360")
  -max-write-bulk-size int
    	max writeBulk size (default 4096)
  -max-write-channel-size int
    	max writeChannel size (default 4096)
  -open-tasks
    	if opened, Lolid will execute tasks soon
  -rmq-address string
    	rabbitmq address
  -rmq-key string
    	rabbitmq queue key
  -send-type int
    	message send type: 0-output 1-dump
  -service-name string
    	the service name which ectd can find it (default "localhost")
  -tcp-address string
    	<addr>:<port> to listen on for HTTP clients (default "0.0.0.0:13361")
  -verbose
    	enable verbose logging
  -version
    	print version string
```

