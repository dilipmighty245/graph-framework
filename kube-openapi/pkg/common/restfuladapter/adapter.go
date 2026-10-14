package restfuladapter

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/kube-openapi/pkg/common"
	"github.com/emicklei/go-restful"
)

// AdaptWebServices adapts a slice of restful.WebService into the common interfaces.
func AdaptWebServices(webServices []*restful.WebService) []common.RouteContainer {
	var containers []common.RouteContainer
	for _, ws := range webServices {
		containers = append(containers, &WebServiceAdapter{ws})
	}
	return containers
}
