// Package task implements any task related operations
//
// This file contains the implementation of the TaskSet structure
// The task-set is a collection of task-groups that are related
// Task-sets control the definition and execution of the task-groups that are executed
// Task-sets makes sure that the task-definitions are defined in the correct order
// Task-sets also control the order of execution of the task-groups
// They also hold the results of the task-groups that are executed and the overall status of the task-set
//
// Author: Caio Cominato

package task

import (
	"client-engine/logger"
	"fmt"

	"github.com/google/uuid"
)

type TaskSetResult struct {
	// This did not exist in the original Python implementation of the task set
	//
	//	Ultimately its best to separate out task set results and task group results
	//	This will allow for more flexibility in the future
	//  We still want to allow the user to get the full result object and not just the output object
	//  This needs to be considered if we want to allow the user to get the full result object and not just the output object
	TaskSetID     string
	Results       []Result
	OverallStatus string
}

type TaskSet struct {
	TaskSetId         string
	TaskSetStatus     string
	TaskSetProgress   string
	RootNode          *TaskGroup
	TaskSetReporter   *Reporter
	TaskReporter      *Reporter
	TaskGroupReporter *Reporter
	//ResourceReporter *Reporter
	Session            *Session
	TaskSetResultCache *TaskSetResult // This is used to cache the result of the task set
}

func (t TaskSet) generateDefaultTaskGroup(factory TaskFactory, options map[string]string) TaskGroup {
	// This will generate the default options for the task group
	//   This will allow the user to override the default options
	idx := fmt.Sprintf("tg-%s", uuid.New().String())
	name := fmt.Sprintf("TaskGroup[%s]", idx)
	tskGrp := *NewTaskGroup(
		idx,
		factory,
		options,
		name,
		t.TaskGroupReporter,
		t.TaskReporter,
	)
	return tskGrp
}

func (t *TaskSet) AddNewTaskGroup(taskGroup TaskGroup) error {
	// Traverse the task set to find the last node
	//   This will allow us to add the new task group to the end of the task set
	if t.RootNode == nil {
		t.RootNode = &taskGroup
		return nil
	}
	currentNode := t.RootNode
	for {
		if currentNode.GetNextNode() == nil {
			break
		}
		currentNode = currentNode.GetNextNode()
	}
	e := currentNode.InsertNextNode(&taskGroup)
	return e
}

func (t *TaskSet) Transform(transformerDef TransformerDefinition, options map[string]string) error {
	// Transform cannot be the root node
	if t.RootNode == nil {
		return fmt.Errorf("RootNode is nil - cannot add transformer as root node")
	}
	// This will create a new task set with the transformer as the root node
	transformerFactory := TransformFactory{Transformer: transformerDef}
	taskGroup := t.generateDefaultTaskGroup(transformerFactory, options)
	e := t.AddNewTaskGroup(taskGroup)
	return e
}

func (t *TaskSet) Map(mappers []MapperDefinition, options map[string]string) error {
	// Map can only be the root node
	if t.RootNode != nil {
		return fmt.Errorf("RootNode is not nil - mappers can only be the root node")
	}
	// This will create a new task set with the mappers as the root node
	mapFactory := MapFactory{Mappers: mappers}
	taskGroup := t.generateDefaultTaskGroup(mapFactory, options)
	e := t.AddNewTaskGroup(taskGroup)
	return e
}

func (t *TaskSet) Reduce(reducerDef ReducerDefinition, options map[string]string) error {
	// Reduce cannot be the root node
	if t.RootNode == nil {
		return fmt.Errorf("RootNode is nil - cannot add transformer as root node")
	}
	// This will create a new task set with the reducer as the root node
	reduceFactory := ReduceFactory{Reducer: reducerDef}
	taskGroup := t.generateDefaultTaskGroup(reduceFactory, options)
	e := t.AddNewTaskGroup(taskGroup)
	return e
}

func (t *TaskSet) Report() {
	if t.TaskSetReporter != nil {
		t.TaskSetReporter.Report(t)
	}
}

func (t *TaskSet) Execute() (TaskSetResult, error) {
	// Taskset needs to have a root node
	if t.RootNode == nil {
		return TaskSetResult{}, fmt.Errorf("RootNode is nil")
	}
	// Logger for the task set
	logger.Info("Executing TaskSet[%s]", t.TaskSetId)
	// Execute the task set
	t.TaskSetProgress = "running"
	t.TaskSetProgress = "running"
	t.Report()
	nextNode := t.RootNode
	lastResult := TaskGroupResult{}
	failing := false
	for nextNode != nil {
		if failing {
			nextNode.SkipRemaining()
		} else {
			// We will execute the task group
			//result, err := nextNode.Execute(lastResult, t.Session)
			result, err := nextNode.RemoteGRPCExecute(lastResult, t.Session)
			if err != nil {
				errorString := fmt.Sprintf("Error executing TaskGroup[%s]", nextNode.TaskGroupID)
				logger.Error(errorString, "TASKSET", err)
				return TaskSetResult{}, err
			}
			lastResult = result
			if lastResult.OverallStatus != "success" {
				failing = true
				nextNode.SkipRemaining()
			}
		}
		nextNode = nextNode.NextNode
	}

	if failing {
		t.TaskSetStatus = "failed"
		t.TaskSetProgress = "compleated"
	} else {
		t.TaskSetStatus = "success"
		t.TaskSetProgress = "compleated"
	}
	t.TaskSetResultCache = &TaskSetResult{
		TaskSetID:     t.TaskSetId,
		Results:       lastResult.Results,
		OverallStatus: lastResult.OverallStatus,
	}
	t.Report()
	// Return the result
	return *t.TaskSetResultCache, nil
}

func (t TaskSet) GetTaskSetResult() TaskSetResult {
	if t.TaskSetProgress != "compleated" {
		return TaskSetResult{OverallStatus: "pending"}
	} else {
		return *t.TaskSetResultCache
	}
}

func OnTaskProgress(Task *TaskRun, Sess *Session, TaskSetIdx string) {
	// For each TaskSetListener in the session, call the OnTaskProgress method
	taskSet := Sess.GetTaskSet(TaskSetIdx)
	for _, listener := range Sess.TaskSetListeners {
		listener.OnTaskProgress(taskSet, Task)
	}
}
func OnTaskGroupProgress(TaskGroup *TaskGroup, Sess *Session, TaskSetIdx string) {
	// For each TaskSetListener in the session, call the OnTaskGroupProgress method
	taskSet := Sess.GetTaskSet(TaskSetIdx)
	for _, listener := range Sess.TaskSetListeners {
		listener.OnTaskGroupProgress(taskSet, TaskGroup)
	}
}
func OnTaskSetProgress(TaskSet *TaskSet, Sess *Session) {
	// For each TaskSetListener in the session, call the OnTaskSetProgress method
	for _, listener := range Sess.TaskSetListeners {
		listener.OnTaskSetProgress(TaskSet)
	}
}

func NewTaskSet(session *Session) *TaskSet {
	idx := fmt.Sprintf("ts-%s", uuid.New().String())
	context := []interface{}{session, idx}

	taskSetReporter := NewReporter(func(args ...interface{}) {
		session := args[1].(*Session)
		OnTaskSetProgress(args[0].(*TaskSet), session)
	}, context)

	taskGroupReporter := NewReporter(func(args ...interface{}) {
		session := args[1].(*Session)
		ts_idx := args[2].(string)
		OnTaskGroupProgress(args[0].(*TaskGroup), session, ts_idx)
	}, context)

	taskReporter := NewReporter(func(args ...interface{}) {
		session := args[1].(*Session)
		ts_idx := args[2].(string)
		OnTaskProgress(args[0].(*TaskRun), session, ts_idx)
	}, context)

	return &TaskSet{
		TaskSetId:         idx,
		TaskSetReporter:   taskSetReporter,
		TaskGroupReporter: taskGroupReporter,
		TaskReporter:      taskReporter,
		Session:           session,
	}
}

func (t TaskSet) String() string {
	return fmt.Sprintf("TaskSet[%s]", t.TaskSetId)
}
