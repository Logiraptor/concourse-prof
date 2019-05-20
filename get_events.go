package main

type InitializeGetEvent struct {
	Time      int64
	BuildInfo BuildInfo `gorm:"EMBEDDED"`
	Origin    struct {
		Origin string `json:"id"`
	} `gorm:"EMBEDDED"`
}

type StartGetEvent struct {
	Time      int64
	BuildInfo BuildInfo `gorm:"EMBEDDED"`
	Origin    struct {
		Origin string `json:"id"`
	} `gorm:"EMBEDDED"`
}

type FinishGetEvent struct {
	Time       int64
	ExitStatus int       `json:"exit_status"`
	BuildInfo  BuildInfo `gorm:"EMBEDDED"`
	Origin     struct {
		Origin string `json:"id"`
	} `gorm:"EMBEDDED"`
	// Version  map[string]string
	// Metadata []struct {
	// 	Name  string
	// 	Value string
	// }
}
