package main

import (
	"client-engine/chunker"
	"client-engine/logger"
	"client-engine/task"
	"context"
	"io"
	"net"
	"runtime"

	pb "client-engine/grpc_ce_protocol"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type TaskDefinition struct {
	Name            string
	CallableBinary  []byte
	ArgumentsBinary [][]byte
	KwargsBinary    map[string][]byte
}

func NewTaskDefinition(name string, callableBinary []byte, argumentsBinary [][]byte, kwargsBinary map[string][]byte) TaskDefinition {
	return TaskDefinition{
		Name:            name,
		CallableBinary:  callableBinary,
		ArgumentsBinary: argumentsBinary,
		KwargsBinary:    kwargsBinary,
	}
}

func (lm *LiveMemory) CreateSession(options map[string]string) *task.Session {
	// Create a new session
	sess := task.NewSession(options)
	lm.Sessions[sess.SessionId] = sess
	return sess
}

func (lm *LiveMemory) CreateTaskSet(sess *task.Session) *task.TaskSet {
	// Create a new task set
	ts := sess.CreateTaskSet()
	lm.TaskSetDefinitions[ts.TaskSetId] = ts
	return ts
}

// End LiveMemory
type CEgRPCServer struct {
	lm *LiveMemory
	pb.UnimplementedSessionServer
	pb.UnimplementedTaskSetServer
}

// CreateSession implements
func (s *CEgRPCServer) CreateSession(ctx context.Context, in *pb.SessionRequest) (*pb.SessionHandler, error) {
	options := in.Options
	sess := s.lm.CreateSession(options)
	logger.Info("Created session", "SESSION-SERVICE", logrus.Fields{"session_id": sess.SessionId})
	logger.Info("Session options", "SESSION-SERVICE", logrus.Fields{"options": in.Options})
	return &pb.SessionHandler{SessionId: sess.SessionId}, nil
}
func (s *CEgRPCServer) CloseSession(ctx context.Context, in *pb.SessionHandler) (*pb.SessionHandler, error) {
	// To close a session we will simply call close on the session and remove it from the live memory
	// Hopefully the garbage collector will take care of the rest
	sess, ok := s.lm.Sessions[in.SessionId]
	if !ok {
		logger.Warn("session_not_found", "SESSION-SERVICE", logrus.Fields{"session_id": in.SessionId})
		return &pb.SessionHandler{SessionId: ""}, nil
	}
	sess.Close()
	delete(s.lm.Sessions, in.SessionId)
	logger.Info("Closed session", "SESSION-SERVICE", logrus.Fields{"session_id": in.SessionId})
	runtime.GC() // Run the garbage collector
	return &pb.SessionHandler{SessionId: "", Success: true}, nil
}
func (s *CEgRPCServer) CreateTaskSet(ctx context.Context, in *pb.SessionHandler) (*pb.TaskSetHandler, error) {
	sess, ok := s.lm.Sessions[in.SessionId]
	if !ok {
		logger.Warn("session_not_found", "SESSION-SERVICE", logrus.Fields{"session_id": in.SessionId})
		return &pb.TaskSetHandler{TaskSetId: "", Success: false}, nil
	}
	ts := s.lm.CreateTaskSet(sess)
	logger.Info("Created task set", "SESSION-SERVICE", logrus.Fields{"task_set_id": ts.TaskSetId})
	return &pb.TaskSetHandler{TaskSetId: ts.TaskSetId}, nil
}

type TaskDefinitionFull struct {
	Name            string
	CallableBinary  []byte
	ArgumentsBinary [][]byte
	KwargsBinary    map[string][]byte
}

func (s *CEgRPCServer) DefineTask(stream pb.TaskSet_DefineTaskServer) error {
	callableBinary := []byte{}
	argumentsBinary := map[uint32][]byte{}
	kwargsBinary := map[string][]byte{}
	name := ""

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name = in.Name
		// The callable binary will be sent in chunks
		callableBinary = append(callableBinary, in.CallableBinary...)
		// The arguments binary will also be sent in chunks but there will be multiple arguments per chunk
		// It will be sent as a map of integer corresponding tp the index of the argument
		for key, value := range in.ArgumentsBinary {
			arg, ok := argumentsBinary[key]
			if !ok {
				argumentsBinary[key] = []byte{}
			}
			argumentsBinary[key] = append(arg, value...)
		}
		// Similarly the kwargs binary will be sent in chunks but there will be multiple kwargs per chunk
		// It will be sent as a map of string corresponding to the key of the kwarg
		for key, value := range in.KwargsBinary {
			kwarg, ok := kwargsBinary[key]
			if !ok {
				kwargsBinary[key] = []byte{}
			}
			kwargsBinary[key] = append(kwarg, value...)
		}
	}

	// once the stream is done we will create a task definition in the live memory
	argumentsSlice := make([][]byte, len(argumentsBinary))
	for key, value := range argumentsBinary {
		argumentsSlice[key] = value
	}

	taskDef := NewTaskDefinition(
		name,
		callableBinary,
		argumentsSlice,
		kwargsBinary,
	)
	taskDefId := uuid.New().String()
	s.lm.TaskDefinitions[taskDefId] = &taskDef
	logger.Info("Defined task", "TASKSET-SERVICE", logrus.Fields{"task_id": taskDefId})
	return stream.SendAndClose(&pb.TaskHandler{TaskID: taskDefId})
}

func (s *CEgRPCServer) AddMap(ctx context.Context, in *pb.MapAdder) (*pb.TaskAdderResult, error) {
	ts, ok := s.lm.TaskSetDefinitions[in.TaskSetHandler.TaskSetId]
	if !ok {
		logger.Warn("task_set_not_found", "TASKSET-SERVICE", logrus.Fields{"task_set_id": in.TaskSetHandler.TaskSetId})
		return &pb.TaskAdderResult{Success: false, ErrorMesssage: "Task Set not found"}, nil
	}

	// For each taskHandler under MappersDefinition we will create a task definition
	taskDefinitions := []task.MapperDefinition{}
	for _, taskHandler := range in.MappersDefinition {
		taskInMem, ok := s.lm.TaskDefinitions[taskHandler.TaskID]
		if !ok {
			logger.Warn("task_definition_not_found", "TASKSET-SERVICE", logrus.Fields{"task_id": taskHandler.TaskID})
			return &pb.TaskAdderResult{Success: false, ErrorMesssage: "Task Definition not found"}, nil
		}
		taskDef := task.NewMapperDefinition(
			taskInMem.Name,
			taskInMem.CallableBinary,
			"remote",
			taskInMem.ArgumentsBinary,
			taskInMem.KwargsBinary,
		)
		taskDefinitions = append(taskDefinitions, taskDef)
		// We can now flush the task definition from the live memory
		delete(s.lm.TaskDefinitions, taskHandler.TaskID)
	}
	opt := map[string]string{}
	ts.Map(taskDefinitions, opt)
	logger.Info("Added map to task set", "TASKSET-SERVICE", logrus.Fields{"task_set_id": in.TaskSetHandler.TaskSetId})
	return &pb.TaskAdderResult{Success: true, ErrorMesssage: ""}, nil
}

func (s *CEgRPCServer) AddReduce(ctx context.Context, in *pb.ReduceAdder) (*pb.TaskAdderResult, error) {
	ts, ok := s.lm.TaskSetDefinitions[in.TaskSetHandler.TaskSetId]
	if !ok {
		logger.Warn("task_set_not_found", "TASKSET-SERVICE", logrus.Fields{"task_set_id": in.TaskSetHandler.TaskSetId})
		return &pb.TaskAdderResult{Success: false, ErrorMesssage: "Task Set not found"}, nil
	}

	// Reducers only have one task definition
	taskHandler := in.ReducerDefinition
	taskInMem, ok := s.lm.TaskDefinitions[taskHandler.TaskID]
	if !ok {
		logger.Warn("task_definition_not_found", "TASKSET-SERVICE", logrus.Fields{"task_id": taskHandler.TaskID})
		return &pb.TaskAdderResult{Success: false, ErrorMesssage: "Task Definition not found"}, nil
	}
	taskDef := task.NewReducerDefinition(
		taskInMem.Name,
		taskInMem.CallableBinary,
		"remote",
		taskInMem.ArgumentsBinary,
		taskInMem.KwargsBinary,
	)
	// We can now flush the task definition from the live memory
	delete(s.lm.TaskDefinitions, taskHandler.TaskID)
	opt := map[string]string{}
	ts.Reduce(taskDef, opt)
	logger.Info("Added reduce to task set", "TASKSET-SERVICE", logrus.Fields{"task_set_id": in.TaskSetHandler.TaskSetId})
	return &pb.TaskAdderResult{Success: true, ErrorMesssage: ""}, nil
}

func (s *CEgRPCServer) AddTransform(ctx context.Context, in *pb.TransformAdder) (*pb.TaskAdderResult, error) {
	ts, ok := s.lm.TaskSetDefinitions[in.TaskSetHandler.TaskSetId]
	if !ok {
		logger.Warn("task_set_not_found", "TASKSET-SERVICE", logrus.Fields{"task_set_id": in.TaskSetHandler.TaskSetId})
		return &pb.TaskAdderResult{Success: false, ErrorMesssage: "Task Set not found"}, nil
	}

	// Transformers only have one task definition
	taskHandler := in.TransformerDefinition
	taskInMem, ok := s.lm.TaskDefinitions[taskHandler.TaskID]
	if !ok {
		logger.Warn("task_definition_not_found", "TASKSET-SERVICE", logrus.Fields{"task_id": taskHandler.TaskID})
		return &pb.TaskAdderResult{Success: false, ErrorMesssage: "Task Definition not found"}, nil
	}
	taskDef := task.NewTransformerDefinition(
		taskInMem.Name,
		taskInMem.CallableBinary,
		"remote",
		taskInMem.ArgumentsBinary,
		taskInMem.KwargsBinary,
	)
	// We can now flush the task definition from the live memory
	delete(s.lm.TaskDefinitions, taskHandler.TaskID)
	opt := map[string]string{}
	ts.Transform(taskDef, opt)
	logger.Info("Added transform to task set", "TASKSET-SERVICE", logrus.Fields{"task_set_id": in.TaskSetHandler.TaskSetId})
	return &pb.TaskAdderResult{Success: true, ErrorMesssage: ""}, nil
}

type TaskSetStreamListener struct {
	// We add a stream here to listen for the task set results
	stream pb.TaskSet_ExecuteServer
}

func (l *TaskSetStreamListener) OnTaskSetProgress(ts *task.TaskSet) {
	// We will send the task set progress to the client
	details := map[string]string{}

	details["taskset_progress"] = ts.TaskSetProgress
	details["taskset_status"] = ts.TaskSetStatus

	tsHandler := &pb.TaskSetHandler{TaskSetId: ts.TaskSetId}
	streamResult := pb.TaskSetProgressReport{
		TaskSetHandler:  tsHandler,
		ProgressType:    pb.ProgressType_TaskSetProgress,
		RelatedID:       ts.TaskSetId,
		ProgressMessage: ts.TaskSetProgress,
		StatusMessage:   ts.TaskSetStatus,
		ProgressDetails: details,
	}
	l.stream.Send(&streamResult)
}

func (l *TaskSetStreamListener) OnTaskGroupProgress(ts *task.TaskSet, tg *task.TaskGroup) {
	// We will send the task set results to the client
	details := map[string]string{}

	details["taskgroup_progress"] = ts.TaskSetProgress
	details["taskgroup_status"] = ts.TaskSetStatus
	details["taskgroup_id"] = tg.TaskGroupID
	details["taskgroup_name"] = tg.Name

	tsHandler := &pb.TaskSetHandler{TaskSetId: ts.TaskSetId}
	streamResult := pb.TaskSetProgressReport{
		TaskSetHandler:  tsHandler,
		ProgressType:    pb.ProgressType_TaskGroupProgress,
		RelatedID:       tg.TaskGroupID,
		ProgressMessage: "",
		StatusMessage:   "",
		ProgressDetails: details,
	}
	l.stream.Send(&streamResult)
}

func (l *TaskSetStreamListener) OnTaskProgress(ts *task.TaskSet, tr *task.TaskRun) {
	// We will send the task set done to the client
	details := map[string]string{}

	details["taskset_progress"] = ts.TaskSetProgress
	details["taskset_status"] = ts.TaskSetStatus
	details["taskrun_id"] = tr.TaskRunID
	details["taskrun_status"] = tr.Status
	details["taskrun_name"] = tr.Task.Name

	tsHandler := &pb.TaskSetHandler{TaskSetId: ts.TaskSetId}
	streamResult := pb.TaskSetProgressReport{
		TaskSetHandler:  tsHandler,
		ProgressType:    pb.ProgressType_TaskProgress,
		RelatedID:       tr.TaskRunID,
		ProgressMessage: "",
		StatusMessage:   tr.Status,
		ProgressDetails: details,
	}

	l.stream.Send(&streamResult)
}

func (s *CEgRPCServer) Execute(in *pb.TaskSetHandler, stream pb.TaskSet_ExecuteServer) error {
	// Prior to executing we will force the garbage collector to run
	runtime.GC() // This is because we are removing task definitions from the live memory and we want to free up the memory
	ts := s.lm.TaskSetDefinitions[in.TaskSetId]
	logger.Info("Executing task set", "TASKSET-SERVICE", logrus.Fields{"task_set_id": ts.TaskSetId})

	// We add a new listener to the task set
	session := ts.Session
	listener := &TaskSetStreamListener{stream: stream}
	uuid := session.AddTaskSetListener(listener)

	result, err := ts.Execute()
	if err != nil {
		logger.Error("task_execution_error", "Error executing task set", err)
		return err
	}

	logger.Info("Task set executed", "TASKSET-SERVICE", logrus.Fields{"task_set_id": result.Results})
	session.RemoveTaskSetListener(uuid)
	return nil
}

// GetTaskSetResults implements (Streaming)
//
//	rpc GetTaskSetResults (TaskSetHandler) returns (stream TaskSetResult) {}
func (s *CEgRPCServer) GetTaskSetResults(in *pb.TaskSetHandler, stream pb.TaskSet_GetTaskSetResultsServer) error {
	ts := s.lm.TaskSetDefinitions[in.TaskSetId]
	logger.Info("Getting task set result", "TASKSET-SERVICE", logrus.Fields{"task_set_id": ts.TaskSetId})

	tsResult := ts.GetTaskSetResult()
	if tsResult.OverallStatus == "pending" {
		// If the task set is still pending then we will just return
		return nil
	} else {
		// If the task set is completed then we will stream the results
		binary_index_map := map[int]*int{}  // map of result index to chunk index
		binary_stdout_map := map[int]*int{} // map of result index to chunk index
		binary_stderr_map := map[int]*int{} // map of result index to chunk index

		done_stream := map[int]bool{} // map of result index to done status
		all_streamed := false
		aggBytes := 0
		for !all_streamed {
			// While not all results ended
			for index, result := range tsResult.Results {
				_, ok := binary_index_map[index]
				if !ok {
					binary_index_map[index] = new(int)
					*binary_index_map[index] = 0
					binary_stdout_map[index] = new(int)
					*binary_stdout_map[index] = 0
					binary_stderr_map[index] = new(int)
					*binary_stderr_map[index] = 0
				}
				_, ok = done_stream[index]
				if !ok {
					done_stream[index] = false
				}

				current_done := done_stream[index]
				if !current_done {
					taskSetResultChunk := &pb.TaskSetResultChunk{
						TaskRunID:               result.TaskRunID,
						ObjectReturnBinaryChunk: chunker.ChunkBytes(result.ObjectReturnBinary, binary_index_map[index]),
						StdoutBinaryChunk:       chunker.ChunkBytes(result.StdoutBinary, binary_stdout_map[index]),
						StderrBinaryChunk:       chunker.ChunkBytes(result.StderrBinary, binary_stderr_map[index]),
						Success:                 result.Success,
					}

					aggBytes += len(taskSetResultChunk.ObjectReturnBinaryChunk)
					if len(taskSetResultChunk.ObjectReturnBinaryChunk) == 0 && len(taskSetResultChunk.StdoutBinaryChunk) == 0 && len(taskSetResultChunk.StderrBinaryChunk) == 0 {
						done_stream[index] = true
					}
					// Send the chunk
					err := stream.Send(taskSetResultChunk)
					if err != nil {
						return err
					}
				}
				// Check if all results are done
				exit_next := true
				for _, done := range done_stream {
					if !done {
						exit_next = false
						break
					}
				}
				all_streamed = exit_next
			}
		}
	}
	return nil
}

func (s *CEgRPCServer) Dismantle(ctx context.Context, in *pb.TaskSetHandler) (*pb.TaskSetHandler, error) {
	// Dismantle the task set - For this we will simply remove the task set from the live memory
	//  Hopefully the garbage collector will take care of the rest
	delete(s.lm.TaskSetDefinitions, in.TaskSetId)
	logger.Info("Dismantled task set", "TASKSET-SERVICE", logrus.Fields{"task_set_id": in.TaskSetId})
	return &pb.TaskSetHandler{TaskSetId: "", Success: true}, nil
}

func NewCEServer(exit chan bool, panic chan bool, port string, lm *LiveMemory) {
	port = ":" + port
	logger.Info("gRPC server started", "SERVER", logrus.Fields{"host": port})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("failed_to_listen", "SERVER", err)
		panic <- true
		return
	}
	s := grpc.NewServer()
	ceServer := &CEgRPCServer{
		lm: lm,
	}
	pb.RegisterSessionServer(s, ceServer)
	pb.RegisterTaskSetServer(s, ceServer)
	// Goroutine to listen for exit signal
	if err := s.Serve(lis); err != nil {
		logger.Error("failed_to_serve", "SERVER", err)
		panic <- true
	}

	go func() {
		<-exit
		logger.Info("Received exit signal, stopping server...", "SERVER")
		s.GracefulStop()
	}()

}
