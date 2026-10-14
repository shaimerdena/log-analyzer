package task

import (
	"log-analyzer/internal/analyzer"
)

type TaskStatus string

const (
	Pending TaskStatus = "Pending"
	Processing TaskStatus = "Processing"
    Completed  TaskStatus = "Completed"
)

type Task struct{
	ID int
	Input []string
	Status TaskStatus
	Result analyzer.AnalysisResult
}