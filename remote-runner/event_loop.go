package main

import (
	"remote-runner/config"
	"remote-runner/logger"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

func GetConfigWithDefaultInt(key string, defaultValue int) int {
	configValue := config.GetConfigs().GetConfigWithDefault(key, strconv.Itoa(defaultValue))
	configValueInt, err := strconv.Atoi(configValue)
	if err != nil {
		logger.Error("Error converting config value to int", "GET_CONFIG_WITH_DEFAULT_INT", err)
		return defaultValue
	}
	return configValueInt
}

func ProcessEventLoop(exitEventLoop chan bool, waitEventLoopExitChan chan bool, lm *LiveMemory) {
	logger.Info("Starting process event loop", "PROCESS_EVENT_LOOP")
	lastHeartbeatTime := time.Now().Unix()
	// Get the config values
	allowParallelProcesses := GetConfigWithDefaultInt("harpy.remoteRunner.allowParallelProcesses", 4)
	processPoolingInterval := time.Duration(GetConfigWithDefaultInt("harpy.remoteRunner.processPoolingIntervalMs", 200)) * time.Millisecond
	timeoutAfterGettingResult := time.Duration(GetConfigWithDefaultInt("harpy.remoteRunner.timeoutAfterGettingResultSec", 30)) * time.Second
	timeoutProcessNotTriggered := time.Duration(GetConfigWithDefaultInt("harpy.remoteRunner.timeoutProcessNotTriggeredSec", 120)) * time.Second
	heartbeatInterval := time.Duration(GetConfigWithDefaultInt("harpy.remoteRunner.heartbeatIntervalSec", 10)) * time.Second

	// This function will loop through all the processes and check if they are done
	// If they are done then it will set the process status to done
	for {
		select {
		case <-exitEventLoop:
			// Cleanup code here
			logger.Info("Exiting event loop", "PROCESS_EVENT_LOOP")
			waitEventLoopExitChan <- true
			return
		default:
			// Current running processes
			currentTime := time.Now().Unix() // Get the current time
			currentlyRunningProcesses := 0
			for _, process := range lm.Process {
				if process.ProcessStatus == PROCESS_RUNNING {
					currentlyRunningProcesses++
				}
			}
			for key, process := range lm.Process {
				// For now we are going to set any process as done for testing
				if process.ProcessStatus == PROCESS_RUNNING {
					// We start by creating the binaries for the process
					if process.DoneChan == nil {
						process.Run()
					}

					// We will now check if the process is done
					select {
					case <-process.DoneChan:
						// The process is done
						process.ProcessStatus = PROCESS_DONE
						ReportToCallbackClient(lm, key, PROCESS_DONE)
						logger.Info("Process done", "PROCESS_EVENT_LOOP", logrus.Fields{"process_id": key})
					default:
						// The process is still running
					}
				}
				// Check if we can start a new process
				if currentlyRunningProcesses < allowParallelProcesses {
					if process.ProcessStatus == PROCESS_QUEUED {
						process.ProcessStatus = PROCESS_RUNNING
						lm.Process[key] = process
						currentlyRunningProcesses++ // Increment the currently running processes
						ReportToCallbackClient(lm, key, PROCESS_RUNNING)
						logger.Info("Starting process", "PROCESS_EVENT_LOOP", logrus.Fields{"process_id": key})
					}
				}
				// Cleanup the process if it has been fetched
				if process.ResultFetchTime != 0 {
					if currentTime-process.ResultFetchTime > int64(timeoutAfterGettingResult.Seconds()) {
						proc := lm.Process[key]
						proc.Cleanup()
						delete(lm.Process, key)
						logger.Info("Cleaning up process after fetching result", "PROCESS_EVENT_LOOP", logrus.Fields{"process_id": key})
					}
				}
				// Cleanup the process if it has not been triggered for a while
				if (process.ProcessRegisterTime != 0) && (process.ProcessStatus == PROCESS_DEFINED) {
					if currentTime-process.ProcessRegisterTime > int64(timeoutProcessNotTriggered.Seconds()) {
						delete(lm.Process, key)
						logger.Warn("Process not triggered for a while, cleaning up", "PROCESS_EVENT_LOOP", logrus.Fields{"process_id": key})
					}
				}
			}
			// Check if we need to send a heartbeat
			if currentTime-lastHeartbeatTime > int64(heartbeatInterval.Seconds()) {
				lastHeartbeatTime = currentTime
				lm.NodeStatusUpdateClient.SendNodeHeartbeat()
			}
			time.Sleep(processPoolingInterval)
		}
	}
}
