package rest

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/compiler/pkg/parser"
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

var HttpCodesResponsesMap = map[string]nexus.HTTPCodesResponse{
	"DefaultHTTPGETResponses":    nexus.DefaultHTTPGETResponses,
	"DefaultHTTPPUTResponses":    nexus.DefaultHTTPPUTResponses,
	"DefaultHTTPDELETEResponses": nexus.DefaultHTTPDELETEResponses,
}

var HttpMethodsResponsesMap = map[string]nexus.HTTPMethodsResponses{
	"DefaultHTTPMethodsResponses": nexus.DefaultHTTPMethodsResponses,
	"HTTPListResponse":            nexus.HTTPListResponse,
}

func ParseResponses(pkgs parser.Packages) (map[string]nexus.HTTPMethodsResponses, map[string]nexus.HTTPCodesResponse) {
	methods := make(map[string]nexus.HTTPMethodsResponses, len(HttpMethodsResponsesMap))
	for k, v := range HttpMethodsResponsesMap {
		methods[k] = v
	}
	codes := make(map[string]nexus.HTTPCodesResponse, len(HttpCodesResponsesMap))
	for k, v := range HttpCodesResponsesMap {
		codes[k] = v
	}

	// Iterate through packages to get all HTTP Codes
	for _, pkg := range pkgs {
		codes = GetHttpCodesResponses(pkg, codes)
	}

	// Iterate through packages to get all HTTP Methods
	for _, pkg := range pkgs {
		methods = GetHttpMethodsResponses(pkg, methods, codes)
	}

	return methods, codes
}
