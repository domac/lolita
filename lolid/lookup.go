package lolid

import "time"
import "fmt"

//主动发现需要去做的服务
func (l *Lolid) lookupTasks() {
	ticker := time.Tick(10 * time.Second)

	for {
		select {
		case <-ticker:
			fmt.Println("loop now")
		case <-l.exitChan:
			goto exit
		}
	}
exit:
	l.logf("LOOKUP: closing")
}
