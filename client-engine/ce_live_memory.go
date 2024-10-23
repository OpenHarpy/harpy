package main

import (
	"client-engine/config"
	"client-engine/logger"
	"client-engine/task"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Begin LiveMemory
type LiveMemory struct {
	// The Live Memory is a struct that holds all the live data of the server
	// It is used to store all the live data of the server
	InstanceID             string
	Sessions               map[string]*task.Session
	TaskSetDefinitions     map[string]*task.TaskSet
	TaskSetSession         map[string]string // This is a map of taskSetID and sessionID
	TaskDefinitions        map[string]*task.Definition
	CallbackPointers       map[string]func(string, task.Status) error
	CommandCallbackPointer map[string]string
	BlocksMapping          map[string]*string // This is a map of blockID and sessionID
}

func NewLiveMemory() *LiveMemory {
	// This function initializes a new LiveMemory object
	instanceID := config.GetConfigs().InstanceID
	return &LiveMemory{
		InstanceID:             instanceID,
		Sessions:               make(map[string]*task.Session),
		TaskSetDefinitions:     make(map[string]*task.TaskSet),
		TaskSetSession:         make(map[string]string),
		TaskDefinitions:        make(map[string]*task.Definition),
		CallbackPointers:       make(map[string]func(string, task.Status) error),
		CommandCallbackPointer: make(map[string]string),
		BlocksMapping:          make(map[string]*string),
	}
}

func (lm *LiveMemory) RegisterCommandCallbackPointer(callback func(string, task.Status) error) string {
	logger.Info("Registering callback pointer", "LIVE_MEMORY", logrus.Fields{"callback": callback})
	callbackUUID := uuid.New().String()
	lm.CallbackPointers[callbackUUID] = callback
	return callbackUUID
}
func (lm *LiveMemory) RegisterCommandID(commandID string, callbackPointerID string) {
	logger.Info("Registering callback pointer", "LIVE_MEMORY", logrus.Fields{"commandID": commandID, "callbackPointerID": callbackPointerID})
	lm.CommandCallbackPointer[commandID] = callbackPointerID
}
func (lm *LiveMemory) DeregisterCommandCallbackPointer(callbackPointerID string) {
	logger.Info("Deregistering callback pointer", "LIVE_MEMORY", logrus.Fields{"callbackPointerID": callbackPointerID})
	// Remove any references to the callback pointer
	for k, v := range lm.CommandCallbackPointer {
		if v == callbackPointerID {
			delete(lm.CommandCallbackPointer, k)
		}
	}
	delete(lm.CallbackPointers, callbackPointerID)
}
