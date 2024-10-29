// Package task implements any task related operations
//
// This file contains the implementation of the task factories
// The task factories are used to create the tasks that are going to be executed
// The abstraction is necessary because each underlying task is a "TaskDefinition" (see task_types.go)
// Dispite that, tasks are different from each other and have different requirements
// Nodes are simple creatures that only do "TaskDefinition" -> "TaskResult".
// But Mapper, Transformer and Reducer are different from each other and have different requirements even if they are all TaskDefinitions.
// The task factories are what allows data to flow from previous tasks to the next tasks
// 	The taskgroup may not know how many tasks are going to get executed befor the taskgroup is executed
//  Hence we need factories because we depend on the taskset results.
//
// Author: Caio Cominato

package task

import "fmt"

// ** MapFactory **
type MapFactory struct {
	Mappers []MapperDefinition
}

func (m MapFactory) MakeTasks(previousResult TaskGroupResult) []TaskDefinition {
	// The mappers can only be the first task in the task group, so we will not have any previous results
	//   We only get the previous results because that is how the interface is defined in the TaskFactory
	//   The interface defined in this way simplifies the implementation of the task group
	//   The taskset will be responsible for managing the task groups and the task groups will be responsible for managing the tasks
	var tasks []TaskDefinition
	for _, mapper := range m.Mappers {
		tasks = append(tasks, TaskDefinition(mapper))
	}
	return tasks
}
func (m MapFactory) String() string {
	return fmt.Sprintf("MapFactory with %d mappers", len(m.Mappers))
}

// ** TransformFactory **
type TransformFactory struct {
	Transformer TransformerDefinition
}

func (t TransformFactory) MakeTasks(previousResult TaskGroupResult) []TaskDefinition {
	// The transformer will follow the same partitioning as the mappers in the previous result
	//   Each result will be passed to the transformer as an argument (parallel execution)
	tasks := []TaskDefinition{}
	for _, result := range previousResult.Results {
		if !result.Success {
			return []TaskDefinition{}
		}
		task := TaskDefinition(t.Transformer)
		args := t.Transformer.ArgumentsBlockIDs
		// Add the result object to the beginning of the arguments
		args = append([]BlockID{result.ObjectReturnBlockID}, args...)
		task.ArgumentsBlockIDs = args
		tasks = append(tasks, task)
	}
	return tasks
}
func (t TransformFactory) String() string {
	return fmt.Sprintf("TransformFactory with transformer %s", t.Transformer.String())
}

// ** ReduceFactory **
type ReduceFactory struct {
	Reducer ReducerDefinition
}

func (r ReduceFactory) MakeTasks(previousResult TaskGroupResult) []TaskDefinition {
	// The reducer will only have one task
	//   It will have all the object return binaries from the previous results as arguments
	//   This is will result in a single task that will reduce all the results into a single result
	task := TaskDefinition(r.Reducer)
	task.ArgumentsBlockIDs = []BlockID{}
	for _, result := range previousResult.Results {
		task.ArgumentsBlockIDs = append(task.ArgumentsBlockIDs, result.ObjectReturnBlockID)
	}
	return []TaskDefinition{task}
}
func (r ReduceFactory) String() string {
	return fmt.Sprintf("ReduceFactory with reducer %s", r.Reducer.String())
}

// ** FanoutFactory **
type FanoutFactory struct {
	Fanout      FanoutDefinition
	FanoutCount int
}

func (f FanoutFactory) MakeTasks(previousResult TaskGroupResult) []TaskDefinition {
	// The fanout works like transformers, but it will exapand each result int "FanoutCount" tasks
	//  Each task will have the same arguments as the transformer
	tasks := []TaskDefinition{}
	for i, result := range previousResult.Results {
		for y := 0; y < f.FanoutCount; y++ {
			if !result.Success {
				return []TaskDefinition{}
			}
			task := TaskDefinition(f.Fanout)
			task.Metadata = map[string]string{"fanout_index": fmt.Sprintf("%d", y), "fanout_result_index": fmt.Sprintf("%d", i)}
			args := f.Fanout.ArgumentsBlockIDs
			// Add the result object to the beginning of the arguments
			args = append([]BlockID{result.ObjectReturnBlockID}, args...)
			task.ArgumentsBlockIDs = args
			tasks = append(tasks, task)
		}
	}
	return tasks
}
func (f FanoutFactory) String() string {
	return fmt.Sprintf("FanoutFactory with fanout %s and fanout count %d", f.Fanout.String(), f.FanoutCount)
}
