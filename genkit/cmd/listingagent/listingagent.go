package main

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent"
	"math/rand"
	"time"
	// Auto set max process
	_ "go.uber.org/automaxprocs"
)

func main() {
	// 全局初始化一次 Seed, 保障后续使用 rand 进行随机数生成不会有重复的问题；
	rand.Seed(time.Now().UTC().UnixNano())

	//basename 最好保持跟 apps/apiserver 这个文件夹名一致；
	//因为这个 basename 会作为 cmd 命令，而 MK 编译处理的 exec 文件是按文件名来的。
	apiserver.NewApp("listingagent").Run()
}
