package rest_test

import (
	"testing"

	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/config"
	log "github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	examplePath    = "../../../example/"
	exampleDSLPath = examplePath + "datamodel"
	baseGroupName  = "tsm.tanzu.vmware.com"
	crdModulePath  = "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/"
)

func TestParserRest(t *testing.T) {
	log.StandardLogger().ExitFunc = nil
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parser Rest Suite")
}

func init() {
	conf := &config.Config{
		GroupName:     baseGroupName,
		CrdModulePath: crdModulePath,
	}
	config.ConfigInstance = conf
}
