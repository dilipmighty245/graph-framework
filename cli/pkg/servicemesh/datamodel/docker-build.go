package datamodel

import (
	"fmt"

	"github.com/dilipmighty245/graph-framework-for-microservices/cli/pkg/common"
	"github.com/dilipmighty245/graph-framework-for-microservices/cli/pkg/utils"
	"github.com/spf13/cobra"
)

var DockerRepo string
var Version string
var BuildDockerImgCmd = &cobra.Command{
	Use:   "dockerbuild",
	Short: "build docker image",
	//Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		return nil
	},
	RunE: BuildDockerImage,
}

func BuildDockerImage(cmd *cobra.Command, args []string) error {
	envList := common.GetEnvList()
	envList = append(envList, fmt.Sprintf("DOCKER_REPO=%s", DockerRepo))
	envList = append(envList, fmt.Sprintf("VERSION=%s", Version))

	err := utils.SystemCommand(cmd, utils.DATAMODEL_DOCKER_IMAGE_BUILD_FAILED, envList, "make", "docker_build")
	if err != nil {
		return fmt.Errorf("docker image %s build failed with error %v", DockerRepo, err)

	}
	fmt.Printf("\u2713 Datamodel docker image %s:%s built successfully\n", DockerRepo, Version)
	return nil
}

func init() {
	BuildDockerImgCmd.PersistentFlags().StringVarP(&DockerRepo, "docker-repo", "d", "", "docker repo name")
	BuildDockerImgCmd.PersistentFlags().StringVarP(&Version, "version", "v", "latest", "docker image version")
}
