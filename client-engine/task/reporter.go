// Package task implements any task related operations
//
// This file contains the implementation of the reporter
// This structure is to be used accorss task-sets and tasks so that functions can be used as callbacks
//
// Author: Caio Cominato

package task

// Reporter is used to report an abstract event to a specific function.
type Reporter struct {
	Function func(args ...interface{})
	Context  []interface{}
}

// NewReporter initializes a new Reporter with the given function and context.
func NewReporter(function func(args ...interface{}), context []interface{}) *Reporter {
	return &Reporter{
		Function: function,
		Context:  context,
	}
}

// Report calls the function with the provided arguments.
func (r *Reporter) Report(args ...interface{}) {
	// Extend the context
	args = append(args, r.Context...)
	r.Function(args...)
}
