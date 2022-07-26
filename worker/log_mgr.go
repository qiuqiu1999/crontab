package worker

import (
	"fmt"
	"github.com/hpcloud/tail"
	"log"
	"time"
)

type LogMgr struct {
	fileList map[string]string //监控的文件
}

// 监控文件
func (l *LogMgr) watcher() {
	fmt.Println("Watcher start")
	for _, v := range l.fileList {
		go func(filename string) {
			var (
				tails *tail.Tail
				err   error
			)

			// todo lock =》return

			// 监控
			if tails, err = tail.TailFile(filename, tail.Config{
				ReOpen:    true,                                 // 重新打开
				Follow:    true,                                 // 是否跟随
				Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件的哪个地方开始读
				MustExist: false,                                // 文件不存在不报错
				Poll:      true,                                 // 监听新行，使用tail -f，这个参数非常重要
			}); err != nil {
				log.Println("tail file failed, err:", err)
				return
			} else {
				// todo lock
			}

			for {
				line, ok := <-tails.Lines
				if !ok {
					log.Printf("tail file close reopen, filename:%s\n", tails.Filename)
					time.Sleep(time.Second)
					// todo unlock
					return
				}
				// todo line.Text => clickhouseChan
				log.Println(line.Text)
			}
		}(v)
	}
}
