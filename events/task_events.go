package events

type InitializeTaskEvent struct {
	Time   int64
	Config struct {
		Platform string
		Image    string
		Run      struct {
			Path string
			Args []string `gorm:"-"`
			Dir  string
		} `gorm:"EMBEDDED"`
		Inputs []struct {
			Name string
			Path string
		}
	} `gorm:"EMBEDDED"`
	BuildInfo BuildInfo `gorm:"EMBEDDED"`
	Origin    struct {
		Origin string `json:"id"`
	} `gorm:"EMBEDDED"`
}

type StartTaskEvent struct {
	Time   int64
	Config struct {
		Platform string
		Image    string
		Run      struct {
			Path string
			Args []string `gorm:"-"`
			Dir  string
		} `gorm:"EMBEDDED"`
		Inputs []struct {
			Name string
			Path string
		}
	} `gorm:"EMBEDDED"`
	BuildInfo BuildInfo `gorm:"EMBEDDED"`
	Origin    struct {
		Origin string `json:"id"`
	} `gorm:"EMBEDDED"`
}

type FinishTaskEvent struct {
	Time       int64
	ExitStatus int       `json:"exit_status"`
	BuildInfo  BuildInfo `gorm:"EMBEDDED"`
	Origin     struct {
		Origin string `json:"id"`
	} `gorm:"EMBEDDED"`
}
