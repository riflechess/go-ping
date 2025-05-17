package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tatsushid/go-fastping"
)

type PingResult struct {
	Hostname   string
	IP         net.IP
	RTT        time.Duration
	Status     string
	History    []bool // true = success, false = failure
	RTTHistory []time.Duration
}

type stringList []string

func (s *stringList) String() string {
	return fmt.Sprint(*s)
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var mu sync.Mutex
var results map[string]*PingResult = make(map[string]*PingResult)
var hostOrder []string

// func addToHistory(history string, success bool) string {
// 	symbol := "\033[32m*\033[0m" // green
// 	if !success {
// 		symbol = "\033[31m-\033[0m" // red
// 	}
// 	newHistory := symbol + history
// 	if len([]rune(newHistory)) > 10*len(symbol) {
// 		// Truncate to 10 colored asterisks
// 		runes := []rune(newHistory)
// 		newHistory = string(runes[:10*len(symbol)])
// 	}
// 	return newHistory
// }

func formatRTT(rtt time.Duration) string {
	if rtt == 0 {
		return "--------"
	}
	return fmt.Sprintf("%.3fms", float64(rtt.Microseconds())/1000)
}

func pingHost(host string, timeout time.Duration) {
	p := fastping.NewPinger()
	p.MaxRTT = timeout

	ra, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		mu.Lock()
		results[host] = &PingResult{Hostname: host, Status: "Resolve Error"}
		mu.Unlock()
		return
	}

	mu.Lock()
	result, ok := results[host]
	if !ok {
		result = &PingResult{
			Hostname: host,
			IP:       ra.IP,
		}
		results[host] = result
	}
	result.Status = "Pending"
	result.IP = ra.IP // in case DNS changes
	mu.Unlock()

	p.AddIPAddr(ra)

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		mu.Lock()
		result.RTT = rtt
		result.Status = "Responding"
		result.History = append([]bool{true}, result.History...)
		result.RTTHistory = append([]time.Duration{rtt}, result.RTTHistory...)

		if len(result.RTTHistory) > 10 {
			result.RTTHistory = result.RTTHistory[:10]
		}

		if len(result.History) > 10 {
			result.History = result.History[:10]
		}
		mu.Unlock()
	}

	p.OnIdle = func() {
		mu.Lock()
		if result.Status != "Responding" {
			result.Status = "Timeout"
			result.RTT = 0
			result.History = append([]bool{false}, result.History...)
			result.RTTHistory = append([]time.Duration{0}, result.RTTHistory...)

			if len(result.RTTHistory) > 10 {
				result.RTTHistory = result.RTTHistory[:10]
			}

			if len(result.History) > 10 {
				result.History = result.History[:10]
			}
		}
		mu.Unlock()
	}

	_ = p.Run()
}

func renderHistory(history []bool) string {
	result := ""
	for _, ok := range history {
		if ok {
			result += "\033[32m+\033[0m" // green
		} else {
			result += "\033[31m-\033[0m" // red
		}
	}
	return result
}

func renderRTTSparkline(history []time.Duration) string {
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	maxRTT := time.Duration(0)
	for _, rtt := range history {
		if rtt > maxRTT {
			maxRTT = rtt
		}
	}
	if maxRTT == 0 {
		return "----------" // all timeouts
	}

	result := ""
	for _, rtt := range history {
		if rtt == 0 {
			result += "-" // timeout
		} else {
			level := int((rtt * time.Duration(len(blocks)-1)) / maxRTT)
			if level >= len(blocks) {
				level = len(blocks) - 1
			}
			result += string(blocks[level])
		}
	}
	return result
}

func draw() {
	fmt.Print("\033[H\033[2J") // Clear screen

	// Header
	fmt.Printf("%-18s %-18s %-12s %-10s %-10s %s\n",
		"Host", "IP", "Status", "RTT", "History", "RTT Graph")
	fmt.Println(strings.Repeat("-", 90))

	mu.Lock()
	defer mu.Unlock()

	for _, host := range hostOrder {
		r := results[host]
		if r == nil {
			continue
		}

		statusColor := ""
		switch r.Status {
		case "Responding":
			statusColor = "\033[32m" // green
		case "Timeout", "Resolve Error":
			statusColor = "\033[31m" // red
		default:
			statusColor = "\033[33m" // yellow
		}

		fmt.Printf("%-18s %-18s %s%-12s\033[0m %-10s %-10s %s\n",
			r.Hostname,
			r.IP,
			statusColor, r.Status,
			formatRTT(r.RTT),
			renderHistory(r.History),
			renderRTTSparkline(r.RTTHistory),
		)
	}

	// Footer Summary
	summary := fmt.Sprintf(
		"\n%s | Total hosts: %d | Ctrl+C to stop",
		time.Now().Format("2006-01-02 15:04:05"),
		len(hostOrder),
	)
	fmt.Println(summary)
}

func expandSynthetic(pattern string) ([]string, error) {
	re := regexp.MustCompile(`\[(.*?)\]`)

	// Find all range matches (including ranges and selections)
	matches := re.FindAllStringSubmatchIndex(pattern, -1)
	if len(matches) == 0 {
		return []string{pattern}, nil
	}

	expanded := []string{""}
	lastIndex := 0

	for _, match := range matches {
		startBracket, endBracket := match[0], match[1]
		contentStart, contentEnd := match[2], match[3]

		prefix := pattern[lastIndex:startBracket]
		rangeContent := pattern[contentStart:contentEnd]

		var rangeValues []string

		// Check if it's a range (e.g., 01:03) or selection (e.g., dca,dcb)
		if rangeParts := regexp.MustCompile(`^(\d+):(\d+)$`).FindStringSubmatch(rangeContent); len(rangeParts) == 3 {
			start, _ := strconv.Atoi(rangeParts[1])
			end, _ := strconv.Atoi(rangeParts[2])
			width := len(rangeParts[1]) // Preserve leading zeros

			for i := start; i <= end; i++ {
				pad := fmt.Sprintf("%0*d", width, i)
				rangeValues = append(rangeValues, pad)
			}
		} else {
			// It's a selection, split by commas
			rangeValues = append(rangeValues, splitAndTrim(rangeContent, ",")...)
		}

		var newExpanded []string
		for _, e := range expanded {
			for _, r := range rangeValues {
				newExpanded = append(newExpanded, e+prefix+r)
			}
		}
		expanded = newExpanded

		lastIndex = endBracket
	}

	// Append any remaining suffix
	suffix := pattern[lastIndex:]
	for i := range expanded {
		expanded[i] += suffix
	}

	return expanded, nil
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range regexp.MustCompile(sep).Split(s, -1) {
		parts = append(parts, part)
	}
	return parts
}

func main() {
	var timeoutMs int
	var intervalSec int
	var syntheticPatterns stringList
	var printHosts bool

	flag.IntVar(&timeoutMs, "timeout", 5000, "timeout in milliseconds per ping")
	flag.Var(&syntheticPatterns, "s", "synthetic host pattern like host-[001:004] (repeatable)")
	flag.IntVar(&intervalSec, "i", 1, "seconds between pings")
	flag.BoolVar(&printHosts, "printhosts", false, "just print the expanded host list and exit")
	flag.Parse()

	var hosts []string
	if len(syntheticPatterns) > 0 {
		for _, pattern := range syntheticPatterns {
			expanded, err := expandSynthetic(pattern)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error in -s pattern '%s': %v\n", pattern, err)
				os.Exit(1)
			}
			hosts = append(hosts, expanded...)
		}
	} else if flag.NArg() > 0 {
		hosts = flag.Args()
	} else {
		fmt.Println("Usage: go-ping [-i seconds] [-s pattern] host1 host2 ...")
		os.Exit(1)
	}

	// if -printhosts is set, dump the list and quit
	if printHosts {
		for _, h := range hosts {
			fmt.Println(h)
		}
		return
	}

	hostOrder = hosts

	// Handle CTRL-C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Println("\nExiting...")
		os.Exit(0)
	}()

	// Continuous ping loop
	for {
		var wg sync.WaitGroup
		for _, host := range hosts {
			wg.Add(1)
			go func(h string) {
				defer wg.Done()
				pingHost(h, time.Duration(timeoutMs)*time.Millisecond)
			}(host)
		}
		wg.Wait()

		draw()
		time.Sleep(time.Duration(intervalSec) * time.Second)
	}
}
