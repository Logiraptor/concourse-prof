package processor

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Logiraptor/concourse-prof/events"
)

type flyClient interface {
	ListPipelines() ([]string, error)
	ListJobs(pipeline string) ([]string, error)
	ListBuilds(pipeline, job string) ([]string, error)
	ListEvents(pipeline, job, build string) ([]interface{}, error)
}

type downloadUI interface {
	ShowError(err error)
	SetBarTotal(name string, total int)
	IncrementBar(name string, value int)
}

type eventSink interface {
	OnInitializeTaskEvent(event *events.InitializeTaskEvent) error
	OnStartTaskEvent(event *events.StartTaskEvent) error
	OnFinishTaskEvent(event *events.FinishTaskEvent) error
	OnInitializePutEvent(event *events.InitializePutEvent) error
	OnStartPutEvent(event *events.StartPutEvent) error
	OnFinishPutEvent(event *events.FinishPutEvent) error
	OnInitializeGetEvent(event *events.InitializeGetEvent) error
	OnStartGetEvent(event *events.StartGetEvent) error
	OnFinishGetEvent(event *events.FinishGetEvent) error
	OnLogEvent(event *events.LogEvent) error
	OnStatusEvent(event *events.StatusEvent) error
	OnErrorEvent(event *events.ErrorEvent) error
}

type processor struct {
	client   flyClient
	ui       downloadUI
	sink     eventSink
	jobQueue chan func() error
}

func NewProcessor(client flyClient, ui downloadUI, sink eventSink, wg *sync.WaitGroup) processor {
	p := processor{
		client:   client,
		ui:       ui,
		sink:     sink,
		jobQueue: make(chan func() error),
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go p.worker(wg)
	}
	return p
}

func (p processor) worker(wg *sync.WaitGroup) {
	for job := range p.jobQueue {
		err := job()
		if err != nil {
			p.ui.ShowError(err)
		}
	}
	wg.Done()
}

func (p processor) Close() {
	close(p.jobQueue)
}

func (p processor) ProcessPipeline(pipeline string) error {
	jobs, err := p.client.ListJobs(pipeline)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		err := p.ProcessJob(pipeline, job)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p processor) ProcessJob(pipeline, job string) error {
	builds, err := p.client.ListBuilds(pipeline, job)
	if err != nil {
		return err
	}

	for _, build := range builds {
		err := p.ProcessBuild(pipeline, job, build)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p processor) ProcessBuild(pipeline, job, build string) error {
	name := fmt.Sprintf("%s/%s/%s", pipeline, job, build)
	p.jobQueue <- func() error {
		events, err := p.client.ListEvents(pipeline, job, build)
		if err != nil {
			return err
		}
		total := len(events)
		p.ui.SetBarTotal(name, total)
		for _, event := range events {
			err = dispatch(event, p.sink)
			if err != nil {
				return err
			}
			p.ui.IncrementBar(name, 1)
		}
		return nil
	}
	return nil
}

func dispatch(event interface{}, sink eventSink) error {
	switch v := event.(type) {
	case *events.InitializeTaskEvent:
		return sink.OnInitializeTaskEvent(v)
	case *events.StartTaskEvent:
		return sink.OnStartTaskEvent(v)
	case *events.FinishTaskEvent:
		return sink.OnFinishTaskEvent(v)
	case *events.InitializePutEvent:
		return sink.OnInitializePutEvent(v)
	case *events.StartPutEvent:
		return sink.OnStartPutEvent(v)
	case *events.FinishPutEvent:
		return sink.OnFinishPutEvent(v)
	case *events.InitializeGetEvent:
		return sink.OnInitializeGetEvent(v)
	case *events.StartGetEvent:
		return sink.OnStartGetEvent(v)
	case *events.FinishGetEvent:
		return sink.OnFinishGetEvent(v)
	case *events.LogEvent:
		return sink.OnLogEvent(v)
	case *events.StatusEvent:
		return sink.OnStatusEvent(v)
	case *events.ErrorEvent:
		return sink.OnErrorEvent(v)
	default:
		return errors.New(fmt.Sprintf("Cannot dispatch event type: %T", v))
	}
}
