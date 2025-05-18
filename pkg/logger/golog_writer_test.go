package logger

import (
	"fmt"
	"testing"
	"time"
)

func Test_StderrWrite(t *testing.T) {
	w := NewWriter(&WriterCfg{OpenMsgFormat: true, TimeFormat: time.RFC3339})
	w.Write(LvInfo, time.Now(), "error msg", []byte(`"k1":"v1"`))
	w.Write(LvInfo, time.Now(), "error msg", []byte(`"k1":"v1"`))
	w.Write(LvFault, time.Now(), "fault msg")
}

func Test_FileWrite(t *testing.T) {
	w := NewWriter(&WriterCfg{OpenMsgFormat: true, File: WriterFile{
		Enable:   true,
		Filename: "./out.log",
		MaxSize:  1,
		Compress: true,
	}})
	for i := 0; i <= 5000; i++ {
		w.Write(LvDebug, time.Now(), fmt.Sprintf("debug msg %d", i), []byte(`"k1":"v1"`))
		w.Write(LvInfo, time.Now(), fmt.Sprintf("error msg %d", i), []byte(`"k1":"v1", "k2":"v2"`))
		w.Write(LvWarn, time.Now(), fmt.Sprintf("warn msg %d", i), []byte(`"k1:"v1", "k3": "v3"`))
		w.Write(LvError, time.Now(), fmt.Sprintf("error msg %d", i))
		w.Write(LvFault, time.Now(), fmt.Sprintf("fault msg %d", i))
	}
	w = NewWriter(&WriterCfg{OpenMsgFormat: true, TimeFormat: time.RFC3339, File: WriterFile{
		Enable:   true,
		Filename: "./out.log",
		MaxSize:  1,
		Compress: true,
	}})
	for i := 0; i <= 5000; i++ {
		w.Write(LvDebug, time.Now(), fmt.Sprintf("debug msg %d", i), []byte(`"k1":"v1"`))
		w.Write(LvInfo, time.Now(), fmt.Sprintf("error msg %d", i), []byte(`"k1":"v1", "k2":"v2"`))
		w.Write(LvWarn, time.Now(), fmt.Sprintf("warn msg %d", i), []byte(`"k1:"v1", "k3": "v3"`))
		w.Write(LvError, time.Now(), fmt.Sprintf("error msg %d", i))
		w.Write(LvFault, time.Now(), fmt.Sprintf("fault msg %d", i))
	}
}
