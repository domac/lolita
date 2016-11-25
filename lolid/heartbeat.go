package lolid

import (
	"fmt"
	"github.com/domac/lolita/version"
	"time"
)

func TouchHeart(ip string) string {
	return fmt.Sprintf("%s-v%s-%d", ip, version.Binary, time.Now().Unix())
}
