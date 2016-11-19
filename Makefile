.PHONY: all clean
# 被编译的文件
BUILDFILE=main.go
# 编译后的静态链接库文件名称
TARGETNAME=lolita
# GOOS为目标主机系统 
# mac os : "darwin"
# linux  : "linux"
# windows: "windows"
GOOS=darwin 
# GOARCH为目标主机CPU架构, 默认为amd64 
GOARCH=amd64

VER=$(shell sh ./version/ver.sh)

BUILDPATH=$(TARGETNAME)-$(VER)

all: format test build clean

test:
	go test -v . 

format:
	gofmt -w .

build:
	mkdir -p builds/$(BUILDPATH)
	cp config/cfg_sample.conf builds/$(BUILDPATH)/$(TARGETNAME).conf
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o builds/$(BUILDPATH)/$(TARGETNAME) $(BUILDFILE) 
	tar -zcvf ./builds/$(BUILDPATH).tar.gz ./builds/$(BUILDPATH)

clean:
	go clean -i
