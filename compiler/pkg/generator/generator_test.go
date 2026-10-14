package generator_test

import (
	"encoding/json"
	"go/format"
	"os"

	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/util"
	"gopkg.in/yaml.v2"

	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/parser/rest"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/generator"
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/parser"
)

const (
	baseGroupName              = "nexus.com"
	crdModulePath              = "nexustempmodule/"
	examplePath                = "../../example/"
	exampleDSLPath             = examplePath + "datamodel"
	gnsDocPath                 = "testdata/gns_doc.go.golden"
	gnsRegisterGroupPath       = "testdata/gns_register_group.go.golden"
	gnsRegisterCRDPath         = "testdata/gns_register_crd.go.golden"
	gnsTypesPath               = "testdata/gns_types.go.golden"
	gnsCrdBasePath             = "testdata/gns_gns.yaml.golden"
	exampleCRDClientOutputPath = "testdata/client.go.golden"
	exampleCRDHelperOutputPath = "testdata/helper.go.golden"
)

var _ = Describe("Template renderers tests", func() {
	var (
		pkgs       parser.Packages
		pkg        parser.Package
		parentsMap map[string]parser.NodeHelper
		ok         bool
		methods    map[string]nexus.HTTPMethodsResponses
		codes      map[string]nexus.HTTPCodesResponse
		gql        generator.GraphDetails
	)

	BeforeEach(func() {
		pkgs = parser.ParseDSLPkg(exampleDSLPath)
		pkg, ok = pkgs["github.com/dilipmighty245/graph-framework-for-microservices/compiler/example/datamodel/config/gns"]
		Expect(ok).To(BeTrue())
		graphqlQueries := parser.ParseGraphqlQuerySpecs(pkgs)
		graph, _, _ := parser.ParseDSLNodes(exampleDSLPath, baseGroupName, pkgs, graphqlQueries)
		parentsMap = parser.CreateParentsMap(graph)
		Expect(parentsMap).To(HaveLen(15))

		methods, codes = rest.ParseResponses(pkgs)
		gql.BaseImportPath = crdModulePath
		var err error
		gql.Nodes, err = generator.GenerateGraphqlResolverVars(baseGroupName, crdModulePath, pkgs, parentsMap)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should parse doc template", func() {
		docBytes, err := generator.RenderDocTemplate(baseGroupName, pkg)
		Expect(err).NotTo(HaveOccurred())
		formatted, err := format.Source(docBytes.Bytes())
		Expect(err).NotTo(HaveOccurred())

		expectedDoc, err := os.ReadFile(gnsDocPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(formatted)).To(Equal(string(expectedDoc)))
	})

	It("should parse register group template", func() {
		regBytes, err := generator.RenderRegisterGroupTemplate(baseGroupName, pkg)
		Expect(err).NotTo(HaveOccurred())
		formatted, err := format.Source(regBytes.Bytes())
		Expect(err).NotTo(HaveOccurred())

		expectedRegisterGroup, err := os.ReadFile(gnsRegisterGroupPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(formatted)).To(Equal(string(expectedRegisterGroup)))
	})

	It("should parse register CRD template", func() {
		regBytes, err := generator.RenderRegisterCRDTemplate(crdModulePath, baseGroupName, pkg)
		Expect(err).NotTo(HaveOccurred())
		formatted, err := format.Source(regBytes.Bytes())
		Expect(err).NotTo(HaveOccurred())

		expectedRegisterCRD, err := os.ReadFile(gnsRegisterCRDPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(formatted)).To(Equal(string(expectedRegisterCRD)))
	})

	It("should parse base crd template", func() {
		files, err := generator.RenderCRDBaseTemplate(baseGroupName, pkg, parentsMap, methods, codes)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(5))

		expectedSdk, err := os.ReadFile(gnsCrdBasePath)
		Expect(err).NotTo(HaveOccurred())

		Expect("gns_gns.yaml").To(Or(Equal(files[1].Name), Equal(files[2].Name)))
		Expect(string(expectedSdk)).To(Or(Equal(files[1].File.String()), Equal(files[2].File.String())))
	})

	When("nexus-deferred-delete annotation is added in the DSL", func() {
		It("should set deferred-delete to true in the nexus annotation of the CRD", func() {
			data, err := os.ReadFile(gnsCrdBasePath)
			Expect(err).NotTo(HaveOccurred())

			var crd map[string]interface{}
			Expect(yaml.Unmarshal(data, &crd)).To(Succeed())

			nexus := crd["metadata"].(map[interface{}]interface{})["annotations"].(map[interface{}]interface{})["nexus"].(string)

			var nexusData map[string]interface{}
			Expect(json.Unmarshal([]byte(nexus), &nexusData)).To(Succeed())
			Expect(nexusData["deferred-delete"]).To(BeTrue())
		})
	})

	It("should parse types template", func() {
		typesBytes, err := generator.RenderTypesTemplate(crdModulePath, pkg)
		Expect(err).NotTo(HaveOccurred())

		formatted, err := format.Source(typesBytes.Bytes())
		Expect(err).NotTo(HaveOccurred())

		expectedTypes, err := os.ReadFile(gnsTypesPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(formatted)).To(Equal(string(expectedTypes)))
	})

	It("should parse client template", func() {
		clientsBytes, err := generator.RenderClientTemplate(baseGroupName, crdModulePath, pkgs, parentsMap)
		Expect(err).NotTo(HaveOccurred())

		formatted, err := format.Source(clientsBytes.Bytes())
		Expect(err).NotTo(HaveOccurred())

		expectedTypes, err := os.ReadFile(exampleCRDClientOutputPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(formatted)).To(Equal(string(expectedTypes)))
	})

	It("should parse helper template", func() {
		helperBytes, err := generator.RenderHelperTemplate(parentsMap, crdModulePath)
		Expect(err).NotTo(HaveOccurred())

		formatted, err := format.Source(helperBytes.Bytes())
		Expect(err).NotTo(HaveOccurred())

		expectedTypes, err := os.ReadFile(exampleCRDHelperOutputPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(formatted)).To(Equal(string(expectedTypes)))
	})

	It("should parse nexus-gql server template", func() {
		var vars generator.ServerVars
		vars.BaseImportPath = crdModulePath
		_, err := generator.RenderGqlServerTemplate(vars)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should parse graph schema graphql template", func() {
		_, err := generator.RenderGraphqlSchemaTemplate(gql, crdModulePath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should parse nexus-gql gqlgen config template", func() {
		_, err := generator.RenderGQLGenTemplate(gql, crdModulePath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should parse graphql resolver template", func() {
		_, err := generator.RenderGraphqlResolverTemplate(gql, crdModulePath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should test GetPackageName ", func() {
		expectedPackageName := "nexus_demo_com"

		groupName := "nexus-demo.com"
		packageName := util.GetPackageName(groupName)
		Expect(packageName).To(Equal(expectedPackageName))

		groupName = "nexus.demo-com"
		packageName = util.GetPackageName(groupName)
		Expect(packageName).To(Equal(expectedPackageName))
	})

	It("should test GetGroupGoName", func() {
		expectedGroupGoName := "Nexus"
		groupName := "nexus-demo.com"
		packageName := util.GetGroupGoName(groupName)
		Expect(packageName).To(Equal(expectedGroupGoName))
	})

	It("should handle hyphen in group-name", func() {
		datamodelPath := "../../example/test-utils/group-name-with-hyphen-datamodel"
		groupName := "nexus-demo.com"
		crdModulePath := "nexustempmodule/"
		outputDir, err := os.MkdirTemp("", "nexus-hyphen-model-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(outputDir)

		pkgs := parser.ParseDSLPkg(datamodelPath)
		graphlqQueries := parser.ParseGraphqlQuerySpecs(pkgs)
		graph, nonNexusTypes, fileset := parser.ParseDSLNodes(datamodelPath, groupName, pkgs, graphlqQueries)
		methods, codes := rest.ParseResponses(pkgs)
		err = generator.RenderCRDTemplate(groupName, crdModulePath, pkgs, graph, outputDir, methods, codes, nonNexusTypes, fileset, nil)
		Expect(err).NotTo(HaveOccurred())
	})
})
