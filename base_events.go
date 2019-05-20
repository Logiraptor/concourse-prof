package main

type StatusEvent struct {
	Time      int64
	Status    string
	BuildInfo BuildInfo `gorm:"EMBEDDED"`
}

type LogEvent struct {
	Time      int64
	Payload   string
	BuildInfo BuildInfo `gorm:"EMBEDDED"`
	Origin    struct {
		Origin string `json:"id"`
		Source string
	} `gorm:"EMBEDDED"`
}

type ErrorEvent struct {
	Message   string
	BuildInfo BuildInfo `gorm:"EMBEDDED"`
}
