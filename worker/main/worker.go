package main

import (
	"crontab/worker"
	"flag"
	"log"
	"runtime"
	"time"
)

var (
	confFile string // 配置文件路径
)

// 解析命令行参数
func initArgs() {
	flag.StringVar(&confFile, "config", "./config/worker.json", "worker.json")
	flag.Parse()
}

// 初始化线程数量
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var err error
	initArgs()
	initEnv()

	// 加载配置
	if err = worker.InitConfig(confFile); err != nil {
		log.Fatalf("worker.InitConfig err: %v", err)
	}

	// 服务注册
	if err = worker.InitRegister(); err != nil {
		log.Fatalf("worker.InitRegister err: %v", err)
	}

	// 启动日志协程
	if err = worker.InitLogSink(); err != nil {
		log.Fatalf("worker.InitLogSink err: %v", err)
	}

	// 启动调度器
	worker.InitScheduler()

	// 初始化任务管理器
	if err = worker.InitJobMgr(); err != nil {
		log.Fatalf("worker.InitJobMgr err: %v", err)
	}

	for {
		time.Sleep(1 * time.Second)
	}
}
