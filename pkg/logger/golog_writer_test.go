package logger

import (
	"fmt"
	"testing"
)

func Test_StderrWrite(t *testing.T) {
	w := NewWriter(&WriterCfg{OpenMsgFormat: true})
	w.Write(LvDebug, "debug msg", "ext1")
	w.Write(LvInfo, "error msg", "ext1", "ext2")
	w.Write(LvWarn, "warn msg", "ext1", "ext2", "ext3")
	w.Write(LvError, "error msg", "ext1", "ext2", "ext3")
	w.Write(LvFault, "fault msg")
}

func Test_FileWrite(t *testing.T) {
	w := NewWriter(&WriterCfg{OpenMsgFormat: true, File: WriterFile{
		Enable:   true,
		Filename: "./out.log",
		MaxSize:  1,
		Compress: true,
	}})
	for i := 0; i <= 5000; i++ {
		w.Write(LvDebug, fmt.Sprintf("debug msg %d", i), "k1", "v1")
		w.Write(LvInfo, fmt.Sprintf("error msg %d", i), "k1", "v1", "k2", "v2")
		w.Write(LvWarn, fmt.Sprintf("warn msg %d", i), "k1", "v1", "k3", "v3")
		w.Write(LvError, fmt.Sprintf("error msg %d", i), "k1")
		w.Write(LvFault, fmt.Sprintf("fault msg %d", i))
	}
}
