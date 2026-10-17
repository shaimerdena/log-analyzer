package analyzer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type LogEntry struct {
	Priority int
	Message  string
	Unit     string
	Source   string
}

type AnalysisResult struct {
	TotalEntries int
	ErrorCount   int
	WarningCount int
	ErrorDetails []LogEntry
	ErrorsByUnit map[string]int
}

type journalRecord struct {
	Priority string `json:"PRIORITY"`
	Message  string `json:"MESSAGE"`
	Unit     string `json:"_SYSTEMD_UNIT"`
	Source   string `json:"SYSLOG_IDENTIFIER"`
}

type Analyzer struct{}

func parseRecord(data []byte) (journalRecord, error) {
	var record journalRecord
	err := json.Unmarshal(data, &record)
	return record, err
}

func toLogEntry(record journalRecord) (LogEntry, error) {
	var entry LogEntry
	entry.Message = record.Message
	entry.Source = record.Source
	entry.Unit = record.Unit
	val, err := strconv.Atoi(record.Priority)
	if err != nil {
		return LogEntry{}, err
	}
	entry.Priority = val
	return entry, nil
}

func validateEntry(entry LogEntry) error {
	if entry.Priority < 0 || entry.Priority > 7 {
		return fmt.Errorf("Priority should be between 0 and 7")
	}
	if strings.TrimSpace(entry.Message) == "" {
		return fmt.Errorf("Message cannot be empty")
	}
	return nil
}

func ParseLogEntry(data []byte) (LogEntry, error) {
	record, err := parseRecord(data)
	if err != nil {
		return LogEntry{}, err
	}
	entry, err := toLogEntry(record)
	if err != nil {
		return LogEntry{}, err
	}
	err = validateEntry(entry)
	if err != nil {
		return LogEntry{}, err
	}
	return entry, nil
}

func (a *Analyzer) Analyze(input []string) (AnalysisResult, error) {
	var ar AnalysisResult
	if len(input) == 0 {
		return AnalysisResult{}, fmt.Errorf("input is empty")
	}
	ar.ErrorsByUnit = make(map[string]int)
	for _, v := range input {
		log, err := ParseLogEntry([]byte(v))
		if err != nil {
			return AnalysisResult{}, fmt.Errorf("parse log entry: %w", err)
		}
		ar.TotalEntries++
		if log.Priority <= 3 {
			ar.ErrorCount++
			ar.ErrorDetails = append(ar.ErrorDetails, log)
			ar.ErrorsByUnit[log.Unit]++
		}
		if log.Priority == 4 {
			ar.WarningCount++
		}
	}
	return ar, nil
}
