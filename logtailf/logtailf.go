package logtailf

import (
	"context"
	"fmt"
	"time"

	"github.com/hpcloud/tail"
)

// 实现日志文件的监控功能
func WatchLogFile(datapath string, ctx context.Context) {
	fmt.Println("begin goroutine watch log file ", datapath)
	tailFile, err := tail.TailFile(datapath, tail.Config{
		// 文件被移除或被打包，需要重新打开
		ReOpen: true,
		//实时跟踪
		Follow: true,
		// 如果程序异常， 保存上次读取的位置， 避免重复读取
		Location: &tail.SeekInfo{Offset: 0, Whence: 2},
		// 支持文件不存在
		MustExist: false,
		Poll:      true,
	})
	if err != nil {
		fmt.Println("tail file err:", err)
		return
	}

	for true {
		select {
		case msg, ok := <-tailFile.Lines:
			if !ok {
				fmt.Printf("tail file close reopen, filename: %s\n", tailFile.Filename)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			fmt.Println("msg:", msg.Text)
		case <-ctx.Done():
			fmt.Println("receive main goroutine exit msg")
			fmt.Println("watch log file ", datapath, " goroutine exited")
			return
		}
	}
}
