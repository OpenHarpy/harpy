// Package task implements any task related operations
//
// This file contains the implementation of the TaskGroup structure
// The task-group is a collection of tasks that are related
// Task-groups control the execution and collect the results of the tasks that are executed
// Task-groups depend factories to generate the tasks that are going to be executed
//
// Author: Caio Cominato

package task

import (
	"client-engine/logger"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	POOLING_INTERVAL = 200 * time.Millisecond
)

// TaskGroupResult
type TaskGroupResult struct {
	TaskGroupID   string
	Results       []Result
	OverallStatus string
}

func (t TaskGroupResult) String() string {
	return fmt.Sprintf("TaskGroupResult[%s]; overall_status=[%s]", t.TaskGroupID, t.OverallStatus)
}
func (t TaskGroupResult) GetErrors() []Result {
	var errors []Result
	for _, result := range t.Results {
		if !result.Success {
			errors = append(errors, result)
		}
	}
	return errors
}
func (t TaskGroupResult) GetSuccesses() []Result {
	var successes []Result
	for _, result := range t.Results {
		if result.Success {
			successes = append(successes, result)
		}
	}
	return successes
}
func (t TaskGroupResult) GetAll() []Result {
	return t.Results
}

// TaskFactory
type TaskFactory interface {
	MakeTasks(previousResult TaskGroupResult) []TaskDefinition
	String() string
}

// TaskGroup
type TaskGroup struct {
	TaskGroupID           string
	TaskGenerator         TaskFactory
	Options               map[string]string
	NodeMappingCallbackID map[string]string
	CommandIDNodeMapping  map[string]string // This is a map between nodeID -> TaskRunID to commandID
	CommandIDTaskMapping  map[string]*TaskRun
	Name                  string
	TaskGroupReporter     *Reporter
	TaskReporter          *Reporter
	Tasks                 []TaskDefinition
	TaskRuns              []*TaskRun
	TaskSet               *TaskSet
	NextNode              *TaskGroup
	MonolithicIndex       int
}

func (t TaskGroup) String() string {
	return fmt.Sprintf("TaskGroup[%s]", t.TaskGroupID)
}

func (t *TaskGroup) Report() {
	if t.TaskGroupReporter != nil {
		t.TaskGroupReporter.Report(t)
	}
}
func (t *TaskGroup) InsertNextNode(nextNode *TaskGroup) error {
	if t.NextNode != nil {
		return fmt.Errorf("NextNode is not nil for TaskGroup[%s]", t.TaskGroupID)
	}
	t.NextNode = nextNode
	return nil
}
func (t *TaskGroup) GetNextNode() *TaskGroup {
	return t.NextNode
}
func (t *TaskGroup) generateTasks(taskResults TaskGroupResult) error {
	t.Tasks = t.TaskGenerator.MakeTasks(taskResults)
	if t.Tasks == nil {
		return fmt.Errorf("tasks have not been generated for TaskGroup[%s]", t.TaskGroupID)
	}
	// For each task, create a TaskRun
	for _, task := range t.Tasks {
		TaskRunID := fmt.Sprintf("%s-tr-%d", t.TaskGroupID, t.MonolithicIndex)
		t.TaskRuns = append(t.TaskRuns, NewTaskRun(&task, TaskRunID, t.TaskReporter, t.TaskSet))
		t.MonolithicIndex++ // Increment the monolithic index
	}
	return nil
}
func (t TaskGroup) SkipRemaining() {
	for _, taskRun := range t.TaskRuns {
		if taskRun.Status == "pending" {
			taskRun.Skip()
		}
	}
}

func goFetchResult(commandID string, taskRun *TaskRun, tg *TaskGroup) {
	nodeID := tg.CommandIDNodeMapping[commandID]
	node := tg.TaskSet.Session.NodeTracker.GetNode(nodeID)
	err := node.GetTaskOutput(commandID, taskRun)
	if err != nil {
		logger.Error("Error fetching task output", "TASKGROUP", err)
	}
	taskRun.Fetched = true
}

func (t *TaskGroup) CallbackHandler(commandID string, status Status) error {
	taskRun, ok := t.CommandIDTaskMapping[commandID]
	if !ok {
		println("commandID not found in CommandIDTaskMapping")
		return fmt.Errorf("commandID not found in CommandIDTaskMapping")
	}
	taskRun.SetStatus(status)
	if status == STATUS_DONE {
		go goFetchResult(commandID, taskRun, t)
	}
	return nil
}

func (t TaskGroup) RemoteGRPCExecute(previousResult TaskGroupResult, session *Session) (TaskGroupResult, error) {
	if t.Tasks == nil {
		return TaskGroupResult{}, fmt.Errorf("tasks have not been generated for TaskGroup[%s]", t.TaskGroupID)
	}
	t.Report()
	t.generateTasks(previousResult)
	t.Report()
	var results []Result = []Result{}

	// Initializing the mappings for the tasks
	// This is necessary to track the tasks
	t.CommandIDNodeMapping = make(map[string]string)
	t.CommandIDTaskMapping = make(map[string]*TaskRun)

	// Register the callback handler
	callbackHandlerID := session.RegisterCallbackPointer(t.CallbackHandler)
	logger.Info("CallbackHandler registered", "TASKGROUP", logrus.Fields{"blockGroupID": t.TaskSet.CurrentBlockGroupID, "callbackHandlerID": callbackHandlerID})
	for _, task := range t.TaskRuns {
		// For each task we get the node from the session
		node := session.NodeTracker.GetNextNode()
		if node == nil {
			errorString := fmt.Sprintf("No nodes available for TaskGroup[%s]", t.TaskGroupID)
			err := errors.New(errorString)
			logger.Info(errorString, "TASKGROUP")
			return TaskGroupResult{}, err
		}
		// We need to register the task to the node
		commandID, err := node.RegisterTask(task.Task, t.TaskSet)
		if err != nil {
			return TaskGroupResult{}, err
		}
		// We need to apply the mappings so we can track the task
		t.CommandIDNodeMapping[commandID] = node.nodeID
		t.CommandIDTaskMapping[commandID] = task
		// We need to register the commandID to the callbackHandler
		session.RegisterCallbackID(commandID, callbackHandlerID)
	}
	// Now for each command id lets ask the node to run the command
	// Why is this done in a separate loop?
	//  - Nodes could fail to register the command
	//  - Mainly, we trade off a bit of performance for more reliability in the system
	//  - The loop above could be tuned to do some retries if we start to see common issues
	for commandID, nodeID := range t.CommandIDNodeMapping {
		node := session.NodeTracker.GetNode(nodeID)
		err := node.RunCommand(commandID)
		// Under the hood this will span a thread in the node and this action will be non-blocking
		if err != nil {
			logger.Info("Error running command", "TASKGROUP")
			return TaskGroupResult{}, err
		}
	}
	t.Report()

	for {
		allDone := true // Assume all tasks are done at the beginning of each iteration
		// We need to check if all the tasks are done
		for _, taskRun := range t.TaskRuns {
			if !taskRun.Fetched {
				allDone = false // Set allDone to false if any task is not fetched
				break
			}
		}
		if allDone {
			break // Exit the loop if all tasks are done
		}
		time.Sleep(POOLING_INTERVAL) // The pooling interval can be small because this is not a network call
		// Lets change this to use channels instead of polling
		// TODO: Change this after we test the new implementation of the server callback
	}

	// We need to get the results from the commandIDTaskMapping
	// This may not be safe to do concurrently so we will do it sequentially once all the results are fetched
	for _, taskRun := range t.CommandIDTaskMapping {
		results = append(results, *taskRun.Result)
	}

	session.DeregisterCommandCallbackPointer(callbackHandlerID)

	t.Report()

	overallStatus := "success"
	for _, result := range results {
		if !result.Success {
			overallStatus = "failure"
			break
		}
	}

	return TaskGroupResult{TaskGroupID: t.TaskGroupID, Results: results, OverallStatus: overallStatus}, nil
}

// NewTaskGroup is a constructor function for TaskGroup
func NewTaskGroup(
	taskGroupID string,
	taskGenerator TaskFactory,
	options map[string]string,
	name string,
	taskGroupReporter *Reporter,
	taskReporter *Reporter,
	taskSet *TaskSet,
) *TaskGroup {
	if taskSet == nil {
		panic(errors.New("taskSet cannot be nil"))
	}
	return &TaskGroup{
		TaskGroupID:       taskGroupID,
		TaskGenerator:     taskGenerator,
		Options:           options,
		Name:              name,
		Tasks:             []TaskDefinition{}, // Default empty slice
		TaskRuns:          []*TaskRun{},       // Default empty slice
		NextNode:          nil,                // Default nil
		TaskGroupReporter: taskGroupReporter,
		TaskReporter:      taskReporter,
		MonolithicIndex:   0,
		TaskSet:           taskSet,
	}
}
