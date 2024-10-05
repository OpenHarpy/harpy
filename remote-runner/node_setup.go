package main

import (
	"os"
	"os/exec"
	"remote-runner/config"
	"remote-runner/logger"
)

func NodeSetup() {
	logger.Info("Setting up node", "NODE_SETUP")
	// We will start by running the command
	setupScript := config.GetConfigs().GetConfigsWithDefault("enviroment_setup_script", "")
	if setupScript == "" {
		logger.Error("No setup script found", "NODE_SETUP", nil)
		os.Exit(1)
	}
	// We will now run the setup script
	cmd := exec.Command(setupScript)
	err := cmd.Start()
	if err != nil {
		logger.Error("Error running setup script", "NODE_SETUP", err)
		os.Exit(1)
	}
	err = cmd.Wait()
	if err != nil {
		logger.Error("Error running setup script", "NODE_SETUP", err)
		os.Exit(1)
	}
	logger.Info("Node setup complete", "NODE_SETUP")

}
