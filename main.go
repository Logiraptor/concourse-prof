package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/vbauerster/mpb/v4"
)

type BaseEvent struct {
	Event   string
	Version string
	Data    json.RawMessage
}

type BuildInfo struct {
	Pipeline, Job, Build string
	EventId              string
}

func main() {
	db, err := gorm.Open("postgres", "host=localhost sslmode=disable")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	db.AutoMigrate(&LogEvent{}, &StatusEvent{}, &ErrorEvent{})
	db.AutoMigrate(&InitializeTaskEvent{}, &StartTaskEvent{}, &FinishTaskEvent{})
	db.AutoMigrate(&InitializeGetEvent{}, &StartGetEvent{}, &FinishGetEvent{})
	db.AutoMigrate(&InitializePutEvent{}, &StartPutEvent{}, &FinishPutEvent{})
	if db.Error != nil {
		log.Fatal(err)
		return
	}

	wg := &sync.WaitGroup{}
	progress := mpb.New(mpb.WithWaitGroup(wg))

	p := processor{
		db:       db,
		jobQueue: make(chan func() error),
		progress: progress,
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(wg, p.jobQueue)
	}
	pipelines, err := p.fly("pipelines")
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, pipeline := range pipelines {
		p.processPipeline(pipeline)
	}
	close(p.jobQueue)
	progress.Wait()
}

func worker(wg *sync.WaitGroup, jobs chan func() error) {
	for job := range jobs {
		err := job()
		if err != nil {
			fmt.Println(err)
		}
	}
	wg.Done()
}
