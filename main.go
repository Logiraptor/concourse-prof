package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"net/http"
	"regexp"

	"database/sql"

	"github.com/donovanhide/eventsource"
	_ "github.com/lib/pq"
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
	db *sql.DB
}

func (p processor) insertEvent(event interface{}) {
	fmt.Printf("inserting %T event\n", event)
	switch v := event.(type) {
	case *LogEvent:
		_, err := p.db.Exec("INSERT INTO log_events VALUES ($1, $2, to_timestamp($3), $4);",
			v.Origin.Id,
			v.Origin.Source,
			v.Time,
			v.Payload)
		if err != nil {
			log.Fatal(err)
			return
		}
	case *StatusEvent:
	case *InitializeGetEvent:
	case *StartGetEvent:
	case *FinishGetEvent:
	case *InitializePutEvent:
	case *StartPutEvent:
	case *FinishPutEvent:
	case *InitializeTaskEvent:
	case *StartTaskEvent:
	case *FinishTaskEvent:

	default:
		log.Fatalf("Cannot insert event: %T", event)
	}
}

func (p processor) processBuildEvents(build string) {
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

	log.Println("Token is", string(match[0]), string(url))
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
		default:
			panic("Cannot handle event " + base.Event)
		}

		json.Unmarshal(base.Data, &actual)
		p.insertEvent(actual)
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

	p.processFlyTable(func(values []string) bool {
		return p.processBuilds(pipeline, job, values)
	}, "builds", "-j", pipeline+"/"+job)
	return true
}

func (p processor) processBuilds(pipeline, job string, values []string) bool {
	if values[3] == "succeeded" {
		build := values[0]

		lastLogTime, err := time.Parse("2006-01-02@15:04:05-0700", values[4])
		if err != nil {
			log.Fatal(err)
			return false
		}
		p.analyzeLogs(pipeline, job, build, lastLogTime)
		return false
	}
	return true
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

	p := processor{
		db: db,
	}
	p.processFlyTable(p.processPipelines, "pipelines")
}

func (p processor) analyzeLogs(pipeline, job, build string, lastLogTime time.Time) {
	fmt.Println(pipeline, job, build, lastLogTime)
	p.processBuildEvents(build)
}
