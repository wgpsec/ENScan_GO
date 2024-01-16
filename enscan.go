package main

/*
 * BY WgpSec Keac~
 * 如果你也有兴趣一起写点好玩的东西，欢迎加入我们团队~
 */
import (
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/db"
	"github.com/wgpsec/ENScan/runner"
)

import (
	"github.com/adjust/rmq/v4"
	"github.com/wgpsec/ENScan/api"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	var enOptions common.ENOptions
	common.Flag(&enOptions)
	common.Parse(&enOptions)
	//如果不是API模式就直接运行了
	if !enOptions.IsApiMode && !enOptions.IsWebMode {
		runner.RunEnumeration(&enOptions)
	} else if enOptions.IsWebMode {
		//web模式，不加入队列，适合轻量级环境
		//db.ConnDB = db.InitWebDb(&enOptions)
		go api.RunWeb(&enOptions)
		//监听系统信号判断退出操作
		var quitSig = make(chan os.Signal, 1)
		signal.Notify(quitSig, os.Interrupt, os.Kill)
		select {
		case <-quitSig:
			log.Fatal("exit.by.signal")
		}
	} else {
		db.InitRedis(&enOptions)
		db.InitMongo(&enOptions)
		db.InitQueue(&enOptions)
		if enOptions.ClientMode == "" {
			go api.RunApiWeb(&enOptions)
		}
		go runner.Worker(&enOptions)
		if enOptions.ClientMode != "" {
			//定时清理队列信息
			cleaner := rmq.NewCleaner(db.RmqC)
			for range time.Tick(time.Hour) {
				returned, err := cleaner.Clean()
				if err != nil {
					gologger.Errorf("failed to clean: %s\n", err)
					continue
				}
				gologger.Infof("cleaned %d\n", returned)
			}
		}
		//监听系统信号判断退出操作
		var quitSig = make(chan os.Signal, 1)
		signal.Notify(quitSig, os.Interrupt, os.Kill)
		select {
		case <-quitSig:
			log.Fatal("exit.by.signal")
		}
	}

}
