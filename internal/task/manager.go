package task

import (
	"fmt"
	"log-analyzer/internal/analyzer"
)

type TaskManager struct {
	nextID int
	tasks  map[int]Task
}

func NewTaskManager() *TaskManager {
	var manager TaskManager
	manager.tasks = make(map[int]Task)
	manager.nextID = 0
	return &manager
}

func (t *TaskManager) CreateTask(input []string) (Task, error) {
	if len(input) == 0 {
		return Task{}, fmt.Errorf("input logs cannot be empty")
	}
	for _, v := range input {
		_, err := analyzer.ParseLogEntry([]byte(v))
		if err != nil {
			return Task{}, fmt.Errorf("invalid log entry: %w", err)
		}
	}
	inputCopied := make([]string, len(input))
	copy(inputCopied, input)
	task := Task{
		ID:     t.nextID,
		Input:  inputCopied,
		Status: Pending,
	}
	t.nextID++
	t.tasks[task.ID] = task
	return task, nil
}

func (t *TaskManager) GetTask(id int) (Task, error) {
	if task, ok := t.tasks[id]; ok {
		return task, nil
	}
	return Task{}, fmt.Errorf("task with ID %d not found", id)
}

func (t *TaskManager) ProcessTask(id int) error {
	task, err := t.GetTask(id)
	if err != nil {
		return err
	}
	if task.Status != Pending {
		return fmt.Errorf("task is either on process or completed")
	}
	task.Status = Processing
	t.tasks[id] = task

	var a analyzer.Analyzer
	ar, err := a.Analyze(task.Input)
	if err != nil {
		task.Status = Pending
		t.tasks[id] = task
		return err
	}
	task.Result = ar
	task.Status = Completed
	t.tasks[id] = task
	return nil
}
