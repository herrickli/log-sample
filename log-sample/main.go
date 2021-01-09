package main

import (
	"context"
	"fmt"
	"sync"

	"log-sample/logconfig"
	"log-sample/logtailf"

	"github.com/spf13/viper"
)

var mainOnce sync.Once
var configMgr map[string]*logconfig.ConfigData

// ConstructMgr 控制著协成资源析构，
// 并且通过ConstructMgr全局函数构造configMgr这样的map记录最新的配置信息。
func ConstructMgr(configPaths interface{}, keyChan chan string) {
	configDatas := configPaths.(map[string]interface{})
	for conkey, confval := range configDatas {
		fmt.Println("conkey:", conkey)
		fmt.Println("conval:", confval)
		configData := new(logconfig.ConfigData)
		configData.ConfigKey = conkey
		configData.ConfigValue = confval.(string)
		ctx, cancel := context.WithCancel(context.Background())
		configData.ConfigCancel = cancel
		configMgr[conkey] = configData
		// 添加协程启动逻辑，将协程的ctx保存在map中，这样住协程可以根据热更新
		// 启动和关闭这个协程
		go logtailf.WatchLogFile(conkey, configData.ConfigValue, ctx, keyChan)
	}
}

func main() {
	v := viper.New()
	configPaths, confres := logconfig.ReadConfig(v)
	if configPaths == nil || !confres {
		fmt.Println("read config failed")
		return
	}
	KEYCHANSIZE := 1
	keyChan := make(chan string, KEYCHANSIZE)
	configMgr = make(map[string]*logconfig.ConfigData)
	ConstructMgr(configPaths, keyChan)
	ctx, cancel := context.WithCancel(context.Background())
	pathChan := make(chan interface{})
	go logconfig.WatchConfig(ctx, v, pathChan)
	defer func() { //析构函数
		mainOnce.Do(func() {
			if err := recover(); err != nil {
				fmt.Println("main gorroutine panic ", err)
			}
			cancel()
			for _, oldval := range configMgr {
				oldval.ConfigCancel()
			}
			configMgr = nil
		})
	}()

	for {
		select {
		case pathData, ok := <-pathChan:
			if !ok {
				return
			}
			fmt.Println("main goroutine reveive pathData")
			fmt.Println(pathData)
			pathDataNew := pathData.(map[string]interface{}) // golang Type Assertion, 类型断言

			for oldkey, oldval := range configMgr {
				_, ok := pathDataNew[oldkey]
				if ok {
					continue
				}
				oldval.ConfigCancel()
				delete(configMgr, oldkey)
			}

			for conkey, conval := range pathDataNew {
				oldval, ok := configMgr[conkey]
				if !ok {
					configData := new(logconfig.ConfigData)
					configData.ConfigKey = conkey
					configData.ConfigValue = conval.(string)
					ctx, cancel := context.WithCancel(context.Background())
					configData.ConfigCancel = cancel
					configMgr[conkey] = configData
					fmt.Println(conval.(string))
					go logtailf.WatchLogFile(conkey, configData.ConfigValue, ctx, keyChan)
					continue
				}

				if oldval.ConfigValue != conval.(string) {
					oldval.ConfigValue = conval.(string)
					oldval.ConfigCancel()
					ctx, cancel := context.WithCancel(context.Background())
					oldval.ConfigCancel = cancel
					go logtailf.WatchLogFile(conkey, conval.(string), ctx, keyChan)
					continue
				}
			}

			for mgrkey, mgrval := range configMgr {
				fmt.Println(mgrkey)
				fmt.Println(mgrval)
			}
		case keystr := <-keyChan:
			val, ok := configMgr[keystr]
			if !ok {
				print("get keystr filed")
				continue
			}
			fmt.Println("recover goroutine watch ", keystr)
			var ctxcover context.Context
			ctxcover, val.ConfigCancel = context.WithCancel(context.Background())
			go logtailf.WatchLogFile(keystr, val.ConfigValue, ctxcover, keyChan)
		}
	}
}
