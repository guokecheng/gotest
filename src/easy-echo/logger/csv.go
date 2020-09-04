package logger

import (
	"encoding/csv"
	"fmt"
	"github.com/golang/glog"
	"io"
	"os"
	"time"
)

const (
	timeFormat    = "2006010215"
	logTimeFormat = "2006-01-02 15:04:05.000"
)

func writeCSV(category string, msgs []string) (err error) {

	timeStr := time.Now().Format(timeFormat)

	//构造文件名
	fileName := fmt.Sprintf("%s.%s_%s.csv", timeStr[:8], category, timeStr[8:])

	logTime := time.Now().Format(logTimeFormat)
	data := []string{
		logTime, category, fmt.Sprintf("%s", msgs),
	}

	err = writeCSVFile(fileName, data)
	if err != nil {
		return  err
	}

	return  nil
}

func writeCSVFile(fileName string, data []string) error {

	nfs, err := os.OpenFile(fmt.Sprintf("%s/%s", logFileDir, fileName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		glog.Errorf("writeCSVFile err is %s", err)
	}

	defer nfs.Close()
	//防止中文乱码
	_, _ = nfs.WriteString("\xEF\xBB\xBF")
	_, _ = nfs.Seek(0, io.SeekEnd)

	w := csv.NewWriter(nfs) //创建一个新的写入文件流
	w.Comma = ','
	w.UseCRLF = false

	_ = w.Write(data) //写入数据
	w.Flush()
	return nil
}
