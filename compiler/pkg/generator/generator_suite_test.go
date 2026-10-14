package generator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/config"
)

func TestGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generator Suite")
}

func init() {
	conf := &config.Config{
		GroupName:     baseGroupName,
		CrdModulePath: crdModulePath,
	}
	config.ConfigInstance = conf
}
