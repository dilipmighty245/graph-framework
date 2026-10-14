package parser_test

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/config"
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/parser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pkg parser tests", func() {
	It("should parse dsl", func() {
		pkgs := parser.ParseDSLPkg(exampleDSLPath)
		_, ok := pkgs["github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/datamodel"]
		Expect(ok).To(BeTrue())
	})
	It("should ignore ignored dirs", func() {
		config.ConfigInstance.IgnoredDirs = []string{"ignored"}
		pkgs := parser.ParseDSLPkg(exampleDSLPath)
		ignored_imported := false
		for _, pkg := range pkgs {
			for _, f := range config.ConfigInstance.IgnoredDirs {
				if pkg.Name == f {
					ignored_imported = true
				}
			}
		}
		Expect(ignored_imported).To(BeFalse())
	})
})
