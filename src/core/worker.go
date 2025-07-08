package core

import (
	"net/http"
	"sync"

	"github.com/ZekiromNet/load-tester/src/methods"
)

type Result struct {
	Success bool
	Status  int
	Err     error
}

func Worker(cfg Config, jobs <-chan int, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{Timeout: cfg.Timeout}

	for range jobs {
		var status int
		var err error

		switch cfg.Method {
		case "GET":
			status, err = methods.DoGet(client, cfg.URL)
		case "POST":
			status, err = methods.DoPost(client, cfg.URL)
		default:
			err = ErrUnsupportedMethod(cfg.Method)
		}

		results <- Result{
			Success: err == nil,
			Status:  status,
			Err:     err,
		}
	}
}

type unsupportedMethodError string

func (e unsupportedMethodError) Error() string {
	return "unsupported method: " + string(e)
}

func ErrUnsupportedMethod(method string) error {
	return unsupportedMethodError(method)
}
