// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package file is a file source. Expected format is json
package file

import (
	"io"
	"os"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"
	"gopkg.in/yaml.v3"
)

type file struct {
	stopped       atomic.Bool
	exit          chan bool
	path          string
	enableWatcher bool
	fw            *fsnotify.Watcher
}

func (f *file) Read() (source.SourceData, error) {
	fh, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	b, err := io.ReadAll(fh)
	if err != nil {
		return nil, err
	}
	cs := source.NewBytesSourceData(source.PriorityFile, b, yaml.Unmarshal)
	return cs, nil
}

func (f *file) Name() string {
	return "file"
}

func (f *file) Changeable() bool {
	return f.enableWatcher
}

func (f *file) Watch() (<-chan source.SourceData, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	_ = fw.Add(f.path)
	f.fw = fw
	change := make(chan source.SourceData, 1)
	xgo.Go(func() {
		defer func() {
			close(change)
		}()
		for {
			chg, err := f.watch()
			if err != nil {
				logger.Errorf("fault to watch file, err: %v", err)
				continue
			}
			if chg == nil {
				return
			}
			change <- chg
		}
	}, nil)
	return change, nil
}

func (f *file) watch() (source.SourceData, error) {
	select {
	case <-f.exit:
		return nil, nil
	default:
	}

	// try get the event
	select {
	case event, _ := <-f.fw.Events:
		if event.Op == fsnotify.Rename {
			// check existence of file, and add watch again
			_, err := os.Stat(event.Name)
			if err == nil || os.IsExist(err) {
				_ = f.fw.Add(event.Name)
			}
		}

		c, err := f.Read()
		if err != nil {
			return nil, err
		}
		return c, nil
	case err := <-f.fw.Errors:
		return nil, err
	case <-f.exit:
		return nil, nil
	}
}

func (f *file) Close() error {
	if f.stopped.CompareAndSwap(false, true) {
		close(f.exit)
	}
	return nil
}

func NewSource(path string, watchable bool) source.Source {
	return &file{
		path:          path,
		enableWatcher: watchable,
	}
}
