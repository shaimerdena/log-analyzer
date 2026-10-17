# log-analyzer

## MVP (Checkpoint 1)

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