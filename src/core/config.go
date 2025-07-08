package core

import "time"

type Config struct {
	URL            string
	Method         string
	NumRequests    int
	MaxConcurrency int
	Timeout        time.Duration
	StatusInterval time.Duration
	Verbose        bool
}
