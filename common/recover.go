package common

import (
	"fmt"
	"runtime/debug"

	"gorm.io/gorm"

	"tianxian.com/tee-signer-core/log"
)

/*
Recover 通用协程 panic 的记录方法
请将该函数置于 主进程 或 协程的开始处

example：

	func main() {
		defer common.Recover()
		...
	}
*/
func Recover() {
	if r := recover(); r != nil {
		err := fmt.Errorf("%v\nstacktrace from panic: %s", r, string(debug.Stack()))
		log.Entry.Error(err)
		// 以防 log 失效
		fmt.Printf("\n------------- avoid log failure -------------\n%s\n---------------------------------------------\n",
			err)
	}
}