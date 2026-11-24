# log-analyzer

## Checkpoint 1 - MVP

Log Analyzer is a lightweight, in-memory log analysis tool written in Go. It processes plain-text log entries, validates their format, and generates a structured summary to help identify errors and understand application events.

### Core Features

- **Log Parsing:** Parse log entries and extract the severity level, component, and message.
- **Input Validation:** Reject empty input and malformed log entries.
- **Log Analysis:** Calculate the total number of entries, count entries by severity level (`INFO`, `WARN`, `ERROR`), and identify components with the most errors.
- **Task Management:** Create tasks with unique IDs, store them in memory, and retrieve them by ID.
- **Task Lifecycle:** Track task states from `Pending` to `Processing` and `Completed`.
- **Error Handling:** Handle invalid input, missing tasks, and invalid state transitions using Go's error-handling mechanisms.
- **Unit Testing:** Test core functionality, error scenarios, analysis results, and independent task processing using Go's standard `testing` package.

### Task Lifecycle

Each task follows a simple lifecycle:

1. **Pending** — The task has been created, validated, and stored in memory. It is waiting to be processed.
2. **Processing** — The analyzer processes the task's log entries and generates an analysis result.
3. **Completed** — Processing has finished successfully, and the analysis result is stored in the task.

A task can only be processed while it is in the `Pending` state. Attempts to process a task again or perform invalid operations return an error.

### Scope

The MVP focuses on synchronous task processing and in-memory storage. It does not require a database, external APIs, or additional infrastructure.

---

## Checkpoint 2 — Concurrent Task Processing

Checkpoint 2 extends the MVP by introducing asynchronous task processing through a worker pool. Instead of processing each task immediately and synchronously, tasks can be submitted to a queue and processed concurrently by multiple workers.

### Core Features

- **Worker Pool:** Implement a fixed-size worker pool consisting of three workers to process tasks concurrently.
- **Task Queue:** Introduce a buffered channel with a capacity of five tasks to manage pending submissions.
- **Asynchronous Submission:** Submit tasks for processing without waiting for the analysis to finish.
- **Task Lifecycle:** Extend the lifecycle with `Queued` and `Failed` states to track asynchronous execution and processing failures.
- **Queue Management:** Reject task submissions when the queue is full and restore the task to its `Pending` state.
- **Concurrency Safety:** Protect shared task data and task ID generation using synchronization mechanisms to prevent data races.
- **Task Isolation:** Ensure that tasks are processed independently and that one task's processing does not interfere with other tasks.
- **Graceful Shutdown:** Stop accepting new tasks, process already accepted tasks, and wait for workers to finish before shutting down.
- **Backward Compatibility:** Preserve synchronous task processing functionality from Checkpoint 1.
- **Concurrent Testing:** Test asynchronous processing, queue capacity, duplicate submissions, concurrent task access, and shutdown behavior using Go's standard `testing` package and race detector.

### Task Lifecycle

Tasks now follow an extended lifecycle:

1. **Pending** — The task has been created, validated, and stored in memory.
2. **Queued** — The task has been successfully submitted to the task queue and is waiting for a worker.
3. **Processing** — A worker has picked up the task and started analyzing its log entries.
4. **Completed** — Processing has finished successfully, and the analysis result is stored.
5. **Failed** — Processing has failed, and the error is recorded in the task.

Invalid state transitions and duplicate task submissions are rejected.

### Worker Pool Architecture

The TaskManager manages task storage, task submission, the task queue, and worker lifecycle.

- The task queue uses a buffered channel with a capacity of five.
- Three workers consume task IDs from the queue.
- Workers process tasks independently and update their status and analysis results.
- Shared task data is protected by mutexes to ensure concurrency safety.
- The queue uses a non-blocking submission mechanism to reject tasks when capacity is exhausted.

### Shutdown Behavior

The TaskManager supports graceful shutdown:

- New task creation and submissions are rejected after shutdown begins.
- The task queue is closed.
- Workers finish processing tasks that were already accepted into the queue.
- Shutdown waits for all workers to exit before returning.
- Repeated shutdown calls are handled safely.

### Scope

Checkpoint 2 focuses on concurrent task processing, in-memory task management, synchronization, queue management, and worker lifecycle. The project remains self-contained and does not require a database, external APIs, HTTP endpoints, or additional infrastructure.
