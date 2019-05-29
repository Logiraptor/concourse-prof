package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Logiraptor/concourse-prof/events"
	"github.com/dgrijalva/jwt-go"
	"github.com/donovanhide/eventsource"
)

type apiClient struct {
	concourseUrl string
	url          string
	token        string
}

func NewApiClient(concourseUrl, url, token string) (*apiClient, error) {
	if token == "" {
		return nil, errors.New("token is blank")
	}

	parts := strings.Split(token, " ")
	if len(parts) != 2 {
		return nil, errors.New("token should be formatted as 'bearer xxxx'")
	}

	_, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		return nil, token.Claims.Valid()
	})

	if err != nil && err.Error() != jwt.ErrInvalidKeyType.Error() {
		return nil, errors.New("token validation failed with error " + err.Error())
	}

	return &apiClient{concourseUrl: concourseUrl, url: url, token: token}, nil
}

func (a apiClient) newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Building request to url", url)
	req.Header.Add("Authorization", a.token)
	req.Header.Add("X-Concourse-URL", a.concourseUrl)

	return req, nil
}

func (a apiClient) listResource(url string, fieldName string) ([]string, error) {
	req, err := a.newRequest(url)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	output := make([]string, 0, len(results))
	for _, result := range results {
		output = append(output, fmt.Sprint(result[fieldName]))
	}
	return output, nil
}

func (a apiClient) ListPipelines() ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/teams/main/pipelines", a.url)
	return a.listResource(url, "name")
}

func (a apiClient) ListJobs(pipeline string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/teams/main/pipelines/%s/jobs", a.url, pipeline)
	return a.listResource(url, "name")
}

func (a apiClient) ListBuilds(pipeline, job string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/teams/main/pipelines/%s/jobs/%s/builds?limit=50", a.url, pipeline, job)
	return a.listResource(url, "id")
}

func (a apiClient) ListEvents(pipeline, job, build string) ([]interface{}, error) {
	baseURL, err := url.Parse(a.url)
	if err != nil {
		return nil, err
	}

	var eventsUrl = baseURL
	eventsUrl.Path = "/api/v1/builds/" + build + "/events"

	req, err := a.newRequest(eventsUrl.String())
	if err != nil {
		return nil, err
	}

	stream, err := eventsource.SubscribeWithRequest("", req)
	if err != nil {
		return nil, err
	}

	var planUrl = baseURL
	planUrl.Path = "/api/v1/builds/" + build + "/plan"
	req, err = a.newRequest(planUrl.String())
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var idLookup = make(map[string]string)
	var plan map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&plan)
	if err != nil {
		return nil, err
	}

	var traverse func(interface{})
	traverse = func(plan interface{}) {
		switch v := plan.(type) {
		case []interface{}:
			for _, a := range v {
				traverse(a)
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
			for _, a := range v {
				traverse(a)
			}
		}
	}
	traverse(plan)

	eventsInStream := []interface{}{}

	for event := range stream.Events {
		if event.Event() == "end" {
			stream.Close()
			break
		}
		buildInfo := events.BuildInfo{
			Pipeline: pipeline,
			Job:      job,
			Build:    build,
			EventId:  event.Id(),
		}
		actual, err := decodeEvent(event, buildInfo, idLookup)
		if err != nil {
			return nil, err
		}
		eventsInStream = append(eventsInStream, actual)
	}

	return eventsInStream, nil
}

func decodeEvent(event eventsource.Event, buildInfo events.BuildInfo, idLookup map[string]string) (interface{}, error) {
	var base events.BaseEvent
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
		actual = &events.InitializeTaskEvent{BuildInfo: buildInfo}
	case "start-task":
		actual = &events.StartTaskEvent{BuildInfo: buildInfo}
	case "finish-task":
		actual = &events.FinishTaskEvent{BuildInfo: buildInfo}

	case "initialize-put":
		actual = &events.InitializePutEvent{BuildInfo: buildInfo}
	case "start-put":
		actual = &events.StartPutEvent{BuildInfo: buildInfo}
	case "finish-put":
		actual = &events.FinishPutEvent{BuildInfo: buildInfo}

	case "initialize-get":
		actual = &events.InitializeGetEvent{BuildInfo: buildInfo}
	case "start-get":
		actual = &events.StartGetEvent{BuildInfo: buildInfo}
	case "finish-get":
		actual = &events.FinishGetEvent{BuildInfo: buildInfo}

	case "log":
		actual = &events.LogEvent{BuildInfo: buildInfo}
	case "status":
		actual = &events.StatusEvent{BuildInfo: buildInfo}
	case "error":
		actual = &events.ErrorEvent{BuildInfo: buildInfo}
	default:
		return nil, errors.New("Cannot handle event " + base.Event)
	}

	err = json.Unmarshal(base.Data, &actual)
	if err != nil {
		return nil, err
	}
	return actual, nil
}
