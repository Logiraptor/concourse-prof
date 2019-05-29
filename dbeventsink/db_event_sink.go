package dbeventsink

import (
	"github.com/Logiraptor/concourse-prof/events"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type dbEventSink struct {
	db *gorm.DB
}

func NewDbEventSink(connection string) (*dbEventSink, error) {
	db, err := gorm.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&events.LogEvent{}, &events.StatusEvent{}, &events.ErrorEvent{})
	db.AutoMigrate(&events.InitializeTaskEvent{}, &events.StartTaskEvent{}, &events.FinishTaskEvent{})
	db.AutoMigrate(&events.InitializeGetEvent{}, &events.StartGetEvent{}, &events.FinishGetEvent{})
	db.AutoMigrate(&events.InitializePutEvent{}, &events.StartPutEvent{}, &events.FinishPutEvent{})
	if db.Error != nil {
		return nil, err
	}

	return &dbEventSink{
		db: db,
	}, nil
}

func (d dbEventSink) Close() error {
	return d.db.Close()
}

func (d dbEventSink) OnInitializeTaskEvent(event *events.InitializeTaskEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnStartTaskEvent(event *events.StartTaskEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnFinishTaskEvent(event *events.FinishTaskEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnInitializePutEvent(event *events.InitializePutEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnStartPutEvent(event *events.StartPutEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnFinishPutEvent(event *events.FinishPutEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnInitializeGetEvent(event *events.InitializeGetEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnStartGetEvent(event *events.StartGetEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnFinishGetEvent(event *events.FinishGetEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnLogEvent(event *events.LogEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnStatusEvent(event *events.StatusEvent) error {
	d.db.Create(event)
	return d.db.Error
}

func (d dbEventSink) OnErrorEvent(event *events.ErrorEvent) error {
	d.db.Create(event)
	return d.db.Error
}
