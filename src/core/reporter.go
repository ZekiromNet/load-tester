package core

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

var ansiRegexp = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

func stripANSI(str string) string {
	return ansiRegexp.ReplaceAllString(str, "")
}

func StatusReporter(stats *Stats, cfg Config, done <-chan struct{}) {
	totalRequests := cfg.NumRequests
	start := time.Now()
	for {
		select {
		case <-time.After(200 * time.Millisecond):
			successful := stats.Successful
			failed := stats.Failed
			total := stats.Total
			elapsed := time.Since(start)
			if total > 0 {
				currentRPS := float64(total) / elapsed.Seconds()
				successRate := float64(successful) / float64(total) * 100
				failRate := float64(failed) / float64(total) * 100
				if cfg.Verbose {
					pterm.Info.Printfln("[%v] Progress: %.1f%% (%d/%d) | Success: %d (%.1f%%) | Failed: %d (%.1f%%) | RPS: %.2f",
						elapsed.Round(time.Second), float64(total)/float64(cfg.NumRequests)*100, total, totalRequests,
						successful, successRate, failed, failRate, currentRPS)
					statusCodes := stats.GetStatusCodes()
					if len(statusCodes) > 0 {
						statusStr := ""
						first := true
						for code, count := range statusCodes {
							if !first {
								statusStr += " | "
							}
							percentage := float64(count) / float64(successful) * 100
							statusStr += fmt.Sprintf("%d: %d (%.1f%%)", code, count, percentage)
							first = false
						}
					}
				} else {
					pterm.Print("\r\033[K")
					elapsedStr := pterm.FgLightWhite.Sprintf("[%s]", elapsed.Round(time.Second).String())
					progressStr := pterm.FgCyan.Sprintf("%d/%d", total, totalRequests)
					successStr := pterm.FgGreen.Sprintf("%d (%.1f%%)", successful, successRate)
					failStr := pterm.FgRed.Sprintf("%d (%.1f%%)", failed, failRate)
					rpsStr := pterm.FgYellow.Sprintf("%.2f", currentRPS)
					barText := pterm.Bold.Sprintf("%s │ %s │ %s │ %s │ %s │", elapsedStr, progressStr, successStr, failStr, rpsStr)
					width, _, _ := pterm.GetTerminalSize()
					barLen := len([]rune(stripANSI(barText)))
					progressBar := ""
					minBarWidth := 5
					barWidth := width - barLen - 3
					if barWidth < minBarWidth {
						barWidth = minBarWidth
					}
					if width-barLen-3 >= barWidth {
						progress := float64(total) / float64(totalRequests)
						filled := int(progress * float64(barWidth))
						if filled < 0 {
							filled = 0
						}
						if filled > barWidth {
							filled = barWidth
						}
						empty := barWidth - filled
						if empty < 0 {
							empty = 0
						}
						progressBar = " [" + pterm.FgGreen.Sprintf("%s", strings.Repeat("█", filled)) + pterm.FgGray.Sprintf("%s", strings.Repeat("░", empty)) + "]"
					}
					finalBar := barText + progressBar
					finalLen := len([]rune(finalBar))
					if finalLen < width {
						pad := width - finalLen
						finalBar = finalBar + pterm.FgGray.Sprintf("%s", string(make([]rune, pad))) + " "
					}
					pterm.Print(finalBar)
				}
			}
		}
	}
}

func PrintFinalReport(stats *Stats, cfg Config, elapsed time.Duration) {
	successful := stats.Successful
	failed := stats.Failed
	if !cfg.Verbose {
		pterm.Print("\r\033[K\r")
	} else {
		pterm.Println()
	}
	pterm.FgCyan.Printf("  Total requests: ")
	pterm.FgLightWhite.Println(cfg.NumRequests)
	pterm.FgGreen.Printf("  Successful: ")
	pterm.FgLightWhite.Println(successful)
	pterm.FgRed.Printf("  Failed: ")
	pterm.FgLightWhite.Println(failed)
	pterm.FgCyan.Printf("  Success rate: ")
	pterm.FgGreen.Printf("%.2f%%\n", float64(successful)/float64(cfg.NumRequests)*100)
	if elapsed.Seconds() > 0 {
		pterm.FgCyan.Printf("  Requests per second: ")
		pterm.FgYellow.Printf("%.2f\n", float64(cfg.NumRequests)/elapsed.Seconds())
	} else {
		pterm.FgCyan.Printf("  Requests per second: ")
		pterm.FgYellow.Println("N/A")
	}
	pterm.Println()

	for code, count := range stats.GetStatusCodes() {
		var codeStyle *pterm.Style
		if code >= 200 && code < 300 {
			codeStyle = pterm.NewStyle(pterm.FgGreen)
		} else if code >= 400 && code < 600 {
			codeStyle = pterm.NewStyle(pterm.FgRed)
		} else {
			codeStyle = pterm.NewStyle(pterm.FgYellow)
		}
		codeStr := codeStyle.Sprint(fmt.Sprintf("  %d:", code))
		pterm.Printf("%s %d requests (%.2f%%)\n", codeStr, count, float64(count)/float64(successful)*100)
	}
}
