// Package task implements any task related operations
//
// This file contains the implementation of the task types
// The task types are used to define the structure of the tasks that are going to be executed
// These definitions are used to hold important information about the tasks that are going to be executed
//
// Author: Caio Cominato

package task

import (
	"fmt"
)

type BlockID string

// DefinitionInterface defines the abstract method
type DefinitionInterface interface {
	String() string
}

type Definition struct {
	// This is different from the original python code
	//	We will have a language barrier here, so we need to treat the arguments as a list of strings
	//	This will be the same for the kwargs
	// 	This changes the structure of the definition so we will need to update the RPM calls to reflect this
	//  TODO: Update the RPM calls to reflect the new structure
	Id                string
	Name              string
	CallableBlockID   BlockID
	ExecutionType     string
	ArgumentsBlockIDs []BlockID
	KwargsBlockIDs    map[string]BlockID
	Metadata          map[string]string
}

// TransformerDefinition --> This is used to define a transformer that will belong to a task group
type TransformerDefinition struct {
	Definition
}

func (t TransformerDefinition) String() string {
	return fmt.Sprintf("TransformerDefinition[%s]", t.Name)
}

func NewTransformerDefinition(definition Definition) TransformerDefinition {
	return TransformerDefinition{
		Definition: definition,
	}
}

// ReducerDefinition --> This is used to define a reducer that will belong to a task group
type ReducerDefinition struct {
	Definition
}

func (r ReducerDefinition) String() string {
	return fmt.Sprintf("ReducerDefinition[%s]", r.Name)
}

func NewReducerDefinition(definition Definition) ReducerDefinition {
	return ReducerDefinition{
		Definition: definition,
	}
}

// MapperDefinition --> This is used to define a mapper that will belong to a task group
type MapperDefinition struct {
	Definition
}

func (m MapperDefinition) String() string {
	return fmt.Sprintf("MapperDefinition[%s]", m.Name)
}

func NewMapperDefinition(definition Definition) MapperDefinition {
	return MapperDefinition{
		Definition: definition,
	}
}

// FanoutDefinition --> This is used to define a fanout that will belong to a task group
type FanoutDefinition struct {
	Definition
}

func (f FanoutDefinition) String() string {
	return fmt.Sprintf("FanoutDefinition[%s]", f.Name)
}

func NewFanoutDefinition(definition Definition) FanoutDefinition {
	return FanoutDefinition{
		Definition: definition,
	}
}

// TaskDefinition --> This is used to define a task that will belong to a task group
//
//	This is different from the other definitions because it is the only one that will be executed in the task group
//	Taskgroups will not have a transformer, reducer, or mapper definition directly associated with them

type TaskDefinition struct {
	Definition
}

func (t TaskDefinition) String() string {
	return fmt.Sprintf("TaskDefinition[%s]", t.Name)
}
