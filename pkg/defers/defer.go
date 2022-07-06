package defers

import (
	"sync"
)

var globalDefers = &Defer{fns: make([]func() error, 0)}

type Defer struct {
	sync.Mutex
	fns []func() error
}

func NewDefer() *Defer {
	return &Defer{
		fns: make([]func() error, 0),
	}
}

func (d *Defer) Register(fns ...func() error) {
	d.Lock()
	defer d.Unlock()
	d.fns = append(d.fns, fns...)
}

func (d *Defer) Done() {
	d.Lock()
	defer d.Unlock()
	for i := len(d.fns) - 1; i >= 0; i-- {
		_ = d.fns[i]()
	}
}

// Register 注册一个defer函数
func Register(fns ...func() error) {
	globalDefers.Register(fns...)
}

// Clean 清除
func Done() {
	globalDefers.Done()
}
