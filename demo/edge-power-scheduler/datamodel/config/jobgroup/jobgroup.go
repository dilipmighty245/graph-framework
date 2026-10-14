package jobgroup

import (
	"powerschedulermodel/config/jobgroup/jobscheduler"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

type Jobgroup struct {
	nexus.Node
	Jobs jobscheduler.Job `nexus:"children"`
}
