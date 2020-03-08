// +build !windows

package log

import (
	"os"
	"os/signal"
	"syscall"
)

/*HandleSignalChangeLogLevel 实现了根据信号量修改日志等级的方法
使用方法 go HandleSignalChangeLogLevel()
监听2个信号，kill -10 pid 修改日志等级为 all
kill -12 pid 修改日志等级为 error
*/
func HandleSignalChangeLogLevel() {
	chSignalUser1 := make(chan os.Signal)
	chSignalUser2 := make(chan os.Signal)
	signal.Notify(chSignalUser1, syscall.SIGUSR1)
	signal.Notify(chSignalUser2, syscall.SIGUSR2)
	for {
		select {
		case <-chSignalUser1:
			SetLogLevelAll()
		case <-chSignalUser2:
			SetLogLevelError()
		}
	}
}
