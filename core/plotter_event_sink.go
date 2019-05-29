package main

import (
	"time"

	"github.com/Logiraptor/concourse-prof/events"
)

type interval struct {
	Init   time.Time
	Start  time.Time
	Finish time.Time
}

type plotterEventSink struct {
	Intervals map[string]*interval
}

func (p *plotterEventSink) Reset() {
	p.Intervals = make(map[string]*interval)
}

func (p *plotterEventSink) interval(origin string) *interval {
	if p.Intervals == nil {
		p.Reset()
	}
	if i, ok := p.Intervals[origin]; ok {
		return i
	} else {
		i := &interval{}
		p.Intervals[origin] = i
		return i
	}
}

func (p *plotterEventSink) OnInitializeTaskEvent(event *events.InitializeTaskEvent) error {
	p.interval(event.Origin.Origin).Init = time.Unix(event.Time, 0)
	return nil
}

func (p *plotterEventSink) OnStartTaskEvent(event *events.StartTaskEvent) error {
	p.interval(event.Origin.Origin).Start = time.Unix(event.Time, 0)
	return nil
}

func (p *plotterEventSink) OnFinishTaskEvent(event *events.FinishTaskEvent) error {
	p.interval(event.Origin.Origin).Finish = time.Unix(event.Time, 0)
	return nil
}

func (p *plotterEventSink) OnInitializePutEvent(event *events.InitializePutEvent) error {
	p.interval(event.Origin.Origin).Init = time.Unix(event.Time, 0)
	return nil
}

func (p *plotterEventSink) OnStartPutEvent(event *events.StartPutEvent) error {
	p.interval(event.Origin.Origin).Start = time.Unix(event.Time, 0)
	return nil
}

func (p *plotterEventSink) OnFinishPutEvent(event *events.FinishPutEvent) error {
	p.interval(event.Origin.Origin).Finish = time.Unix(event.Time, 0)
	return nil
}

func (p *plotterEventSink) OnInitializeGetEvent(event *events.InitializeGetEvent) error {
	p.interval(event.Origin.Origin).Init = time.Unix(event.Time, 0)
	return nil
}

func (p *plotterEventSink) OnStartGetEvent(event *events.StartGetEvent) error {
	p.interval(event.Origin.Origin).Start = time.Unix(event.Time, 0)
	return nil
}

func (p *plotterEventSink) OnFinishGetEvent(event *events.FinishGetEvent) error {
	p.interval(event.Origin.Origin).Finish = time.Unix(event.Time, 0)
	return nil
}

func (p *plotterEventSink) OnLogEvent(event *events.LogEvent) error {
	return nil
}

func (p *plotterEventSink) OnStatusEvent(event *events.StatusEvent) error {
	return nil
}

func (p *plotterEventSink) OnErrorEvent(event *events.ErrorEvent) error {
	return nil
}
