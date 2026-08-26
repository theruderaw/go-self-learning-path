package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp time.Time
	Message   string
}

type LogStat struct {
	Count   int
	Entries []LogEntry
}

type LogStatItem struct {
	Level string
	Count int
}

var logs = make(map[string]LogStat)

func ParseLogs(line string) error {
	parts := strings.Fields(line)

	if len(parts) < 4 {
		return fmt.Errorf("invalid log line: %s", line)
	}

	timestamp, err := time.Parse(
		"2006-01-02T15:04:05",
		parts[0]+"T"+parts[1],
	)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %s", line)
	}

	entry := LogEntry{
		Timestamp: timestamp,
		Message:   strings.Join(parts[3:], " "),
	}

	log := logs[parts[2]]
	log.Count++
	log.Entries = append(log.Entries, entry)
	logs[parts[2]] = log

	return nil
}

func ReadLogs(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		err := ParseLogs(line)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	return scanner.Err()
}

func SortStats(logs map[string]LogStat) []LogStatItem {
	stats := []LogStatItem{}

	for level, log := range logs {
		stats = append(stats, LogStatItem{
			Level: level,
			Count: log.Count,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}

func SortEntries(entries []LogEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})
}

func DisplayStats(stats []LogStatItem) {
	fmt.Println("Log Summary")
	fmt.Println("----------------")
	fmt.Println()

	total := 0

	for _, stat := range stats {
		fmt.Printf("%s: %d\n", stat.Level, stat.Count)
		total += stat.Count
	}

	fmt.Println("----------------")
	fmt.Printf("Count: %d\n", total)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: golog /path/to/log.file")
		return
	}

	err := ReadLogs(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	stats := SortStats(logs)
	DisplayStats(stats)

	fmt.Println()
	fmt.Println("Logs by datetime")
	fmt.Println("----------------")

	for level, log := range logs {
		SortEntries(log.Entries)

		fmt.Printf("\n%s:\n", level)

		for _, entry := range log.Entries {
			fmt.Printf(
				"%s %s\n",
				entry.Timestamp.Format("2006-01-02T15:04:05"),
				entry.Message,
			)
		}
	}
}
