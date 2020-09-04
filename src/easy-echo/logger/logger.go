package logger

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	constlogbuffersize = 256 * 1024
)

var (
	logFileDir string
	logFileType string
	global     *logger
)

type entry struct {
	sync.RWMutex
	buff *bytes.Buffer
}

//Logger ...
type logger struct {
	sync.RWMutex
	writer map[string]*entry
}

//GetLogger ...
func getLogger(category string) *entry {
	var sublogger *entry
	global.RLock()
	sublogger = global.writer[category]
	global.RUnlock()

	if sublogger == nil {
		global.Lock()
		sublogger = global.writer[category]
		if sublogger == nil {
			sublogger = new(entry)
			sublogger.buff = bytes.NewBuffer(make([]byte, constlogbuffersize))
			sublogger.buff.Reset()
			global.writer[category] = sublogger
		}
		global.Unlock()
	}
	return sublogger
}

//Init ...
func Init(logPath, logType string) {
	logFileDir = logPath
	logFileType = logType
	if logFileDir == "" {
		logFileDir = "./log"
	}
	err := os.MkdirAll(logPath, 0644)
	if err != nil {
		panic(fmt.Sprintf("mkdir %s fail:", logPath))
	}
	global = new(logger)
	global.writer = make(map[string]*entry)
	go flushlogtimely()
}

//Error ...
func Error(err string) {
	WriteLog("error", err)
}

//Info ...
func Info(msg string) {
	WriteLog("info", msg)
}

//WriteLog ...
func WriteLog(category string, msgs ...string) {

	length := len(msgs)
	if length == 0 {
		return
	}

	if logFileType == "csv" {
		for i := 0; i < length; i++ {
			modifiedStr := "'" + strings.ReplaceAll(msgs[i], "\"", "'") + "'"
			msgs[i] = modifiedStr
		}
		err := writeCSV(category, msgs)
		if err != nil {
			log.Println("writeCSV failure, err:", err.Error())
		}
	}else {
		sublogger := getLogger(category)
		for i := 0; i < length; i++ {
			modifiedstr := "\"" + strings.Replace(msgs[i], "\"", "\"\"", 0) + "\""
			if i < length-1 {
				sublogger.buff.WriteString(modifiedstr + ",")
			} else {
				sublogger.buff.WriteString(modifiedstr)
			}
		}
		sublogger.buff.WriteString("\n")

		if sublogger.buff.Len() > constlogbuffersize {
			flush(category)
		}
	}
}

//flush flush string
func flush(category string) {
	sublogger := getLogger(category)

	sublogger.Lock()

	filename := filepath.Join(logFileDir, time.Now().Format("20060102")+"."+category+".auditlog")
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(fmt.Sprintf("open logfile %s fail: %s", filename, err.Error()))
	}
	defer file.Close()
	file.Write(sublogger.buff.Bytes())
	sublogger.buff.Reset()

	defer sublogger.Unlock()
}

//flushAll flush all
func flushAll() {
	for category := range global.writer {
		flush(category)
	}
}

func flushlogtimely() {
	for {
		flushAll()
		select {
		case <-time.After(30 * time.Second):
		}
	}
}
