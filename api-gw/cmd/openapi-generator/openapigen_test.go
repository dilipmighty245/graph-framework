package main

import (
	"api-gw/pkg/utils"
	"bytes"
	"io"
	"log"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

const (
	DATAMODEL_CONFIG_FILE       string = "test/nexus.yaml"
	EXPECTED_OPENAPI_SPEC_FILE  string = "test/openapispec/generated_file.json"
	GENERATED_OPENAPI_SPEC_FILE string = "test/_generated/generated.json"
	CRDS_DIR                    string = "test/crds"
	OPENAPI_SCHEMA_NAME         string = "test"
)

func TestGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenApi Generator Suite")
}

var _ = Describe("Generator", func() {
	BeforeEach(func() { utils.OpenApiIgnoredParentPathParams = make(map[string]struct{}) })
	It("will not generate openapi spec as expected if optional path params are not configured properly", func() {
		run(CRDS_DIR, OPENAPI_SCHEMA_NAME, GENERATED_OPENAPI_SPEC_FILE)
		expectedSpec, err := os.ReadFile(EXPECTED_OPENAPI_SPEC_FILE)
		if err != nil {
			log.Fatal(err)
		}
		actualSpec, err := os.ReadFile(GENERATED_OPENAPI_SPEC_FILE)
		if err != nil {
			log.Fatal(err)
		}
		Expect(bytes.Equal(expectedSpec, actualSpec)).Should(BeFalse())
	})

	It("will generate openapi spec as expected with optional params configured correctly", func() {
		utils.InitOpenApiIgnoredParentPathParams(DATAMODEL_CONFIG_FILE)
		run(CRDS_DIR, "test", GENERATED_OPENAPI_SPEC_FILE)

		expectedSpec, err := os.ReadFile(EXPECTED_OPENAPI_SPEC_FILE)
		if err != nil {
			log.Fatal(err)
		}
		actualSpec, err := os.ReadFile(GENERATED_OPENAPI_SPEC_FILE)
		if err != nil {
			log.Fatal(err)
		}
		Expect(bytes.Equal(expectedSpec, actualSpec)).Should(BeTrue())
	})
})

var _ = Describe("Optional parent params", func() {
	It("are initialized as expected", func() {
		utils.InitOpenApiIgnoredParentPathParams(DATAMODEL_CONFIG_FILE)

		var config utils.DatamodelConfig
		file, err := os.Open(DATAMODEL_CONFIG_FILE)
		Expect(err).To(BeNil())

		configStr, err := io.ReadAll(file)
		Expect(err).To(BeNil())

		err = yaml.Unmarshal(configStr, &config)
		if err != nil {
			log.Fatalf("failed to unmarshal config file %s with error %s", DATAMODEL_CONFIG_FILE, err)
		}

		for _, param := range config.IgnoredParentPathParams {
			Expect(utils.OpenApiIgnoredParentPathParams).To(HaveKey(param))
		}
	})
})
