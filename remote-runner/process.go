package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"remote-runner/config"
	"remote-runner/logger"
	"time"
)

const (
	PROCESS_DEFINED = "defined"
	PROCESS_QUEUED  = "queued"
	PROCESS_KILLING = "killing"
	PROCESS_RUNNING = "running"
	PROCESS_DONE    = "done"
	PROCESS_PANIC   = "panic"
)

type Process struct {
	ProcessID           string
	ProcessStatus       string
	CallableBinary      []byte
	ArgumentsBinary     [][]byte
	KwargsBinary        map[string][]byte
	ObjectReturnBinary  []byte
	StdoutBinary        []byte
	StderrBinary        []byte
	ProcessRegisterTime int64
	CallbackID          string
	ResultFetchTime     int64
	Success             bool
	KillChan            chan bool
	DoneChan            chan bool
}

func NewProcess(processID string, callableBinary []byte, argumentsBinary [][]byte, kwargsBinary map[string][]byte) *Process {
	currentTime := time.Now().Unix()
	return &Process{
		ProcessID:       processID,
		ProcessStatus:   PROCESS_DEFINED,
		CallableBinary:  callableBinary,
		ArgumentsBinary: argumentsBinary,
		KwargsBinary:    kwargsBinary,

		ProcessRegisterTime: currentTime,
		ResultFetchTime:     0,
		CallbackID:          "",
		Success:             false,
		KillChan:            nil,
		DoneChan:            nil,
	}
}

func RunProcess(killChan chan bool, doneChan chan bool, process *Process) {
	// This function execute the process
	entrypoint := config.GetConfigs().GetConfigsWithDefault("command_entrypoint", "")
	binaryLocation := config.GetConfigs().GetConfigsWithDefault("temporary_binaries", "")
	if entrypoint == "" || binaryLocation == "" {
		logger.Error("Missing configuration", "RunProcess", nil)
		panic("Missing configuration")
	}

	fullCommand := entrypoint + " " + binaryLocation + " " + process.ProcessID
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", fullCommand)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		logger.Error("Error starting process", "RunProcess", err)
	}
	// Wait for the process to finish
	err = cmd.Wait()
	if err != nil {
		logger.Error("Error running process", "RunProcess", err)
		process.Success = false
	} else {
		process.Success = true
	}
	// Read the output to the process object
	process.StdoutBinary = stdout.Bytes()
	process.StderrBinary = stderr.Bytes()
	binaryLocation = binaryLocation + "/" + process.ProcessID + "/return_object.cloudpickle"
	if _, err := os.Stat(binaryLocation); os.IsNotExist(err) {
		// The return object does not exist so we will set return object to empty
		process.ObjectReturnBinary = []byte{}
		doneChan <- true
		return
	}
	// Read the return object
	returnObject, err := os.ReadFile(binaryLocation)
	if err != nil {
		logger.Error("Error reading return object", "RunProcess", err)
	}
	process.ObjectReturnBinary = returnObject
	// Send the done signal
	doneChan <- true

	// TODO: cleanup the process implement kill
}

func WriteBinary(location string, binary []byte) error {
	file, err := os.Create(location)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(binary)
	if err != nil {
		return err
	}
	return nil
}

func MakeBinariesForProcess(process *Process) {
	binaryLocation := config.GetConfigs().GetConfigsWithDefault("temporary_binaries", "./_live_objects")
	// Make folder if not exists
	root := binaryLocation + "/" + process.ProcessID
	err := os.MkdirAll(root, os.ModePerm)
	if err != nil {
		logger.Error("Error creating folder", "MakeBinariesForProcess", err)
	}
	// Convert the process to binary
	callableBinaryLocation := root + "/callable.cloudpickle"
	err = WriteBinary(callableBinaryLocation, process.CallableBinary)
	if err != nil {
		logger.Error("Error writing binary", "MakeBinariesForProcess", err)
	}
	// Arguments to binary
	for i, arg := range process.ArgumentsBinary {
		argLocation := root + "/arg_" + fmt.Sprint(i) + ".cloudpickle"
		err = WriteBinary(argLocation, arg)
		if err != nil {
			logger.Error("Error writing binary", "MakeBinariesForProcess", err)
		}
	}
	// Keyword arguments to binary
	kwargIdx := 0
	for key, kwarg := range process.KwargsBinary {
		kwargLocation := root + "/kwarg_" + fmt.Sprint(kwargIdx) + ".cloudpickle"
		// For the kwarg we need to encode the key as well
		// To do this we will add the key in plaintext encoded as UTF-8
		// The binary will always be separated by C0 80 as this will not be a valid UTF-8 character
		// This will allow us to split the key from the value

		// Write the key
		keyBytesUTF8 := []byte(key)                     // Convert the key to bytes
		keyBytesUTF8 = append(keyBytesUTF8, 0xC0, 0x80) // Add the separator
		fullData := append(keyBytesUTF8, kwarg...)      // Combine the key and the value
		err = WriteBinary(kwargLocation, fullData)
		if err != nil {
			logger.Error("Error writing binary", "MakeBinariesForProcess", err)
		}
		kwargIdx++
	}
}

func CleanupBinaries(process *Process) {
	binaryLocation := config.GetConfigs().GetConfigsWithDefault("temporary_binaries", "./_live_objects")
	root := binaryLocation + "/" + process.ProcessID
	err := os.RemoveAll(root)
	if err != nil {
		logger.Error("Error removing folder", "CleanupBinaries", err)
	}
}

func (p *Process) Run() {
	MakeBinariesForProcess(p)
	// Create the chans for the process
	p.DoneChan = make(chan bool)
	p.KillChan = make(chan bool)
	// Start the process
	go RunProcess(p.KillChan, p.DoneChan, p)
}
func (p *Process) Cleanup() {
	CleanupBinaries(p)
}
func (p *Process) QueueProcess(CallbackID string) {
	p.ProcessStatus = PROCESS_QUEUED
	p.CallbackID = CallbackID
}
func (p *Process) KillProcess(CallbackID string) {
	p.ProcessStatus = PROCESS_KILLING
	p.CallbackID = CallbackID
	p.KillChan <- true
}
func (p *Process) SetResultFetchTime() {
	p.ResultFetchTime = time.Now().Unix()
}

func ExitCleanAll(lm *LiveMemory) {
	// TODO: Implement this function
}
