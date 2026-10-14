package jobmgmt

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

var JobInfoRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/v1alpha1/info/site/{sitedc.SiteDC}/edge/{edgedc.EdgeDC}/jobs/{jobmgmt.JobInfo}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/v1alpha1/info/site/{sitedc.SiteDC}edge/{edgedc.EdgeDC}/jobs",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:JobInfoRestAPISpec
// nexus-description: This is for adding jobs to an edge
type JobInfo struct {
	nexus.Node
	RequestorJob      string
	RequestorJobGroup string
	PowerNeeded       uint32    // amount of power needed Unit : watt-seconds.
	State             JobStatus `nexus:"status"`
}

type JobStatus struct {
	Progress  uint32 // Percent completed
	StartTime int64  // Time when job was started.
	EndTime   int64  // time when updates where made. when percent is 100 this will be the end time.
}
