// Package task implements any task related operations
//
// This file contains the implementation of the task run
// The task run is the structure that holds the information about a task that is being executed
// A "TaskDefinition" is generated by a "TaskFactory" and a task yields a "TaskRun" that contains a "Result"
//
// Author: Caio Cominato

package task

import "fmt"

type Status string // This is a string because for now we are not going to map the statuses to a specific grpc enum

const (
	STATUS_UNDEFINED Status = "undefined"
	STATUS_DEFINED   Status = "defined"
	STATUS_QUEUED    Status = "queued"
	STATUS_RUNNING   Status = "running"
	STATUS_FETCHING  Status = "fetching"
	STATUS_DONE      Status = "done"
	STATUS_PANIC     Status = "panic"
	STATUS_KILLING   Status = "killing"
)

// Result
type Result struct {
	ObjectReturnBlockID BlockID
	StdoutBlockID       BlockID
	StderrBlockID       BlockID
	TaskRunID           string
	Success             bool
}

func (r Result) String() string {
	return fmt.Sprintf("Result[%s]; success=[%t]", r.TaskRunID, r.Success)
}

func NewResult(taskRunID string, success bool, objectReturnBlockID BlockID, stdoutBlockID BlockID, stderrBlockID BlockID) *Result {
	return &Result{
		TaskRunID:           taskRunID,
		Success:             success,
		ObjectReturnBlockID: objectReturnBlockID,
		StdoutBlockID:       stdoutBlockID,
		StderrBlockID:       stderrBlockID,
	}
}

// TaskRun

type TaskRun struct {
	Task      *TaskDefinition
	Reporter  *Reporter
	Result    *Result
	TaskRunID string
	Status    Status
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
func (t *TaskRun) SetStatus(status Status) {
	t.Status = status
	t.Report()
}
func (t *TaskRun) SetResult(resultBlockID BlockID, stdoutBlockID BlockID, stderrBlockID BlockID, success bool) {
	t.Result = NewResult(
		t.TaskRunID,
		success,
		resultBlockID,
		stdoutBlockID,
		stderrBlockID,
	)
}
