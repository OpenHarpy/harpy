package main

import (
	"os"
	"os/exec"
	"remote-runner/logger"
)

func ExecCommandNoEcho(path string, args ...string) error {
	// We will start by running the command
	cmd := exec.Command(path)
	err := cmd.Start()
	if err != nil {
		logger.Error("error starting command", "EXEC_COMMAND_NO_ECHO", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logger.Error("error running command", "EXEC_COMMAND_NO_ECHO", err)
		return err
	}
	logger.Info("command executed", "EXEC_COMMAND_NO_ECHO")
	return nil
}

func ExecCommandEcho(path string, args ...string) error {
	// We will start by running the command
	cmd := exec.Command(path, args...)
	// Route the STDOUT and STDERR to the OS's STDOUT and STDERR
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		logger.Error("error starting command", "EXEC_COMMAND_ECHO", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logger.Error("error running command", "EXEC_COMMAND_ECHO", err)
		return err
	}
	logger.Info("command executed", "EXEC_COMMAND_ECHO")
	return nil
}
