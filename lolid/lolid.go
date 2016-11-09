package lolid

import (
	"fmt"
	"github.com/domac/lolita/util"
	"sync"
)

type Lolid struct {
	sync.RWMutex
	waitGroup util.WaitGroupWrapper
}

func New() *Lolid {
	return &Lolid{}
}

func (l *Lolid) Main() {
	fmt.Println("lolita love you")
}

//后台程序退出
func (l *Lolid) Exit() {
	l.waitGroup.Wait()
}
