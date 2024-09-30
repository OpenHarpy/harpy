package task

import (
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
	SessionId        string
	TaskSets         map[string]*TaskSet
	TaskSetListeners map[string]TaskSetListener
	Options          map[string]string
	NodeTracker      *NodeTracker
}

func NewSession(options map[string]string) (*Session, error) {
	idx := fmt.Sprintf("session-%s", uuid.New().String())
	nt := NewNodeTracker()
	return &Session{
		SessionId:        idx,
		TaskSets:         make(map[string]*TaskSet),
		TaskSetListeners: make(map[string]TaskSetListener),
		Options:          options,
		NodeTracker:      nt,
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
