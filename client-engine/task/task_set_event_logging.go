package task

import (
	"encoding/json"
	"fmt"
)

/*
Event logging for tasksets will follow the format of the following json:
{
	"taskset_id": "taskset_id",
	"session_id": "session_id",
	"graph": {
		"task_group_id": {
			"tasks": [
				{ "id": "task_id", "name": "task_name" },
			],
			"runs": [
				{ "id": "run_id", "status": "status", "is_done": "is_done", "success": "success" },
			],
			"options": {"option": "value", "option2": "value2", ...},
			"next_node_id": "next_task_group_id"
		},
		"task_group_id": { ... },
		...
	},
	"task_set_progress": "progress",
	"task_set_status": "status",
	"task_group_root": "task_group_id",
	"options": {"option": "value", "option2": "value2", ...},
}

*/

// To this effect we will need to create a new structure to hold the event logging information
// This will be a structure that will hold the taskset and session information
// This will be used to log the events that are happening in the taskset

// TaskJson is a structure that will hold the information of the task
type TaskJson struct {
	Name string `json:"name"`
}

// RunJson is a structure that will hold the information of the run
type RunJson struct {
	Status  string `json:"status"`
	IsDone  bool   `json:"is_done"`
	Success bool   `json:"success"`
}

// TaskGroupJson is a structure that will hold the information of the task group
type TaskGroupJson struct {
	Tasks      map[string]TaskJson `json:"tasks"`
	Runs       map[string]RunJson  `json:"runs"`
	Options    map[string]string   `json:"options"`
	NextNodeID string              `json:"next_node_id"`
}

type TaskSetJson struct {
	TaskSetID       string                   `json:"taskset_id"`
	SessionID       string                   `json:"session_id"`
	Graph           map[string]TaskGroupJson `json:"graph"`
	TaskGroupRoot   string                   `json:"task_group_root"`
	TaskSetProgress string                   `json:"task_set_progress"`
	TaskSetStatus   string                   `json:"task_set_status"`
	TaskSetOptions  map[string]string        `json:"task_set_options"`
}

func LogTasksetEvent(t *TaskSet) {
	// Create the json structure
	tasksetJson := TaskSetJson{
		TaskSetID:       t.TaskSetId,
		SessionID:       t.Session.SessionId,
		Graph:           make(map[string]TaskGroupJson),
		TaskGroupRoot:   t.RootNode.TaskGroupID,
		TaskSetStatus:   string(t.TaskSetStatus),
		TaskSetProgress: string(t.TaskSetProgress),
		TaskSetOptions:  t.TaskSetOptions,
	}

	// Populate the taskset json
	currentTaskGroup := t.RootNode
	for currentTaskGroup != nil {
		taskGroupJson := TaskGroupJson{
			Tasks:      make(map[string]TaskJson),
			Runs:       make(map[string]RunJson),
			Options:    currentTaskGroup.Options,
			NextNodeID: "",
		}

		// Populate the tasks
		for _, task := range currentTaskGroup.Tasks {
			taskGroupJson.Tasks[task.Id] = TaskJson{
				Name: task.Name,
			}
		}

		// Populate the runs
		for _, run := range currentTaskGroup.TaskRuns {
			success := false
			if run.Result != nil {
				success = run.Result.Success
			}
			taskGroupJson.Runs[run.TaskRunID] = RunJson{
				Status:  string(run.Status),
				Success: success,
				IsDone:  run.IsDone,
			}
		}

		// Populate the next node
		if currentTaskGroup.NextNode != nil {
			taskGroupJson.NextNodeID = currentTaskGroup.NextNode.TaskGroupID
		}

		// Add the task group to the taskset json
		tasksetJson.Graph[currentTaskGroup.TaskGroupID] = taskGroupJson
		currentTaskGroup = currentTaskGroup.NextNode
	}

	// Log the event
	eventLogJsonString, _ := json.Marshal(tasksetJson)
	event_group_id := fmt.Sprintf("%s-%s", t.TaskSetId, t.Session.SessionId)
	eventLog := NewEventLog("taskset", string(eventLogJsonString), event_group_id)
	t.Session.NodeScheduler.ResourceManager.LogEvent(eventLog)
}
