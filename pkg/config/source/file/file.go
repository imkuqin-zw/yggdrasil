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
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"
	"go.uber.org/atomic"
	"gopkg.in/yaml.v2"
)

type file struct {
	stopped       atomic.Bool
	exit          chan bool
	path          string
	enableWatcher bool
	fw            *fsnotify.Watcher
}

func (f *file) Read() (types.ConfigSourceData, error) {
	fh, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	b, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil, err
	}
	cs := source.NewBytesSourceData(types.ConfigPriorityFile, b, yaml.Unmarshal)
	return cs, nil
}

func (f *file) Name() string {
	return "file"
}

func (f *file) Changeable() bool {
	return f.enableWatcher
}

func (f *file) Watch() (<-chan types.ConfigSourceData, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	_ = fw.Add(f.path)
	f.fw = fw
	change := make(chan types.ConfigSourceData, 1)
	xgo.Go(func() {
		defer func() {
			close(change)
		}()
		for {
			chg, err := f.watch()
			if err != nil {
				log.Errorf("fault to watch file, err: %v", err)
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

func (f *file) watch() (types.ConfigSourceData, error) {
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
	if f.stopped.CAS(false, true) {
		close(f.exit)
	}
	return nil
}

func NewSource(path string, watchable bool) types.ConfigSource {
	return &file{
		path:          path,
		enableWatcher: watchable,
	}
}
