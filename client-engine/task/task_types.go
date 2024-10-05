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

	"github.com/google/uuid"
)

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
	Id              string
	Name            string
	CallableBinary  []byte
	ExecutionType   string
	ArgumentsBinary [][]byte
	KwargsBinary    map[string][]byte
}

// TransformerDefinition --> This is used to define a transformer that will belong to a task group
type TransformerDefinition struct {
	Definition
}

func (t TransformerDefinition) String() string {
	return fmt.Sprintf("TransformerDefinition[%s]", t.Name)
}

func NewTransformerDefinition(name string, callableBinary []byte, executionType string, argumentsBinary [][]byte, kwargsBinary map[string][]byte) TransformerDefinition {
	idx := fmt.Sprintf("tf-%s", uuid.New().String())
	return TransformerDefinition{
		Definition: Definition{
			Id:              idx,
			Name:            name,
			CallableBinary:  callableBinary,
			ExecutionType:   executionType,
			ArgumentsBinary: argumentsBinary,
			KwargsBinary:    kwargsBinary,
		},
	}
}

// ReducerDefinition --> This is used to define a reducer that will belong to a task group
type ReducerDefinition struct {
	Definition
}

func (r ReducerDefinition) String() string {
	return fmt.Sprintf("ReducerDefinition[%s]", r.Name)
}

func NewReducerDefinition(name string, callableBinary []byte, executionType string, argumentsBinary [][]byte, kwargsBinary map[string][]byte) ReducerDefinition {
	idx := fmt.Sprintf("rd-%s", uuid.New().String())
	return ReducerDefinition{
		Definition: Definition{
			Id:              idx,
			Name:            name,
			CallableBinary:  callableBinary,
			ExecutionType:   executionType,
			ArgumentsBinary: argumentsBinary,
			KwargsBinary:    kwargsBinary,
		},
	}
}

// MapperDefinition --> This is used to define a mapper that will belong to a task group

type MapperDefinition struct {
	Definition
}

func (m MapperDefinition) String() string {
	return fmt.Sprintf("MapperDefinition[%s]", m.Name)
}

func NewMapperDefinition(name string, callableBinary []byte, executionType string, argumentsBinary [][]byte, kwargsBinary map[string][]byte) MapperDefinition {
	idx := fmt.Sprintf("mp-%s", uuid.New().String())

	return MapperDefinition{
		Definition: Definition{
			Id:              idx,
			Name:            name,
			CallableBinary:  callableBinary,
			ExecutionType:   executionType,
			ArgumentsBinary: argumentsBinary,
			KwargsBinary:    kwargsBinary,
		},
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
