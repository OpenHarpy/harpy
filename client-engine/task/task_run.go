package task

import "fmt"

// Result
type Result struct {
	ObjectReturnBinary []byte
	StdoutBinary       []byte
	StderrBinary       []byte
	TaskRunID          string
	Success            bool
}

func (r Result) String() string {
	return fmt.Sprintf("Result[%s]; success=[%t]", r.TaskRunID, r.Success)
}

func NewResult(taskRunID string, success bool, objectReturnBinary []byte, stdoutBinary []byte, stderrBinary []byte) *Result {
	return &Result{
		TaskRunID:          taskRunID,
		Success:            success,
		ObjectReturnBinary: objectReturnBinary,
		StdoutBinary:       stdoutBinary,
		StderrBinary:       stderrBinary,
	}
}

// TaskRun

type TaskRun struct {
	Task      *TaskDefinition
	Reporter  *Reporter
	Result    *Result
	TaskRunID string
	Status    string
}

func NewTaskRun(task *TaskDefinition, taskRunID string, reporter *Reporter) *TaskRun {
	return &TaskRun{
		Task: task, TaskRunID: taskRunID, Reporter: reporter, Status: "pending", Result: nil,
	}
}

func (t *TaskRun) Skip() {
	t.Status = "skipped"
}
func (t TaskRun) String() string {
	return fmt.Sprintf("TaskRun[%s]", t.TaskRunID)
}
func (t *TaskRun) Report() {
	t.Reporter.Report(t)
}
func (t *TaskRun) SetStatus(status string) {
	t.Status = status
	t.Report()
}
func (t *TaskRun) SetResult(binaryObject []byte, stdout []byte, stderr []byte, success bool) {
	t.Result = NewResult(
		t.TaskRunID,
		success,
		binaryObject,
		stdout,
		stderr,
	)
}
