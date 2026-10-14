package task

import (
	"testing"
)

func TestCreateTask(t *testing.T){
	manager := NewTaskManager()

	input := []string{
		`{
		"PRIORITY": "6",
		"MESSAGE": "Starting demo.service - Demo service.",
		"_SYSTEMD_UNIT": "user-demo.service",
		"SYSLOG_IDENTIFIER": "demo"
		}`,
		`{
		"PRIORITY": "3",
		"MESSAGE": "Demo service failed",
		"_SYSTEMD_UNIT": "user-demo.service",
		"SYSLOG_IDENTIFIER": "demo"
		}`,
	}

 	createdTask, err := manager.CreateTask(input)
	if err != nil{
		t.Fatalf("expected no error, got %v", err)
	}
	if createdTask.ID != 0 {
		t.Error("expected id to be 0")
	}
	if createdTask.Status != Pending{
		t.Error("expected status to be pending")
	}

	storedTask, ok := manager.tasks[createdTask.ID]
	if ok == false{
		t.Error("expected task to be in map")
	}
	if storedTask.ID != createdTask.ID {
		t.Error("expected IDs to be the same")
	}
}

func TestEmptiness(t *testing.T){
	manager := NewTaskManager()

	input := []string{}

	_, err := manager.CreateTask(input)

	if err == nil {
		t.Fatal("expected error for empty input logs")
	}
}

func TestJsonValidation(t *testing.T){
	manager := NewTaskManager()

	input := []string{`{"PRIORITY":`}

	_, err := manager.CreateTask(input)

	if err == nil{
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetTask(t *testing.T){
	manager := NewTaskManager()

	input := []string{
		`{
		"PRIORITY": "6",
		"MESSAGE": "Starting demo.service - Demo service.",
		"_SYSTEMD_UNIT": "user-demo.service",
		"SYSLOG_IDENTIFIER": "demo"
		}`,
	}

	createdTask, err := manager.CreateTask(input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	task, err := manager.GetTask(createdTask.ID)

	if err != nil {
		t.Fatalf("expected to get task by ID: %v", err)
	}
	if task.ID != createdTask.ID {
		t.Error("expected to get the same id")
	}
}

func TestGetTaskNotFound(t *testing.T){
	manager := NewTaskManager()

	input := []string{
		`{
		"PRIORITY": "6",
		"MESSAGE": "Starting demo.service - Demo service.",
		"_SYSTEMD_UNIT": "user-demo.service",
		"SYSLOG_IDENTIFIER": "demo"
		}`,
	}

	_, err := manager.CreateTask(input)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	_, err = manager.GetTask(99)
	if err == nil {
		t.Fatal("expected to get error")
	}
}