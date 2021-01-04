package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"
)

func writeLog(dataPath string) {
	filew, err := os.OpenFile(dataPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("open file error: ", err.Error())
		return
	}

	w := bufio.NewWriter(filew)
	for i := 0; i < 20; i++ {
		timeStr := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintln(w, "Hello current time is "+timeStr)
		time.Sleep(time.Millisecond * 100)
		w.Flush()
	}
	logBak := time.Now().Format("20200102150405") + ".txt"
	logBak = path.Join(path.Dir(dataPath), logBak)
	filew.Close()
	err = os.Rename(dataPath, logBak)
	if err != nil {
		fmt.Println("Rename error, ", err.Error())
		return
	}
}

func main() {
	logrelative := `../logdir/log.txt`
	_, filename, _, _ := runtime.Caller(0) //filename是当前运行的文件名，包括路径
	fmt.Println("filaneme:", filename)
	datapath := path.Join(path.Dir(filename), logrelative)
	fmt.Println("datapath:", datapath)

	for i := 0; i < 3; i++ {
		writeLog(datapath)
	}
}
