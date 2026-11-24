package task

import (
	"log-analyzer/internal/analyzer"
)

type TaskStatus string

const (
	Pending    TaskStatus = "Pending"
	Queued     TaskStatus = "Queued"
	Processing TaskStatus = "Processing"
	Completed  TaskStatus = "Completed"
	Failed     TaskStatus = "Failed"
)

type Task struct {
	ID     int
	Input  []string
	Status TaskStatus
	Result analyzer.AnalysisResult
	Err    string
}
