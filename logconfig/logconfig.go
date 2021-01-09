package logconfig

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var onceLogConf sync.Once

// ConfigData is a struct
type ConfigData struct {
	ConfigKey    string
	ConfigValue  string
	ConfigCancel context.CancelFunc
}

// ReadConfig load the config file
func ReadConfig(v *viper.Viper) (interface{}, bool) {
	// 设置读取的配置文件
	v.SetConfigName("config")
	// 添加读取的配置文件路径
	_, filename, _, _ := runtime.Caller(0)
	fmt.Println("filename:", filename)
	fmt.Println(path.Dir(filename))
	// 如果不设置AddConfigPath去指定路径， 它会在程序执行的目录去寻找config.yaml
	v.AddConfigPath(path.Dir(filename))
	// 设置配置文件类型
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		fmt.Print("Read config file failed, err is :", err.Error())
		return nil, false
	}

	configPaths := v.Get("configpath")
	if configPaths == nil {
		return nil, false
	}
	return configPaths, true
}

// WatchConfig watchs the config file
func WatchConfig(ctx context.Context, v *viper.Viper, pathChan chan interface{}) {
	defer func() {
		onceLogConf.Do(func() {
			fmt.Println("watch config goroutine exit")
			if err := recover(); err != nil {
				fmt.Println("watch config goroutine panic", err)
			}
			close(pathChan)
		})
	}()

	// 设置监听回调函数,当配置文件变更时会调用这个函数
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("config is change: %s \n", e.String())
		configPaths := v.Get("configpath")
		if configPaths == nil {
			return
		}
		pathChan <- configPaths
	})

	// 开始监听
	v.WatchConfig()
	// 信道不会主动关闭， 可以主动调用cancel关闭
	<-ctx.Done()
}
