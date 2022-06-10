package main

import (
	"flag"
	"github.com/qiuqiu1999/crontab/master"
	"log"
	"runtime"
	"time"
)

var (
	confFile string // 配置文件路径
)

// 解析命令行参数
func initArgs() {
	flag.StringVar(&confFile, "config", "./config/master.json", "指定master.json")
	flag.Parse()
}

// 初始化线程数量
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func init() {
	initArgs()
	initEnv()

	// 加载配置
	if err := master.InitConfig(confFile); err != nil {
		log.Fatalf("master.SetupConfig err: %v", err)
	}
}

func main() {
	var err error

	// 初始化服务发现模块
	if err = master.InitWorkerMgr(); err != nil {
		log.Fatalf("master.InitWorkerMgr err: %v", err)
	}

	// 日志管理器
	if err = master.InitLogMgr(); err != nil {
		log.Fatalf("master.InitLogMgr err: %v", err)

	}

	//  任务管理器
	if err = master.InitJobMgr(); err != nil {
		log.Fatalf("master.InitJobMgr err: %v", err)

	}

	// 启动Api HTTP服务
	if err = master.InitApiServer(); err != nil {
		log.Fatalf("master.InitApiServer err: %v", err)

	}

	// 正常退出
	for {
		time.Sleep(1 * time.Second)
	}
}
