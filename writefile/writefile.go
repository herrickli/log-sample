package main

import (
	"bufio"
	"fmt"
	"log-sample/logconfig"
	"os"
	"path"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// 模拟不同系统向指定日志写文件，
// 用来测试日志写入和备份时，采集系统还是否健壮
func writeLog(datapath string, wg *sync.WaitGroup) {
	filew, err := os.OpenFile(datapath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("open file error ", err.Error())
		return
	}
	defer func() {
		wg.Done()
	}()
	w := bufio.NewWriter(filew)
	for i := 0; i < 20; i++ {
		timeStr := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintln(w, "Hello current time is "+timeStr)
		time.Sleep(time.Millisecond * 100)
		w.Flush()
	}
	logBak := time.Now().Format("20060102150405") + ".txt"
	logBak = path.Join(path.Dir(datapath), logBak)
	filew.Close()
	err = os.Rename(datapath, logBak)
	if err != nil {
		fmt.Println("Rename error ", err.Error())
		return
	}
}

func main() {
	v := viper.New()
	configPaths, confres := logconfig.ReadConfig(v)
	if !confres {
		fmt.Println("config read failed")
		return
	}
	wg := &sync.WaitGroup{}

	for _, confval := range configPaths.(map[string]interface{}) {
		wg.Add(1)
		go writeLog(confval.(string), wg)
	}
	wg.Wait()
}
