package main

import (
	"remote-runner/config"
	"remote-runner/logger"
)

func NodeSetup() error {
	logger.Info("Setting up node", "NODE_SETUP")
	// We will start by running the command
	root := config.GetConfigs().GetConfigsWithDefault("harpy.remoteRunner.scriptsRoot", "")
	setupScript := config.GetConfigs().GetConfigsWithDefault("harpy.remoteRunner.nodeSetupScript", "")
	fullPath := root + "/" + setupScript
	err := ExecCommandNoEcho(fullPath)
	if err != nil {
		logger.Error("error running node setup script", "NODE_SETUP", err)
		return err
	}
	logger.Info("node setup script executed", "NODE_SETUP")
	return nil
}
