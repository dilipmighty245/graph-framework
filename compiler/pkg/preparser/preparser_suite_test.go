package preparser

import (
	"testing"

	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/config"
	log "github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	exampleDSLPath = "../../example/test-utils/global-package"
	baseGroupName  = "nexus.com"
	crdModulePath  = "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/"
)

func TestPreparser(t *testing.T) {
	log.StandardLogger().ExitFunc = nil
	RegisterFailHandler(Fail)
	RunSpecs(t, "Preparser Suite")
}

func init() {
	conf := &config.Config{
		GroupName:     baseGroupName,
		CrdModulePath: crdModulePath,
	}
	config.ConfigInstance = conf
}
