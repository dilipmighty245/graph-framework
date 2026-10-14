package echo_server

import (
	"api-gw/pkg/client"
	"api-gw/pkg/common"
	"api-gw/pkg/config"
	"api-gw/pkg/model"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	labelSelector "k8s.io/apimachinery/pkg/labels"

	"github.com/dilipmighty245/graph-framework-for-microservices/common-library/pkg/nexus"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// kubeSetupProxy is used to set up reverse proxy to an API server
func kubeSetupProxy(e *echo.Echo) *httputil.ReverseProxy {
	proxyUrl, err := url.Parse(client.Host)
	if err != nil {
		log.Warnf("Could not parse proxy URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(proxyUrl)
	if client.HostScheme == "https" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(client.HostTLSClientConfig.CAData)
		cert, err := tls.LoadX509KeyPair(client.HostTLSClientConfig.CertFile, client.HostTLSClientConfig.KeyFile)
		if err != nil {
			log.Warnf("Could not load client certficate: %y+v", err)
		}
		httpTransport := http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				MaxVersion:         tls.VersionTLS13,
				RootCAs:            caCertPool,
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true,
			},
		}
		proxy.Transport = &httpTransport
	}
	proxy.ModifyResponse = UpdateProxyResponse
	if common.IsModeAdmin() {
		e.Any("/api/*", echo.WrapHandler(proxy))
		e.Any("/apis/*", echo.WrapHandler(proxy))
		e.Any("/api", echo.WrapHandler(proxy))
		e.Any("/apis", echo.WrapHandler(proxy))
		e.Any("/readyz", echo.WrapHandler(proxy))
		e.Any("/openapi/*", echo.WrapHandler(proxy))
		e.Any("/openapi", echo.WrapHandler(proxy))
		e.Any("/healthz", echo.WrapHandler(proxy))
		e.Any("/readyz", echo.WrapHandler(proxy))
	} else {
		e.Any("/*", echo.WrapHandler(proxy))
	}
	return proxy
}

func UpdateProxyResponse(response *http.Response) error {
	if config.Cfg.CustomNotFoundPage != "" &&
		(response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusMovedPermanently) {
		resp, err := http.Get(config.Cfg.CustomNotFoundPage)
		if err != nil {
			log.Errorf("Proxy modify response error: %v", err)
			return nil
		}

		response.Body = resp.Body
		response.Header = resp.Header
		response.StatusCode = resp.StatusCode
		return nil
	}
	return nil
}

// kubeGetByNameHandler is used to process 'kubectl get <resource> <name>' requests
func KubeGetByNameHandler(c echo.Context) error {
	nc := c.(*NexusContext)

	gvr := schema.GroupVersionResource{
		Group:    nc.GroupName,
		Version:  "v1",
		Resource: nc.Resource,
	}
	obj, err := client.GetObject(gvr, c.Param("name"), metav1.GetOptions{})
	if err != nil {
		if status := kerrors.APIStatus(nil); errors.As(err, &status) {
			return c.JSON(int(status.Status().Code), status.Status())
		}
		c.Error(err)
	}

	return c.JSON(200, obj)
}

// kubeGetHandler is used to process `kubectl get <resource>' requests
func KubeGetHandler(c echo.Context) error {
	nc := c.(*NexusContext)

	opts := metav1.ListOptions{}
	if c.QueryParams().Has("labelSelector") {
		opts.LabelSelector = c.QueryParams().Get("labelSelector")
	}

	if c.QueryParams().Has("limit") {
		i, err := strconv.ParseInt(c.QueryParams().Get("limit"), 10, 64)
		if err != nil {
			return err
		}
		opts.Limit = i
	}

	if c.QueryParams().Has("continue") {
		opts.Continue = c.QueryParams().Get("continue")
	}

	gvr := schema.GroupVersionResource{
		Group:    nc.GroupName,
		Version:  "v1",
		Resource: nc.Resource,
	}
	log.Debugf("KubeGetHandler: received GET rquest for %+v", gvr)

	obj, err := client.Client.Resource(gvr).List(context.TODO(), opts)
	if err != nil {
		log.Debugf("KubeGetHandler: GetObject for %+v failed with error %+v", gvr, err)
		if status := kerrors.APIStatus(nil); errors.As(err, &status) {
			return c.JSON(int(status.Status().Code), status.Status())
		}
		c.Error(err)
	}
	return c.JSON(200, obj)
}

func processBody(body *unstructured.Unstructured, nc *NexusContext, crdInfo model.NodeInfo) (*unstructured.Unstructured, map[string]string, string, string) {
	displayName := body.GetName()
	labels := body.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["nexus/is_name_hashed"] = "true"
	labels["nexus/display_name"] = displayName

	orderedLabels := nexus.ParseCRDLabels(crdInfo.ParentHierarchy, labels)
	for _, key := range orderedLabels.Keys() {
		value, _ := orderedLabels.Get(key)
		labels[key.(string)] = value.(string)
	}

	hashedName := nexus.GetHashedName(nc.CrdType, crdInfo.ParentHierarchy, labels, displayName)
	body.SetLabels(labels)
	body.SetName(hashedName)

	if crdInfo.DeferredDelete {
		finalizerVal := "nexus.com/nexus-deferred-delete"
		body.SetFinalizers([]string{finalizerVal})
		log.Debugf("Object %s is marked for deferred delete, added finalizer %v", body.GetName(), body.GetFinalizers())
	}

	return body, labels, hashedName, displayName
}

// KubePostHandler is used to process `kubectl apply` requests
func KubePostHandler(c echo.Context) error {
	nc := c.(*NexusContext)
	crdInfo := model.CrdTypeToNodeInfo[nc.CrdType]

	body := &unstructured.Unstructured{}
	if err := c.Bind(&body); err != nil {
		return err
	}

	body, labels, hashedName, _ := processBody(body, nc, crdInfo)

	gvr := schema.GroupVersionResource{
		Group:    nc.GroupName,
		Version:  "v1",
		Resource: nc.Resource,
	}
	log.Debugf("KubePostHandler: received POST request for %+v", gvr)

	// Get object to check if it exists
	obj, err := client.GetObject(gvr, hashedName, metav1.GetOptions{})
	if err != nil {

		// Create object if is not found
		if kerrors.IsNotFound(err) {

			if len(crdInfo.ParentHierarchy) > 0 {
				parentCrdName := crdInfo.ParentHierarchy[len(crdInfo.ParentHierarchy)-1]
				parentCrd := model.CrdTypeToNodeInfo[parentCrdName]
				if _, err := client.GetParent(parentCrdName, parentCrd, labels); err != nil {
					log.Debugf("KubePostHandler: GetParent failed with error %+v", err)
					if status := kerrors.APIStatus(nil); errors.As(err, &status) {
						return c.String(int(status.Status().Code), fmt.Sprintf("parent object lookup failed with error: %v", status.Status()))
					}
				}
			}

			if _, ok := body.UnstructuredContent()["spec"]; !ok {
				content := body.UnstructuredContent()
				content["spec"] = map[string]interface{}{}
				body.SetUnstructuredContent(content)
			}
			obj, err = client.Client.Resource(gvr).Create(context.TODO(), body, metav1.CreateOptions{})
			if err != nil {
				if status := kerrors.APIStatus(nil); errors.As(err, &status) {
					return c.JSON(int(status.Status().Code), status.Status())
				}
				c.Error(err)
			}

			// var err error
			//if len(crdInfo.ParentHierarchy) > 0 {
			//	parentCrdName := crdInfo.ParentHierarchy[len(crdInfo.ParentHierarchy)-1]
			//	parentCrd := model.CrdTypeToNodeInfo[parentCrdName]
			//	err = client.UpdateParentWithAddedChild(parentCrdName, parentCrd, labels, crdInfo, nc.CrdType, displayName, hashedName)
			//}

			//if err != nil {
			//	if status := kerrors.APIStatus(nil); errors.As(err, &status) {
			//		return c.JSON(int(status.Status().Code), status.Status())
			//	}
			//	c.Error(err)
			//}

			return c.JSON(201, obj)
		}

		if status := kerrors.APIStatus(nil); errors.As(err, &status) {
			return c.JSON(int(status.Status().Code), status.Status())
		}
		c.Error(err)
	} else {
		log.Debugf("KubePostHandler: GetObject for %+v name %s failed with error %+v", gvr, hashedName, err)
		if status := kerrors.APIStatus(nil); errors.As(err, &status) {
			return c.String(int(status.Status().Code), fmt.Sprintf("object lookup for %+v name %s failed with error: %v", gvr, hashedName, status.Status()))
		}
	}

	body.SetResourceVersion(obj.GetResourceVersion())
	spec := obj.Object["spec"].(map[string]interface{})
	newSpec := body.Object["spec"].(map[string]interface{})
	for _, v := range crdInfo.Children {
		if value, ok := spec[v.FieldNameGvk]; ok {
			newSpec[v.FieldNameGvk] = value
		}
	}
	for _, v := range crdInfo.Links {
		if value, ok := spec[v.FieldNameGvk]; ok {
			newSpec[v.FieldNameGvk] = value
		}
	}
	body.Object["spec"] = newSpec
	obj, err = client.Client.Resource(gvr).Update(context.TODO(), body, metav1.UpdateOptions{})
	if err != nil {
		if status := kerrors.APIStatus(nil); errors.As(err, &status) {
			return c.JSON(int(status.Status().Code), status.Status())
		}
		c.Error(err)
	}

	return c.JSON(200, obj)
}

func KubeDeleteHandler(c echo.Context) error {
	nc := c.(*NexusContext)
	crdInfo := model.CrdTypeToNodeInfo[nc.CrdType]
	gvr := schema.GroupVersionResource{
		Group:    nc.GroupName,
		Version:  "v1",
		Resource: nc.Resource,
	}
	labels := make(map[string]string)
	name := c.Param("name")
	log.Debugf("KubeDeleteHandler: display name: %s", name)

	if c.QueryParams().Has("labelSelector") {
		labelsMap, err := labelSelector.ConvertSelectorToLabelsMap(c.QueryParams().Get("labelSelector"))
		if err != nil {
			return err
		}
		for key, val := range labelsMap {
			labels[key] = val
		}
	}

	// Kubernetes clients use stored names; graph clients can supply display names.
	if _, err := client.Client.Resource(gvr).Get(c.Request().Context(), name, metav1.GetOptions{}); kerrors.IsNotFound(err) {
		name = nexus.GetHashedName(nc.CrdType, crdInfo.ParentHierarchy, labels, name)
	} else if err != nil {
		return err
	}
	log.Debugf("KubeDeleteHandler: hashedName: %s, labels: %s", name, labels)

	err := client.DeleteObject(gvr, nc.CrdType, crdInfo, name)
	if err != nil {
		if status := kerrors.APIStatus(nil); errors.As(err, &status) {
			response := status.Status()
			response.Message = strings.ReplaceAll(response.Message, name, c.Param("name"))
			if response.Details != nil {
				response.Details.Name = c.Param("name")
			}
			return c.JSON(int(response.Code), response)
		}
		c.Error(err)
	}

	return c.JSON(200, map[string]interface{}{
		"kind":       "Status",
		"apiVersion": "v1",
		"metadata":   map[string]interface{}{},
		"status":     "Success",
		"details": map[string]interface{}{
			"name":  c.Param("name"),
			"group": nc.GroupName,
			"kind":  nc.Resource,
		},
	})
}
