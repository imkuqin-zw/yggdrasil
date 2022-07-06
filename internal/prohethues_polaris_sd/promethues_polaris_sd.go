package prohethues_polaris_sd

import (
	"log"
)

func Run() {
	sd := NewSD()
	fw := NewFileWriter()
	instances, err := sd.GetAllInstance()
	if err != nil {
		log.Fatalf("fault to get all instance, err: %+v", err)
	}
	fw.Write(instances)
	ch, err := sd.Watch()
	if err != nil {
		log.Fatalf("fault to watch, err: %+v", err)
	}
	for {
		select {
		case instances, ok := <-ch:
			if !ok {
				return
			}
			fw.Write(instances)
		}
	}
}
