package parser_test

import (
	"testing"

	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/config"
	log "github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	examplePath    = "../../example/"
	exampleDSLPath = examplePath + "datamodel"
	baseGroupName  = "nexus.com"
	crdModulePath  = "github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/output/generated/"
)

func TestParser(t *testing.T) {
	log.StandardLogger().ExitFunc = nil
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parser Suite")
}

func init() {
	conf := &config.Config{
		GroupName:     baseGroupName,
		CrdModulePath: crdModulePath,
		IgnoredDirs:   []string{"ignored"},
	}
	config.ConfigInstance = conf
}
