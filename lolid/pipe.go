package lolid

import (
	"fmt"
	"github.com/domac/lolita/output/amqp"
	"math"
	"strings"
	"time"
)

//接驳到缓存库的管道
type Pipeline struct {
	ctx *Lolid
}

func NewPipeline(lolid *Lolid) *Pipeline {
	return &Pipeline{
		ctx: lolid,
	}
}

//数据转储
func (p *Pipeline) Dump() {

	p.ctx.logf("init Pipeline and dump messages \n")

	maxWirteBulkSize := p.ctx.opts.MaxWriteBulkSize
	//批量bulk
	packets := make([][]byte, 0, maxWirteBulkSize)

	rmq_address := p.ctx.opts.RmqAddress
	rmq_key := p.ctx.opts.RmqQueueKey

	rmq_address_endpoints := strings.Split(rmq_address, ",")
	store_handler := amqp.NewAmpqHandler(rmq_address_endpoints, rmq_key, "", "")

	err := store_handler.InitAmqpClients()
	if err != nil {
		fmt.Println("cache store server is not working")
		panic(err)
	}
	//关闭messageCollectStartedChan, 宣告输出器的初始化工作已经完成
	//其它工作组件可以往下走
	close(p.ctx.messageCollectStartedChan)

	for {
		select {
		case data := <-p.ctx.outchan:

			if nil != data {
				packets = append(packets, data)
			}

			chanlen := int(math.Min(float64(len(p.ctx.outchan)), float64(maxWirteBulkSize)))

			//如果channel的长度还有数据, 批量最多读取maxWirteBulkSize条数据,再合并写出
			//减少系统调用
			//减少网络传输, 提高资源利用率
			for i := 0; i < chanlen; i++ {
				p := <-p.ctx.outchan
				if nil != data {
					packets = append(packets, p)
				}
			}

			if len(packets) > 0 {
				//执行输出
				store_handler.WriteToMQ(packets)
				//回收包裹空间
				packets = packets[:0]
			}
		case <-p.ctx.exitChan:
			goto exit
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}
exit:
	p.ctx.logf("LOOKUP: closing")
}
