package task

import (
	"testing"
)

func TestProcessTask(t *testing.T) {
	manager := NewTaskManager()

	input := []string{
		`{
            "PRIORITY": "3",
            "MESSAGE": "nginx worker failed",
            "_SYSTEMD_UNIT": "nginx.service",
            "SYSLOG_IDENTIFIER": "nginx"
        }`,
	}

	createdTask, err := manager.CreateTask(input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	err = manager.ProcessTask(createdTask.ID)
	if err != nil {
		t.Fatalf("failed to process task: %v", err)
	}

	processedTask, err := manager.GetTask(createdTask.ID)
	if err != nil {
		t.Fatalf("failed to get processed task: %v", err)
	}

	if processedTask.Status != Completed {
		t.Fatal("failed to complete task")
	}
	if processedTask.Result.TotalEntries != 1 {
		t.Errorf("expected 1 total entry, got %d", processedTask.Result.TotalEntries)
	}
	if processedTask.Result.ErrorCount != 1 {
		t.Errorf("expected 1 error count, got %d", processedTask.Result.ErrorCount)
	}
	if processedTask.Result.ErrorsByUnit["nginx.service"] != 1 {
		t.Errorf("expected 1 nginx.service error, got %d", processedTask.Result.ErrorsByUnit["nginx.service"])
	}
}

func TestProcessTaskTwice(t *testing.T) {
	manager := NewTaskManager()

	input := []string{
		`{
            "PRIORITY": "3",
            "MESSAGE": "nginx worker failed",
            "_SYSTEMD_UNIT": "nginx.service",
            "SYSLOG_IDENTIFIER": "nginx"
        }`,
	}

	createdTask, err := manager.CreateTask(input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	err = manager.ProcessTask(createdTask.ID)
	if err != nil {
		t.Fatalf("failed to process task: %v", err)
	}

	err = manager.ProcessTask(createdTask.ID)
	if err == nil {
		t.Fatal("expected error when processing a completed task")
	}
}

func TestProcessTaskNotFound(t *testing.T) {
	manager := NewTaskManager()

	err := manager.ProcessTask(99)
	if err == nil {
		t.Fatal("expected an error when processing a nonexistent task")
	}
}

func TestIndependentTasks(t *testing.T) {
	manager := NewTaskManager()

	input1 := []string{
		`{
		"PRIORITY": "6",
		"MESSAGE": "Starting demo.service - Demo service.",
		"_SYSTEMD_UNIT": "user-demo.service",
		"SYSLOG_IDENTIFIER": "demo"
		}`,
	}

	input2 := []string{`{
		"PRIORITY": "3",
		"MESSAGE": "Demo service failed",
		"_SYSTEMD_UNIT": "user-demo.service",
		"SYSLOG_IDENTIFIER": "demo"
		}`,
	}

	createdTask1, err := manager.CreateTask(input1)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	createdTask2, err := manager.CreateTask(input2)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	err = manager.ProcessTask(createdTask1.ID)
	if err != nil {
		t.Fatalf("failed to process task: %v", err)
	}

	createdTask2, err = manager.GetTask(createdTask2.ID)
	if err != nil {
		t.Fatalf("failed to get created task: %v", err)
	}

	processedTask1, err := manager.GetTask(createdTask1.ID)
	if err != nil {
		t.Fatalf("failed to get processed task: %v", err)
	}

	if processedTask1.Status != Completed {
		t.Error("expected to get status complete")
	}
	if processedTask1.Result.ErrorCount != 0 {
		t.Errorf("expected to get 0 errors, got %d", processedTask1.Result.ErrorCount)
	}

	if createdTask2.Status != Pending {
		t.Error("expected to get status pending")
	}
	if createdTask2.Result.ErrorCount != 0 {
		t.Errorf("expected to get 0 errors, got %d", createdTask2.Result.ErrorCount)
	}
}
