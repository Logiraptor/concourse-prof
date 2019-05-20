package main

type InitializePutEvent struct {
	Time      int64
	BuildInfo BuildInfo `gorm:"EMBEDDED"`
	Origin    struct {
		Origin string `json:"id"`
	} `gorm:"EMBEDDED"`
}

type StartPutEvent struct {
	Time      int64
	BuildInfo BuildInfo `gorm:"EMBEDDED"`
	Origin    struct {
		Origin string `json:"id"`
	} `gorm:"EMBEDDED"`
}

type FinishPutEvent struct {
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
