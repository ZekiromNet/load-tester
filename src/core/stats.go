package core

import (
	"sync"
	"sync/atomic"
)

type Stats struct {
	Successful  int64
	Failed      int64
	Total       int64
	StatusCodes map[int]*int64
	StatusMutex sync.RWMutex
}

func NewStats() *Stats {
	return &Stats{
		StatusCodes: make(map[int]*int64),
	}
}

func (s *Stats) AddStatusCode(code int) {
	s.StatusMutex.Lock()
	defer s.StatusMutex.Unlock()
	if s.StatusCodes[code] == nil {
		var counter int64
		s.StatusCodes[code] = &counter
	}
	atomic.AddInt64(s.StatusCodes[code], 1)
}

func (s *Stats) GetStatusCodes() map[int]int64 {
	s.StatusMutex.RLock()
	defer s.StatusMutex.RUnlock()
	result := make(map[int]int64)
	for code, counter := range s.StatusCodes {
		result[code] = atomic.LoadInt64(counter)
	}
	return result
}

func (s *Stats) Update(res Result) {
	atomic.AddInt64(&s.Total, 1)
	if res.Success {
		atomic.AddInt64(&s.Successful, 1)
		s.AddStatusCode(res.Status)
	} else {
		atomic.AddInt64(&s.Failed, 1)
	}
}
