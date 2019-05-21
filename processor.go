package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/donovanhide/eventsource"
	"github.com/jinzhu/gorm"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

type processor struct {
	db       *gorm.DB
	jobQueue chan func() error
	progress *mpb.Progress
}

func (p processor) fly(args ...string) ([]string, error) {
	cmd := exec.Command("fly", append([]string{"-t", "eagle"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(out))

	output := []string{}
	for scanner.Scan() {
		values := strings.Fields(scanner.Text())
		output = append(output, values[0])
	}
	err = scanner.Err()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (p processor) processPipeline(pipeline string) error {
	jobs, err := p.fly("jobs", "-p", pipeline)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		err := p.processJob(pipeline, job)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p processor) processJob(pipeline, job string) error {
	p.jobQueue <- func() error {
		builds, err := p.fly("builds", "-j", pipeline+"/"+job)
		if err != nil {
			return err
		}

		name := pipeline + "/" + job
		bar := p.progress.AddBar(int64(len(builds)),
			mpb.PrependDecorators(
				decor.Name(name+" ", decor.WCSyncWidth),
				decor.Percentage(decor.WCSyncWidth),
			),
		)
		for _, build := range builds {
			err := p.processBuild(pipeline, job, build)
			if err != nil {
				return err
			}
			bar.IncrBy(1)
		}
		return nil
	}

	return nil
}

func (p processor) processBuild(pipeline, job, build string) error {
	cmd := exec.Command("fly", "-t", "eagle", "curl", "/api/v1/builds/"+build+"/events", "--print-and-exit")
	curlCmd, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	tokenRegexp := regexp.MustCompile("Bearer [^\"]+")
	match := tokenRegexp.FindSubmatch(curlCmd)
	if match == nil {
		return errors.New("Could not find auth token in " + string(curlCmd))
	}

	curlArgs := bytes.Fields(curlCmd)
	url := curlArgs[len(curlArgs)-1]

	req, _ := http.NewRequest("GET", string(url), nil)
	req.Header.Add("Authorization", string(match[0]))

	stream, err := eventsource.SubscribeWithRequest("", req)
	if err != nil {
		return err
	}

	req, _ = http.NewRequest("GET", "https://eagle.ci.cf-app.com/api/v1/builds/"+build+"/plan", nil)
	req.Header.Add("Authorization", string(match[0]))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	var idLookup = make(map[string]string)
	var plan map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&plan)
	if err != nil {
		return err
	}

	var traverse func(interface{})
	traverse = func(plan interface{}) {
		switch v := plan.(type) {
		case []interface{}:
			for _, p := range v {
				traverse(p)
			}
		case map[string]interface{}:
			if id, ok := v["id"]; ok {
				var name string
				var typ string
				for k := range v {
					if k != "id" {
						typ = k
						break
					}
				}
				if local, ok := v[typ].(map[string]interface{}); ok {
					name, _ = local["name"].(string)
				}
				idLookup[id.(string)] = name + "/" + typ
			}
			for _, p := range v {
				traverse(p)
			}
		}
	}
	traverse(plan)

	for event := range stream.Events {
		if event.Event() == "end" {
			stream.Close()
			break
		}
		buildInfo := BuildInfo{
			Pipeline: pipeline,
			Job:      job,
			Build:    build,
			EventId:  event.Id(),
		}
		actual, err := decodeEvent(event, buildInfo, idLookup)
		if err != nil {
			return err
		}
		p.db.Create(actual)
		if p.db.Error != nil {
			return p.db.Error
		}
	}

	return nil
}

func decodeEvent(event eventsource.Event, buildInfo BuildInfo, idLookup map[string]string) (interface{}, error) {
	var base BaseEvent
	err := json.Unmarshal([]byte(event.Data()), &base)
	if err != nil {
		return nil, err
	}

	var baseData interface{}
	json.Unmarshal(base.Data, &baseData)
	if asMap, ok := baseData.(map[string]interface{}); ok {
		if origin, ok := asMap["origin"]; ok {
			if originAsMap, ok := origin.(map[string]interface{}); ok {
				if id, ok := originAsMap["id"]; ok {
					originAsMap["id"] = idLookup[id.(string)]
				}
			}
		}
	}
	base.Data, _ = json.Marshal(baseData)

	var actual interface{}
	switch base.Event {
	case "initialize-task":
		actual = &InitializeTaskEvent{BuildInfo: buildInfo}
	case "start-task":
		actual = &StartTaskEvent{BuildInfo: buildInfo}
	case "finish-task":
		actual = &FinishTaskEvent{BuildInfo: buildInfo}

	case "initialize-put":
		actual = &InitializePutEvent{BuildInfo: buildInfo}
	case "start-put":
		actual = &StartPutEvent{BuildInfo: buildInfo}
	case "finish-put":
		actual = &FinishPutEvent{BuildInfo: buildInfo}

	case "initialize-get":
		actual = &InitializeGetEvent{BuildInfo: buildInfo}
	case "start-get":
		actual = &StartGetEvent{BuildInfo: buildInfo}
	case "finish-get":
		actual = &FinishGetEvent{BuildInfo: buildInfo}

	case "log":
		actual = &LogEvent{BuildInfo: buildInfo}
	case "status":
		actual = &StatusEvent{BuildInfo: buildInfo}
	case "error":
		actual = &ErrorEvent{BuildInfo: buildInfo}
	default:
		return nil, errors.New("Cannot handle event " + base.Event)
	}

	err = json.Unmarshal(base.Data, &actual)
	if err != nil {
		return nil, err
	}
	return actual, nil
}
