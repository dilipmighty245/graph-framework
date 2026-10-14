package main

import (
	"os"

	"github.com/dilipmighty245/graph-framework-for-microservices/cli/pkg/servicemesh"
	"github.com/dilipmighty245/graph-framework-for-microservices/cli/pkg/servicemesh/prereq"
	"github.com/dilipmighty245/graph-framework-for-microservices/cli/pkg/servicemesh/upgrade"
	"github.com/dilipmighty245/graph-framework-for-microservices/cli/pkg/utils"
	"github.com/spf13/cobra"
)

var enableDebug = false
var prereqShow = false
var skipPrereqCheck = false

var rootCmd = &cobra.Command{
	Use:               "nexus",
	Short:             "Nexus CLI",
	PersistentPreRunE: RootPreRun,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		servicemesh.ApplyCmd,
		servicemesh.DeleteCmd,
		servicemesh.LoginCmd,
		servicemesh.RuntimeCmd,
		servicemesh.DataModelCmd,
		servicemesh.AppCmd,
		servicemesh.OperatorCmd,
		servicemesh.VersionCmd,
		upgrade.UpgradeCmd,
		prereq.PreReqCmd,
		servicemesh.ConfigCmd,
		servicemesh.DebugCmd,
	)

	rootCmd.PersistentFlags().BoolVarP(&enableDebug, utils.EnableDebugFlag, "", false, "Enables extra logging")
	rootCmd.PersistentFlags().BoolVarP(&prereqShow, utils.ListPrereqFlag, "", false, "List prerequisites")
	rootCmd.PersistentFlags().BoolVarP(&skipPrereqCheck, utils.SkipPrereqCheckFlag, "", false, "Skip prerequisites check")
}
