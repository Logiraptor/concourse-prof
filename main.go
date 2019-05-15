package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"net/http"
	"regexp"

	"database/sql"

	"github.com/donovanhide/eventsource"
	_ "github.com/lib/pq"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

type Task struct {
	name   string
	start  time.Time
	end    time.Time
	events []LogEvent
}

type BaseEvent struct {
	Event   string
	Version string
	Data    json.RawMessage
}

type ErrorEvent struct {
	Message string
	Origin  struct {
		Id string
	}
}

type StatusEvent struct {
	Time   int64
	Status string
}

type LogEvent struct {
	Time   int64
	Origin struct {
		Id     string
		Source string
	}
	Payload string
}

type InitializeTaskEvent struct {
	Time   int64
	Origin struct {
		Id string
	}
	Config struct {
		Platform string
		Image    string
		Run      struct {
			Path string
			Args []string
			Dir  string
		}
		Inputs []struct {
			Name string
			Path string
		}
	}
}

type StartTaskEvent struct {
	Time   int64
	Origin struct {
		Id string
	}
	Config struct {
		Platform string
		Image    string
		Run      struct {
			Path string
			Args []string
			Dir  string
		}
		Inputs []struct {
			Name string
			Path string
		}
	}
}

type FinishTaskEvent struct {
	Time       int64
	ExitStatus int `json:"exit_status"`
	Origin     struct {
		Id string
	}
}

type InitializePutEvent struct {
	Origin struct {
		Id string
	}
	Time int64
}

type StartPutEvent struct {
	Origin struct {
		Id string
	}
	Time int64
}

type FinishPutEvent struct {
	Time       int64
	ExitStatus int `json:"exit_status"`
	Origin     struct {
		Id string
	}
	Version  map[string]string
	Metadata []struct {
		Name  string
		Value string
	}
}

type InitializeGetEvent struct {
	Origin struct {
		Id string
	}
	Time int64
}

type StartGetEvent struct {
	Origin struct {
		Id string
	}
	Time int64
}

type FinishGetEvent struct {
	Time       int64
	ExitStatus int `json:"exit_status"`
	Origin     struct {
		Id string
	}
	Version  map[string]string
	Metadata []struct {
		Name  string
		Value string
	}
}

type processor struct {
	db       *sql.DB
	jobQueue chan func()
	progress *mpb.Progress
	wg       *sync.WaitGroup
}

func (p processor) insertEvent(pipeline, job, build string, event interface{}) {
	// fmt.Printf("inserting %T event\n", event)
	switch v := event.(type) {
	case *LogEvent:
		_, err := p.db.Exec("INSERT INTO log_events VALUES ($1, $2, $3, $4, $5, to_timestamp($6), $7);",
			pipeline,
			job,
			build,
			v.Origin.Id,
			v.Origin.Source,
			v.Time,
			v.Payload)
		if err != nil {
			log.Fatal(err)
			return
		}
	case *StatusEvent:
		_, err := p.db.Exec("INSERT INTO status_events VALUES ($1, $2, $3, $4, to_timestamp($5));",
			pipeline,
			job,
			build,
			v.Status,
			v.Time)
		if err != nil {
			log.Fatal(err)
			return
		}
	case *InitializeGetEvent:
	case *StartGetEvent:
	case *FinishGetEvent:
	case *InitializePutEvent:
	case *StartPutEvent:
	case *FinishPutEvent:
	case *InitializeTaskEvent:
	case *StartTaskEvent:
	case *FinishTaskEvent:
	case *ErrorEvent:

	default:
		log.Fatalf("Cannot insert event: %T", event)
	}
}

func (p processor) processBuildEvents(pipeline, job, build string) {
	cmd := exec.Command("fly", "-t", "eagle", "curl", "/api/v1/builds/"+build+"/events", "--print-and-exit")
	curlCmd, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	tokenRegexp := regexp.MustCompile("Bearer [^\"]+")
	match := tokenRegexp.FindSubmatch(curlCmd)
	if match == nil {
		log.Fatal("Could not find auth token in ", string(curlCmd))
		return
	}

	curlArgs := bytes.Fields(curlCmd)
	url := curlArgs[len(curlArgs)-1]

	req, _ := http.NewRequest("GET", string(url), nil)
	req.Header.Add("Authorization", string(match[0]))

	stream, err := eventsource.SubscribeWithRequest("", req)
	if err != nil {
		log.Fatal(err)
		return
	}

	for event := range stream.Events {
		if event.Event() == "end" {
			stream.Close()
			break
		}
		var base BaseEvent
		json.Unmarshal([]byte(event.Data()), &base)

		var actual interface{}
		switch base.Event {
		case "initialize-task":
			actual = &InitializeTaskEvent{}
		case "start-task":
			actual = &StartTaskEvent{}
		case "finish-task":
			actual = &FinishTaskEvent{}

		case "initialize-put":
			actual = &InitializePutEvent{}
		case "start-put":
			actual = &StartPutEvent{}
		case "finish-put":
			actual = &FinishPutEvent{}

		case "initialize-get":
			actual = &InitializeGetEvent{}
		case "start-get":
			actual = &StartGetEvent{}
		case "finish-get":
			actual = &FinishGetEvent{}

		case "log":
			actual = &LogEvent{}
		case "status":
			actual = &StatusEvent{}
		case "error":
			actual = &ErrorEvent{}
		default:
			panic("Cannot handle event " + base.Event)
		}

		json.Unmarshal(base.Data, &actual)
		p.insertEvent(pipeline, job, build, actual)
	}
}

func (p processor) processFlyTable(onLine func([]string) bool, args ...string) {
	cmd := exec.Command("fly", append([]string{"-t", "eagle"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
		return
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(out))

	for scanner.Scan() {
		values := strings.Fields(scanner.Text())
		if !onLine(values) {
			return
		}
	}
}

func (p processor) processPipelines(values []string) bool {
	pipeline := values[0]

	p.processFlyTable(func(values []string) bool {
		return p.processJobs(pipeline, values)
	}, "jobs", "-p", pipeline)
	return true
}

func (p processor) processJobs(pipeline string, values []string) bool {
	job := values[0]

	p.jobQueue <- func() {
		builds := []string{}
		p.processFlyTable(func(values []string) bool {
			builds = append(builds, values[0])
			return true
		}, "builds", "-j", pipeline+"/"+job)

		if len(builds) > 4 {
			builds = builds[:4]
		}

		name := pipeline + "/" + job
		bar := p.progress.AddBar(int64(len(builds)),
			mpb.PrependDecorators(
				decor.Name(name+" ", decor.WCSyncWidth),
				decor.Percentage(decor.WCSyncWidth),
			),
		)
		for _, build := range builds {
			bar.IncrBy(1)
			p.analyzeLogs(pipeline, job, build)
		}
	}
	return true
}

func (p processor) analyzeLogs(pipeline, job, build string) {
	p.processBuildEvents(pipeline, job, build)
}

func main() {
	db, err := sql.Open("postgres", "host=localhost sslmode=disable")
	if err != nil {
		log.Fatal(err)
		return
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	wg := &sync.WaitGroup{}
	progress := mpb.New(mpb.WithWaitGroup(wg))

	p := processor{
		db:       db,
		jobQueue: make(chan func()),
		progress: progress,
		wg:       wg,
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(wg, p.jobQueue)
	}
	p.processFlyTable(p.processPipelines, "pipelines")
	close(p.jobQueue)
	progress.Wait()
}

func worker(wg *sync.WaitGroup, jobs chan func()) {
	for job := range jobs {
		job()
	}
	wg.Done()
}
