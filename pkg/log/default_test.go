package log

import (
	"testing"
	"time"
)

func Test_logger(t *testing.T) {
	Debug("Debug", "Debug")
	Debugf("this %s", "Debugf")

	Info("Info", "Info")
	Infof("this %s", "Infof")

	Warn("Warn", "Earn")
	Warnf("this %s", "Warnf")

	Error("Error", "Error")
	Errorf("this %s", "Errorf")

	//Fatal("Fatal", "Fatal")
	//Fatalf("this %s", "Fatalf")
	//Fatalw("Fatalw", "type", "Fatalw")
}

func Test(t *testing.T) {
	start := time.Now()
	defer func() {
		Debugf("free: %.3f", time.Since(start).Seconds())
	}()
	time.Sleep(time.Millisecond * 3)
}
