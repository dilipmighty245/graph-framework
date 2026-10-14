package main

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/cli/pkg/log"
	"github.com/dilipmighty245/graph-framework-for-microservices/cli/pkg/servicemesh/config"
	"github.com/dilipmighty245/graph-framework-for-microservices/cli/pkg/utils"
	"github.com/spf13/cobra"
)

func RootPreRun(cmd *cobra.Command, args []string) error {
	nexusConfig, err := config.LoadNexusConfig()
	if err != nil {
		return err
	}

	if nexusConfig.DebugAlways {
		cmd.Flags().Lookup("debug").Changed = true
	}

	// Set logging level to debug if debugging is enabled.
	if utils.IsDebug(cmd) {
		log.SetLevel(log.DebugLevel)
	}
	return nil
}
