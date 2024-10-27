package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"remote-runner/config"
	"remote-runner/logger"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	PROCESS_DEFINED = "defined"
	PROCESS_QUEUED  = "queued"
	PROCESS_KILLING = "killing"
	PROCESS_RUNNING = "running"
	PROCESS_DONE    = "done"
	PROCESS_PANIC   = "panic"
)

type IsolatedEnvironment struct {
	EnvironmentID string // The ID of the environment (Session ID)
	// Later we can use this to store more info like maybe the packages installed or the instance id that requested the environment
	LastAccessTime int64
	deepIsolation  bool
}

func NewIsolatedEnvironment(environmentID string) *IsolatedEnvironment {
	return &IsolatedEnvironment{
		EnvironmentID: environmentID,
	}
}
func (ie *IsolatedEnvironment) Cleanup() error {
	if !ie.deepIsolation {
		return nil
	}
	// Cleanup the environment
	root := config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.scriptsRoot", "")
	cleanupScript := config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.isolatedEnvironmentCleanupScript", "")
	fullPath := root + "/" + cleanupScript
	if cleanupScript == "" || fullPath == "" {
		err := fmt.Errorf("missing configuration")
		logger.Error("Missing configuration", "ISOLATED_ENVIRONMENT", err)
		return err
	}
	err := ExecCommandEcho(fullPath, ie.EnvironmentID)
	if err != nil {
		logger.Error("error running node cleanup script", "ISOLATED_ENVIRONMENT", err)
		return err
	}
	logger.Info("node cleanup script executed", "ISOLATED_ENVIRONMENT")
	return nil
}
func (ie *IsolatedEnvironment) InstallPackage(packageName string) error {
	// Install the package
	return nil
}
func (ie *IsolatedEnvironment) UninstallPackage(packageName string) error {
	// Uninstall the package
	return nil
}
func (ie *IsolatedEnvironment) Begin() error {
	// We start by checking the params
	root := config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.scriptsRoot", "")
	entrypoint := config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.commandEntrypoint", "")
	if entrypoint == "" || root == "" {
		err := fmt.Errorf("missing configuration")
		logger.Error("Missing configuration", "ISOLATED_ENVIRONMENT", err)
		return err
	}

	// Starting the environment
	logger.Info("Setting up node", "ISOLATED_ENVIRONMENT")
	isolatedFlag := config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.deepIsolation", "true")
	if isolatedFlag == "false" {
		logger.Info("Isolated environment disabled", "ISOLATED_ENVIRONMENT")
		ie.deepIsolation = false
		return nil
	} else {
		ie.deepIsolation = true
	}
	// We will start by running the command
	setupScript := config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.isolatedEnvironmentSetupScript", "")
	fullPath := root + "/" + setupScript
	entrypointFull := root + "/" + entrypoint
	if setupScript == "" {
		err := fmt.Errorf("missing configuration")
		logger.Error("Missing configuration", "ISOLATED_ENVIRONMENT", err)
		return err
	}
	err := ExecCommandEcho(fullPath, ie.EnvironmentID, entrypointFull)
	if err != nil {
		logger.Error("error running node setup script", "ISOLATED_ENVIRONMENT", err)
		return err
	}
	logger.Info("node setup script executed", "ISOLATED_ENVIRONMENT")
	return nil
}
func (ie *IsolatedEnvironment) GetEntrypoint() string {
	now := time.Now().Unix()
	root := config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.scriptsRoot", "")
	if !ie.deepIsolation {
		return root + "/" + config.GetConfigs().GetConfigWithDefault("harpy.remoteRunner.commandEntrypoint", "")
	}
	ie.LastAccessTime = now
	return root + "/isolated-" + ie.EnvironmentID + "/entrypoint.sh"
}

type Process struct {
	ProcessID           string
	ProcessStatus       string
	CallableBlock       *Block
	ArgumentsBlocks     []*Block
	KwargsBlocks        map[string]*Block
	ObjectReturnBlock   *Block
	StdoutBlock         *Block
	StderrBlock         *Block
	OutputBlock         *Block
	ProcessRegisterTime int64
	CallbackID          string
	ResultFetchTime     int64
	Success             bool
	KillChan            chan bool
	DoneChan            chan bool
	EnvRef              *IsolatedEnvironment
}

func NewProcess(
	processID string, callableBlock *Block, argumentsBlocks []*Block, kwargsBlocks map[string]*Block,
	OutputBlock *Block,
	StdOutBlock *Block, StdErrBlock *Block,
) *Process {
	currentTime := time.Now().Unix()
	return &Process{
		ProcessID:       processID,
		ProcessStatus:   PROCESS_DEFINED,
		CallableBlock:   callableBlock,
		ArgumentsBlocks: argumentsBlocks,
		KwargsBlocks:    kwargsBlocks,
		OutputBlock:     OutputBlock,
		StdoutBlock:     StdOutBlock,
		StderrBlock:     StdErrBlock,

		ProcessRegisterTime: currentTime,
		ResultFetchTime:     0,
		CallbackID:          "",
		Success:             false,
		KillChan:            nil,
		DoneChan:            nil,
	}
}

func NormalizeCommand(command string) string {
	// Remove multiple spaces and replace with single space
	command = strings.Join(strings.Fields(command), " ")
	// Remove leading and trailing spaces
	command = strings.TrimSpace(command)
	// Remove double slashes
	command = strings.ReplaceAll(command, "//", "/")
	return command
}

func RunProcess(killChan chan bool, doneChan chan bool, process *Process) {
	// Write the callable block to a file
	entrypoint := process.EnvRef.GetEntrypoint()
	if entrypoint == "" || process.OutputBlock.BlockLocation == "" {
		logger.Error("Missing configuration", "RUN_PROCESS", nil)
		panic("Missing configuration")
	}
	// To call this we need to pass in as follows:
	// entrypoint --func <block.location> --output <block.location> --blocks __pos__arg__0=<block.location> __pos__arg__1=<block.location> keyword1=<block.location> keyword2=<block.location> ...
	// We will start by writing the callable block
	funcCommand := " --func " + process.CallableBlock.BlockLocation
	// Write the output block
	outputCommand := " --output " + process.OutputBlock.BlockLocation
	// Write the arguments
	argumentsCommand := " "
	for i, arg := range process.ArgumentsBlocks {
		argumentsCommand = argumentsCommand + " __pos__arg__" + fmt.Sprint(i) + "=" + arg.BlockLocation
	}
	// Write the keyword arguments
	kwargsCommand := " "
	for key, kwarg := range process.KwargsBlocks {
		kwargsCommand = kwargsCommand + " " + key + "=" + kwarg.BlockLocation
	}

	// Combine the command
	fullCommand := entrypoint + funcCommand + outputCommand + " --blocks" + argumentsCommand + kwargsCommand
	fullCommand = NormalizeCommand(fullCommand)
	// Run the command
	logger.Info("Running process", "RUN_PROCESS", logrus.Fields{"command": fullCommand})
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", fullCommand)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		logger.Error("Error starting process", "RUN_PROCESS", err)
	}
	// Wait for the process to finish
	err = cmd.Wait()
	if err != nil {
		logger.Error("Error running process", "RUN_PROCESS", err)
		process.Success = false
	} else {
		process.Success = true
	}

	// We need to add the stdout and stderr to the process object
	WriteBytesToBlock(stdout.Bytes(), process.StdoutBlock)
	WriteBytesToBlock(stderr.Bytes(), process.StderrBlock)

	// Send the done signal
	doneChan <- true
	// TODO: Implement kill signal
}

func CleanupBinaries(process *Process) {
	logger.Info("Cleaning up process", "CleanupBinaries", logrus.Fields{"process_id": process.ProcessID})
}

func (p *Process) Run() {
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

func CleanBlocks(blocks map[string]*Block) {
	for _, block := range blocks {
		block.Cleanup()
	}
}

func ExitCleanAll(lm *LiveMemory) {
	// This function will clean up all the processes
	for _, process := range lm.Process {
		process.Cleanup()
	}
	// Cleanup the isolated environments
	for _, env := range lm.IsolatedEnvironment {
		env.Cleanup()
	}
	// Cleanup the blocks
	CleanBlocks(lm.Blocks)
}
