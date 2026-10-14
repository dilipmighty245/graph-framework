package config

import (
	"powerschedulermodel/config/jobgroup"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

type Config struct {
	nexus.SingletonNode
	JobGroups jobgroup.Jobgroup        `nexus:"children"`
	Scheduler jobgroup.SchedulerConfig `nexus:"child"`
}
