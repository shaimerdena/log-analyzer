package task

import (
	"sync"
	"testing"
	"time"
)

func waitForStatus(
	t *testing.T,
	manager *TaskManager,
	id int,
	expected TaskStatus,
	timeout time.Duration,
) Task {
	t.Helper()
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		task, err := manager.GetTask(id)
		if err != nil {
			t.Fatalf("GetTask() error: %v", err)
		}
		if task.Status == expected {
			return task
		}
		time.Sleep(10 * time.Millisecond)
	}

	task, err := manager.GetTask(id)
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}
	t.Fatalf("timeout waiting for task %d status %s; current status: %s", id, expected, task.Status)

	return Task{}
}

func TestSubmitTaskAsync(t *testing.T) {
	manager := NewTaskManager()
	defer manager.Shutdown()

	input := []string{
		`{"PRIORITY":"3","MESSAGE":"Service failed","_SYSTEMD_UNIT":"demo.service","SYSLOG_IDENTIFIER":"demo"}`,
		`{"PRIORITY":"4","MESSAGE":"High load","_SYSTEMD_UNIT":"demo.service","SYSLOG_IDENTIFIER":"demo"}`,
		`{"PRIORITY":"6","MESSAGE":"Service started","_SYSTEMD_UNIT":"demo.service","SYSLOG_IDENTIFIER":"demo"}`,
	}

	task, err := manager.CreateTask(input)
	if err != nil {
		t.Fatalf("CreateTask() error: %v", err)
	}
	if task.Status != Pending {
		t.Fatalf("expected status Pending, got %s", task.Status)
	}

	err = manager.SubmitTask(task.ID)
	if err != nil {
		t.Fatalf("SubmitTask() error: %v", err)
	}

	task = waitForStatus(
		t,
		manager,
		task.ID,
		Completed,
		3*time.Second,
	)

	if task.Result.TotalEntries != 3 {
		t.Errorf("expected 3 total entries, got %d", task.Result.ErrorCount)
	}
	if task.Result.ErrorCount != 1 {
		t.Errorf("expected 1 error, got %d", task.Result.ErrorCount)
	}
	if task.Result.WarningCount != 1 {
		t.Errorf("expectedd 1 warning, got %d", task.Result.WarningCount)
	}
	if task.Err != "" {
		t.Errorf("expected empty error, got %q", task.Err)
	}
}

func TestMultipleTasks(t *testing.T) {
	manager := NewTaskManager()
	defer manager.Shutdown()

	inputs := [][]string{
		{
			`{"PRIORITY":"3","MESSAGE":"Error 1","_SYSTEMD_UNIT":"api.service","SYSLOG_IDENTIFIER":"api"}`,
		},
		{
			`{"PRIORITY":"4","MESSAGE":"Warning 1","_SYSTEMD_UNIT":"db.service","SYSLOG_IDENTIFIER":"db"}`,
			`{"PRIORITY":"3","MESSAGE":"Error 2","_SYSTEMD_UNIT":"db.service","SYSLOG_IDENTIFIER":"db"}`,
		},
		{
			`{"PRIORITY":"6","MESSAGE":"Service started","_SYSTEMD_UNIT":"worker.service","SYSLOG_IDENTIFIER":"worker"}`,
			`{"PRIORITY":"3","MESSAGE":"Error 3","_SYSTEMD_UNIT":"worker.service","SYSLOG_IDENTIFIER":"worker"}`,
			`{"PRIORITY":"4","MESSAGE":"Warning 2","_SYSTEMD_UNIT":"worker.service","SYSLOG_IDENTIFIER":"worker"}`,
		},
		{
			`{"PRIORITY":"3","MESSAGE":"Error 4","_SYSTEMD_UNIT":"api.service","SYSLOG_IDENTIFIER":"api"}`,
		},
		{
			`{"PRIORITY":"4","MESSAGE":"Warning 3","_SYSTEMD_UNIT":"db.service","SYSLOG_IDENTIFIER":"db"}`,
		},
	}

	var taskIDs []int

	for _, input := range inputs {
		task, err := manager.CreateTask(input)
		if err != nil {
			t.Fatalf("CreateTask() error: %v", err)
		}
		if err := manager.SubmitTask(task.ID); err != nil {
			t.Fatalf("SubmitTask() error: %v", err)
		}

		taskIDs = append(taskIDs, task.ID)
	}

	for _, id := range taskIDs {
		task := waitForStatus(
			t,
			manager,
			id,
			Completed,
			3*time.Second,
		)
		if task.Result.TotalEntries == 0 {
			t.Errorf("task %d returned empty result", id)
		}
	}

	expectedErrors := []int{1, 1, 1, 1, 0}
	expectedWarnings := []int{0, 1, 1, 0, 1}

	for i, id := range taskIDs {
		task, err := manager.GetTask(id)
		if err != nil {
			t.Fatalf("GetTask(%d) error: %v", id, err)
		}
		if task.Result.ErrorCount != expectedErrors[i] {
			t.Errorf(
				"task %d: expected %d errors, got %d",
				id,
				expectedErrors[i],
				task.Result.ErrorCount,
			)
		}
		if task.Result.WarningCount != expectedWarnings[i] {
			t.Errorf(
				"task %d: expected %d warnings, got %d",
				id,
				expectedWarnings[i],
				task.Result.WarningCount,
			)
		}
	}
}

func TestDuplicateSubmissionRejected(t *testing.T) {
	manager := NewTaskManager()
	defer manager.Shutdown()

	input := []string{
		`{"priority":"3","message":"test error","unit":"test.service","source":"test"}`,
	}
	task, err := manager.CreateTask(input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	if err := manager.SubmitTask(task.ID); err != nil {
		t.Fatalf("first submission failed: %v", err)
	}
	if err := manager.SubmitTask(task.ID); err == nil {
		t.Fatal("expected duplicate submission to be rejected")
	}
	if err := manager.ProcessTask(task.ID); err == nil {
		t.Fatal("expected ProcessTask on queued task to be rejected")
	}
}

func TestQueueFull(t *testing.T) {
	manager := &TaskManager{
		tasks: make(map[int]Task),
		queue: make(chan int, QueueSize),
	}
	defer manager.Shutdown()
	for i := 0; i < QueueSize; i++ {
		manager.queue <- i
	}

	task := Task{
		ID:     100,
		Status: Pending,
	}
	manager.tasks[task.ID] = task

	err := manager.SubmitTask(task.ID)
	if err == nil {
		t.Fatal("expected error when queue is full")
	}
	got, err := manager.GetTask(task.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if got.Status != Pending {
		t.Errorf("expected status Pending after failed submission, got %s", got.Status)
	}
}

func TestConcurrentAccess(t *testing.T) {
	manager := NewTaskManager()
	defer manager.Shutdown()

	const goroutines = 20
	var wg sync.WaitGroup
	ids := make(chan int, goroutines)

	input := []string{
		`{"priority":"3","message":"concurrent test","unit":"test.service","source":"test"}`,
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			task, err := manager.CreateTask(input)
			if err != nil {
				t.Errorf("CreateTask failed: %v", err)
				return
			}
			ids <- task.ID
			_, err = manager.GetTask(task.ID)
			if err != nil {
				t.Errorf("GetTask failed: %v", err)
			}
		}()
	}
	wg.Wait()
	close(ids)

	uniqueIDs := make(map[int]bool)
	for id := range ids {
		if uniqueIDs[id] {
			t.Errorf("duplicate task ID: %d", id)
		}
		uniqueIDs[id] = true
	}
	if len(uniqueIDs) != goroutines {
		t.Errorf("expected %d unique task IDs, got %d", goroutines, len(uniqueIDs))
	}
}

func TestShutdownDrainsQueue(t *testing.T) {
	manager := NewTaskManager()

	input := []string{
		`{"priority":"3","message":"test error","unit":"test.service","source":"test"}`,
	}

	const taskCount = 5
	var taskIDs []int

	for i := 0; i < taskCount; i++ {
		task, err := manager.CreateTask(input)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
		if err := manager.SubmitTask(task.ID); err != nil {
			t.Fatalf("failed to submit task %d: %v", task.ID, err)
		}
		taskIDs = append(taskIDs, task.ID)
	}
	manager.Shutdown()

	for _, id := range taskIDs {
		task, err := manager.GetTask(id)
		if err != nil {
			t.Fatalf("failed to get task %d: %v", id, err)
		}
		if task.Status != Completed {
			t.Errorf("task %d: expected Completed, got %s", id, task.Status)
		}
	}
}

func TestCreateSubmitAfterShutdown(t *testing.T) {
	manager := NewTaskManager()

	task, err := manager.CreateTask([]string{
		`{"priority":"3","message":"test error","unit":"test.service","source":"test"}`,
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	taskID := task.ID
	manager.Shutdown()

	_, err = manager.CreateTask([]string{
		`{"priority":"6","message":"test error","unit":"test.service","source":"test"}`,
	})
	if err == nil {
		t.Fatal("expected error when creating task after shutdown")
	}

	err = manager.SubmitTask(taskID)
	if err == nil {
		t.Fatal("expected error when submitting task after shutdown")
	}
}

func TestShutdownTwice(t *testing.T) {
	manager := NewTaskManager()
	manager.Shutdown()
	manager.Shutdown()
}

func TestConcurrentShutdown(t *testing.T) {
	manager := NewTaskManager()
	goroutines := 20
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.Shutdown()
		}()
	}
	wg.Wait()
}
