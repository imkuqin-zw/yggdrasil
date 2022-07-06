package prohethues_polaris_sd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
)

type Job struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

type FileWriterCfg struct {
	Filepath string
}

type FileWriter struct {
	format   string
	filePath string
}

func NewFileWriter() *FileWriter {
	cfg := &FileWriterCfg{}
	if err := config2.Scan("file", cfg); err != nil {
		log.Fatalf("fault to load target_file config, err: %s", err)
	}
	if cfg.Filepath == "" {
		cfg.Filepath = "polaris_targets.json"
	}
	ext := filepath.Ext(cfg.Filepath)
	if ext != ".yaml" && ext != ".yml" && ext != ".json" {
		log.Fatalf("ext of file must be json or yaml")
	}
	return &FileWriter{
		format:   ext,
		filePath: cfg.Filepath,
	}
}

func (fw *FileWriter) Write(instances []Instance) {
	jobs := make([]Job, len(instances))
	for i, item := range instances {
		jobs[i] = Job{
			Targets: []string{item.Endpoint},
			Labels: map[string]string{
				"job":       item.ServiceName,
				"zone":      item.Zone,
				"campus":    item.Campus,
				"region":    item.Region,
				"namespace": item.Namespace,
			},
		}
	}
	data, _ := json.Marshal(jobs)
	backoff := time.Second * 20
	for {
		err := os.WriteFile(fw.filePath, data, 0666)
		if err == nil {
			log.Infof("update file: %s", fw.filePath)
			break
		}
		log.Warnf("fault to write job to file, err: %+v", err)
		if backoff < time.Minute*10 {
			backoff *= 2
		}
	}
}
