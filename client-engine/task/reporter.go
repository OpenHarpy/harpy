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

//func (r *Reporter) getContextKey(key string) interface{} {
//	return r.Context[key]
//}
//func (r *Reporter) setContextKey(key string, value interface{}) {
//	r.Context[key] = value
//}
