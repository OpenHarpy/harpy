package main

import (
	"errors"
	"os/exec"
	"remote-runner/config"
	"remote-runner/logger"
)

func NodeSetup() error {
	logger.Info("Setting up node", "NODE_SETUP")
	// We will start by running the command
	setupScript := config.GetConfigs().GetConfigsWithDefault("enviroment_setup_script", "")
	if setupScript == "" {
		logger.Error("No setup script found", "NODE_SETUP", nil)
		return errors.New("no setup script found")
	}
	// We will now run the setup script
	cmd := exec.Command(setupScript)
	err := cmd.Start()
	if err != nil {
		logger.Error("Error running setup script", "NODE_SETUP", err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		logger.Error("Error running setup script", "NODE_SETUP", err)
		return err
	}
	logger.Info("Node setup complete", "NODE_SETUP")
	return nil
}
