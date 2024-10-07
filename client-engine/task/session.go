// Package task implements any task related operations
//
// This file contains the implementation of the Session structure
// The session is the main structure that holds all the task sets
// This structure is the entry point for everything related to task execution
// Sessions hold the RRT which is used to manage the nodes that are allocated to a session
// Sessions are also a requirement for any task-set to be executed as task-sets require nodes to run
// Sessions do not get discarted when a task-set is completed, they are kept alive until the session is closed
// A user is supposed to use a single session for all the task-sets that are related
// By design, a session will keep the nodes attached to it until the session is closed
//
// Author: Caio Cominato

package task

import (
	"client-engine/logger"
	"fmt"

	"github.com/google/uuid"
)

type TaskSetListener interface {
	OnTaskSetProgress(taskSet *TaskSet)
	OnTaskGroupProgress(taskSet *TaskSet, taskGroup *TaskGroup)
	OnTaskProgress(taskSet *TaskSet, task *TaskRun)
	//OnResourceManagerChange(resourceManager ResourceManager)
}

type Session struct {
	SessionId                        string
	TaskSets                         map[string]*TaskSet
	TaskSetListeners                 map[string]TaskSetListener
	Options                          map[string]string
	NodeTracker                      *NodeTracker
	RegisterCallbackPointer          func(callback func(string, Status) error) string
	RegisterCallbackID               func(string, string)
	DeregisterCommandCallbackPointer func(string)
}

func GetFromMappingWithDefaultValue(mapping map[string]string, key string, defaultValue string) string {
	value, exists := mapping[key]
	if !exists {
		return defaultValue
	}
	return value
}

func NewSession(
	RegisterCommandCallbackPointer func(callback func(string, Status) error) string,
	RegisterCommandID func(string, string),
	DeregisterCommandCallbackPointer func(string),
	options map[string]string,
) (*Session, error) {
	idx := fmt.Sprintf("session-%s", uuid.New().String())
	// Construct the node tracker with the callback URI
	callbackPort := GetFromMappingWithDefaultValue(options, "harpy.clientEngine.callback.port", "50052")
	callbackHost := GetFromMappingWithDefaultValue(options, "harpy.clientEngine.callback.host", "localhost")
	callbackURI := fmt.Sprintf("%s:%s", callbackHost, callbackPort)
	nt, err := NewNodeTracker(callbackURI)
	if err != nil {
		// TODO: HANDLE THIS ERROR
		logger.Error("Error creating node tracker", "SESSION", err)
		return nil, err
	}
	return &Session{
		SessionId:                        idx,
		TaskSets:                         make(map[string]*TaskSet),
		TaskSetListeners:                 make(map[string]TaskSetListener),
		Options:                          options,
		NodeTracker:                      nt,
		RegisterCallbackPointer:          RegisterCommandCallbackPointer,
		RegisterCallbackID:               RegisterCommandID,
		DeregisterCommandCallbackPointer: DeregisterCommandCallbackPointer,
	}, nil
}

func (s *Session) String() string {
	return fmt.Sprintf("Session[%s]", s.SessionId)
}

func (s *Session) AddTaskSetListener(listener TaskSetListener) string {
	listenerUUID := uuid.New().String()
	s.TaskSetListeners[listenerUUID] = listener
	return listenerUUID
}

func (s *Session) RemoveTaskSetListener(listenerUUID string) {
	delete(s.TaskSetListeners, listenerUUID)
}

func (s *Session) CreateTaskSet() *TaskSet {
	ts := NewTaskSet(s)
	// Add the task set to the session
	s.TaskSets[ts.TaskSetId] = ts
	return ts
}

func (s *Session) GetTaskSet(idx string) *TaskSet {
	taskSet, exists := s.TaskSets[idx]
	if !exists {
		panic(fmt.Sprintf("TaskSet with id %s does not exist", idx))
	}
	return taskSet
}

func (s *Session) Close() {
	// To close the session we will simply delete all the tasks from the session tracker
	//  Hopefully the garbage collector will take care of the rest
	s.TaskSets = nil
	s.NodeTracker.Close()
	s.NodeTracker = nil
	// Ensure there is no reference to ResourceManager
}
