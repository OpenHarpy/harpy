package main

import (
	"client-engine/config"
	"client-engine/logger"
	"client-engine/task"
	"context"
	"errors"
	"io"
	"net"
	"runtime"
	"sync"

	pb "client-engine/grpc_ce_protocol"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func (lm *LiveMemory) CreateSession(options map[string]string) *task.Session {
	// Create a new session
	sess, err := task.NewSession(
		lm.RegisterCommandCallbackPointer,
		lm.RegisterCommandID,
		lm.DeregisterCommandCallbackPointer,
		options,
	)
	if err != nil {
		// TODO: We should handle this error
		logger.Error("failed_to_create_session", "SESSION-SERVICE", err)
		return nil
	}
	lm.Sessions[sess.SessionId] = sess
	return sess
}

func (lm *LiveMemory) CreateTaskSet(sess *task.Session) *task.TaskSet {
	// Create a new task set
	ts := sess.CreateTaskSet()
	lm.TaskSetDefinitions[ts.TaskSetId] = ts
	lm.TaskSetSession[ts.TaskSetId] = sess.SessionId
	return ts
}

// End LiveMemory
type CEgRPCServer struct {
	lm *LiveMemory
	pb.UnimplementedSessionServer
	pb.UnimplementedTaskSetServer
	pb.UnimplementedBlockProxyServer
}

// CreateSession implements
func (s *CEgRPCServer) CreateSession(ctx context.Context, in *pb.SessionRequest) (*pb.SessionHandler, error) {
	options := in.Options
	confgsToMerge := config.GetConfigs().GetAllConfigsWithPrefix("harpy.clientEngine")
	for key, value := range confgsToMerge {
		options[key] = value
	}
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

func (s *CEgRPCServer) DefineTask(ctx context.Context, taskDefinition *pb.TaskDefinition) (*pb.TaskHandler, error) {
	taskDefID := uuid.New().String()
	Arguments := make([]task.BlockID, len(taskDefinition.ArgumentsBlocks))
	for i, block := range taskDefinition.ArgumentsBlocks {
		Arguments[i] = task.BlockID(block)
	}
	Kwargs := make(map[string]task.BlockID)
	for key, block := range taskDefinition.KwargsBlocks {
		Kwargs[key] = task.BlockID(block)
	}
	taskDef := task.Definition{
		Id:                taskDefID,
		Name:              taskDefinition.Name,
		ExecutionType:     "remote",
		CallableBlockID:   task.BlockID(taskDefinition.CallableBlock),
		ArgumentsBlockIDs: Arguments,
		KwargsBlockIDs:    Kwargs,
	}
	s.lm.TaskDefinitions[taskDefID] = &taskDef
	logger.Info("Defined task", "TASKSET-SERVICE", logrus.Fields{"task_id": taskDefID})
	return &pb.TaskHandler{TaskID: taskDefID}, nil
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
		taskDef := task.NewMapperDefinition(*taskInMem)
		taskDefinitions = append(taskDefinitions, taskDef)
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
	taskDef := task.NewReducerDefinition(*taskInMem)
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
	taskDef := task.NewTransformerDefinition(*taskInMem)
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
	details["taskrun_status"] = string(tr.Status)
	details["taskrun_name"] = tr.Task.Name

	tsHandler := &pb.TaskSetHandler{TaskSetId: ts.TaskSetId}
	streamResult := pb.TaskSetProgressReport{
		TaskSetHandler:  tsHandler,
		ProgressType:    pb.ProgressType_TaskProgress,
		RelatedID:       tr.TaskRunID,
		ProgressMessage: "",
		StatusMessage:   string(tr.Status),
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
		logger.Error("Error executing task set", "TASKSET-SERVICE", err)
		return err
	}

	logger.Info("Task set executed", "TASKSET-SERVICE", logrus.Fields{"task_set_id": result.Results})
	session.RemoveTaskSetListener(uuid)
	return nil
}

func (s *CEgRPCServer) GetTaskSetResults(ctx context.Context, in *pb.TaskSetHandler) (*pb.TaskSetResult, error) {
	ts, ok := s.lm.TaskSetDefinitions[in.TaskSetId]
	if !ok {
		logger.Warn("task_set_not_found", "TASKSET-SERVICE", logrus.Fields{"task_set_id": in.TaskSetId})
		return &pb.TaskSetResult{TaskSetID: in.TaskSetId, OverallSuccess: false}, nil
	}
	logger.Info("Getting task set results", "TASKSET-SERVICE", logrus.Fields{"task_set_id": ts.TaskSetId})
	results := ts.GetTaskSetResult()
	taskResults := []*pb.TaskResult{}
	for _, result := range results.Results {
		taskResult := &pb.TaskResult{
			TaskRunID:         result.TaskRunID,
			ObjectReturnBlock: string(result.ObjectReturnBlockID),
			StdoutBlock:       string(result.StdoutBlockID),
			StderrBlock:       string(result.StderrBlockID),
			Success:           result.Success,
		}
		taskResults = append(taskResults, taskResult)
	}
	return &pb.TaskSetResult{
		TaskSetID:      ts.TaskSetId,
		OverallSuccess: results.OverallStatus == "success",
		TaskResults:    taskResults,
	}, nil

}

func (s *CEgRPCServer) Dismantle(ctx context.Context, in *pb.TaskSetHandler) (*pb.TaskSetHandler, error) {
	// Dismantle the task set - For this we will simply remove the task set from the live memory
	//  Hopefully the garbage collector will take care of the rest
	delete(s.lm.TaskSetDefinitions, in.TaskSetId)
	sessionID := s.lm.TaskSetSession[in.TaskSetId]
	session := s.lm.Sessions[sessionID]
	session.DismantleTaskSet(in.TaskSetId)
	logger.Info("Dismantled task set", "TASKSET-SERVICE", logrus.Fields{"task_set_id": in.TaskSetId})
	// Force the garbage collector to run
	runtime.GC()
	return &pb.TaskSetHandler{TaskSetId: "", Success: true}, nil
}

func (s *CEgRPCServer) PutBlock(stream pb.BlockProxy_PutBlockServer) error {
	sessionID := ""
	var blockWritter *task.BlockStreamingWriter = nil
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		sessionID = in.SessionHandler.SessionId
		if blockWritter == nil {
			session := s.lm.Sessions[sessionID]
			if session == nil {
				logger.Warn("session_not_found", "BLOCK-PROXY", logrus.Fields{"session_id": sessionID})
				return status.Error(404, "Session not found")
			}
			blockWritter = session.GetBlockWriter()
			if blockWritter == nil {
				err := errors.New("failed to create block writter")
				logger.Error("Failed to create block writter", "BLOCK-PROXY", err)
				return status.Error(500, "Unexpected error, unable to create block writter")
			}
		}
		blockWritter.Write(in.BlockChunk)
	}
	if blockWritter != nil {
		blockWritter.Close()
		stream.SendAndClose(&pb.ProxyBlockHandler{SessionHandler: &pb.SessionHandler{SessionId: sessionID}, BlockID: blockWritter.BlockID})
	} else {
		stream.SendAndClose(&pb.ProxyBlockHandler{SessionHandler: &pb.SessionHandler{SessionId: sessionID}, BlockID: ""})
	}
	return nil
}

func (s *CEgRPCServer) GetBlock(in *pb.ProxyBlockHandler, stream pb.BlockProxy_GetBlockServer) error {
	sessionID := in.SessionHandler.SessionId
	blockID := in.BlockID
	logger.Info("Getting block", "BLOCK-PROXY", logrus.Fields{"session_id": sessionID, "block_id": blockID})

	session := s.lm.Sessions[sessionID]
	if session == nil {
		logger.Warn("session_not_found", "BLOCK-PROXY", logrus.Fields{"session_id": sessionID})
		return status.Error(404, "Session not found")
	}

	blockReader := session.GetBlockReaderForBlock(blockID)
	if blockReader == nil {
		err := errors.New("failed to create block reader")
		logger.Error("Failed to create block reader", "BLOCK-PROXY", err)
		return status.Error(500, "Unexpected error, unable to create block reader")
	}

	sessionHandler := &pb.SessionHandler{SessionId: sessionID}

	for {
		chunk, done := blockReader.Read()
		if chunk != nil {
			err := stream.Send(&pb.ProxyBlockChunk{SessionHandler: sessionHandler, BlockChunk: chunk})
			if err != nil {
				return err
			}
		}
		if done {
			break
		}
	}
	return nil
}

// rpc GetInstanceID (SessionHandler) returns (InstanceMetadata) {}
func (s *CEgRPCServer) GetInstanceID(ctx context.Context, in *pb.SessionHandler) (*pb.InstanceMetadata, error) {
	// We will return the instance ID
	return &pb.InstanceMetadata{InstanceID: s.lm.InstanceID}, nil
}

func NewCEServer(exit chan bool, wg *sync.WaitGroup, lm *LiveMemory, port string) error {
	port = ":" + port
	logger.Info("gRPC server started", "SERVER", logrus.Fields{"host": port})

	s := grpc.NewServer()
	ceServer := &CEgRPCServer{
		lm: lm,
	}
	pb.RegisterSessionServer(s, ceServer)
	pb.RegisterTaskSetServer(s, ceServer)
	pb.RegisterBlockProxyServer(s, ceServer)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("failed_to_listen", "SERVER", err)
		defer wg.Done()
		return err
	} else {

		logger.Info("Server listening", "SERVER", logrus.Fields{"host": port})

		// Goroutine for the server
		go func() {
			if err := s.Serve(lis); err != nil {
				logger.Error("failed_to_serve", "SERVER", err)
				defer wg.Done()
			}
		}()
		go func() {
			<-exit
			logger.Info("Stopping server", "SERVER")
			s.GracefulStop()
			logger.Info("Server stopped", "SERVER")
			defer wg.Done()
		}()

	}
	return nil
}
