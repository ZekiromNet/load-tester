package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/ZekiromNet/load-tester/src/core"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "load-tester",
		Usage: "A fast, clean HTTP load tester supporting GET and POST.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Usage:    "Target URL to test",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "method",
				Usage: "HTTP method: GET or POST",
				Value: "GET",
			},
			&cli.IntFlag{
				Name:  "requests",
				Usage: "Total number of requests to send",
				Value: 10000,
			},
			&cli.IntFlag{
				Name:  "max-concurrency",
				Usage: "Maximum number of concurrent requests",
				Value: 1000,
			},
			&cli.DurationFlag{
				Name:  "timeout",
				Usage: "Request timeout (e.g. 5s, 1m)",
				Value: 0,
			},
			&cli.DurationFlag{
				Name:  "status-interval",
				Usage: "Status reporting interval (e.g. 1s, 2s)",
				Value: 250 * time.Millisecond,
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable detailed status reporting instead of a single-line status.",
			},
		},
		Action: func(c *cli.Context) error {
			cfg := core.Config{
				URL:            c.String("url"),
				Method:         c.String("method"),
				NumRequests:    c.Int("requests"),
				MaxConcurrency: c.Int("max-concurrency"),
				Timeout:        c.Duration("timeout"),
				StatusInterval: c.Duration("status-interval"),
				Verbose:        c.Bool("verbose"),
			}
			pterm.Println(cfg)
			verbose := c.Bool("verbose")
			RunLoadTest(cfg, verbose)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func RunLoadTest(cfg core.Config, verbose bool) {
	start := time.Now()

	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgGreen)).WithFullWidth().Println(" Load Tester ")
	pterm.Println()
	pterm.FgCyan.Printf("  Target URL: ")
	pterm.FgLightWhite.Println(cfg.URL)
	pterm.FgCyan.Printf("  Method: ")
	pterm.FgLightWhite.Println(cfg.Method)
	pterm.FgCyan.Printf("  Number of requests: ")
	pterm.FgLightWhite.Println(cfg.NumRequests)
	pterm.FgCyan.Printf("  Max concurrency: ")
	pterm.FgLightWhite.Println(cfg.MaxConcurrency)
	pterm.FgCyan.Printf("  Request timeout: ")
	pterm.FgLightWhite.Println(cfg.Timeout)
	pterm.FgCyan.Printf("  Status reporting: ")
	pterm.FgLightWhite.Printf("Every %v\n", cfg.StatusInterval)
	pterm.Println()

	stats := core.NewStats()
	jobs := make(chan int, cfg.NumRequests)
	results := make(chan core.Result, cfg.NumRequests)
	var wg sync.WaitGroup

	done := make(chan struct{})
	go core.StatusReporter(stats, cfg, done)

	for i := 0; i < cfg.MaxConcurrency; i++ {
		wg.Add(1)
		go core.Worker(cfg, jobs, results, &wg)
	}

	go func() {
		for i := 0; i < cfg.NumRequests; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		stats.Update(res)
	}

	close(done)
	elapsed := time.Since(start)
	core.PrintFinalReport(stats, cfg, elapsed)
}
