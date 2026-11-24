package task

import (
	"fmt"
	"log-analyzer/internal/analyzer"
	"sync"
)

const (
	WorkerCount = 3
	QueueSize   = 5
)

type TaskManager struct {
	nextID  int
	tasks   map[int]Task
	queue   chan int
	mu      sync.RWMutex
	stopped bool
	workers sync.WaitGroup
}

func NewTaskManager() *TaskManager {
	manager := &TaskManager{
		tasks: make(map[int]Task),
		queue: make(chan int, QueueSize),
	}
	manager.startWorkers(WorkerCount)
	return manager
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

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		return Task{}, fmt.Errorf("task manager is stopped")
	}

	task := Task{
		ID:     t.nextID,
		Input:  inputCopied,
		Status: Pending,
	}
	t.tasks[task.ID] = task
	t.nextID++
	return cloneTask(task), nil
}

func (t *TaskManager) GetTask(id int) (Task, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	task, ok := t.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task with ID %d not found", id)
	}
	return cloneTask(task), nil
}

// pending -> queued
func (t *TaskManager) SubmitTask(id int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		return fmt.Errorf("task manager is stopped")
	}

	task, ok := t.tasks[id]
	if !ok {
		return fmt.Errorf("task with ID %d not found", id)
	}
	if task.Status != Pending {
		return fmt.Errorf("task %d cannot be submitted from state %s", id, task.Status)
	}
	task.Status = Queued
	t.tasks[id] = task

	select {
	case t.queue <- id:
		return nil
	default:
		task.Status = Pending
		t.tasks[id] = task
		return fmt.Errorf("task queue is full")
	}
}

func (t *TaskManager) processTask(
	id int,
	expectedStatus TaskStatus,
	errorStatus TaskStatus,
	rejectIfStopped bool,
) error {
	t.mu.Lock()

	if rejectIfStopped && t.stopped {
		t.mu.Unlock()
		return fmt.Errorf("task manager is stopped")
	}

	task, ok := t.tasks[id]
	if !ok {
		t.mu.Unlock()
		return fmt.Errorf("task with ID %d not found", id)
	}
	if task.Status != expectedStatus {
		t.mu.Unlock()
		return fmt.Errorf("task %d cannot be processed from state %s", id, task.Status)
	}

	task.Status = Processing
	t.tasks[id] = task

	input := append([]string(nil), task.Input...)

	t.mu.Unlock()

	var a analyzer.Analyzer
	result, err := a.Analyze(input)

	t.mu.Lock()
	defer t.mu.Unlock()

	task = t.tasks[id]

	if err != nil {
		task.Status = errorStatus
		task.Err = err.Error()
		t.tasks[id] = task
		return err
	}

	task.Result = result
	task.Status = Completed
	task.Err = ""
	t.tasks[id] = task
	return nil
}

// Pending -> Completed/Pending
func (t *TaskManager) ProcessTask(id int) error {
	return t.processTask(id, Pending, Pending, true)
}

func (t *TaskManager) startWorkers(count int) {
	if count <= 0 {
		return
	}
	t.workers.Add(count)
	for i := 0; i < count; i++ {
		go t.worker()
	}
}

func (t *TaskManager) worker() {
	defer t.workers.Done()

	for id := range t.queue {
		t.processQueuedTask(id)
	}
}

// Queued -> Completed/Failed
func (t *TaskManager) processQueuedTask(id int) {
	_ = t.processTask(id, Queued, Failed, false)
}

func (t *TaskManager) Shutdown() {
	t.mu.Lock()

	if t.stopped {
		t.mu.Unlock()
		t.workers.Wait()
		return
	}
	t.stopped = true
	close(t.queue)

	t.mu.Unlock()
	t.workers.Wait()
}

func cloneTask(task Task) Task {
	task.Input = append([]string(nil), task.Input...)

	task.Result.ErrorDetails = append(
		[]analyzer.LogEntry(nil),
		task.Result.ErrorDetails...,
	)
	if task.Result.ErrorsByUnit != nil {
		copiedMap := make(map[string]int, len(task.Result.ErrorsByUnit))

		for key, value := range task.Result.ErrorsByUnit {
			copiedMap[key] = value
		}
		task.Result.ErrorsByUnit = copiedMap
	}
	return task
}
