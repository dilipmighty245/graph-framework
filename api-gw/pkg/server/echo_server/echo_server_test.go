package echo_server_test

import (
	"api-gw/pkg/client"
	"api-gw/pkg/config"
	"api-gw/pkg/model"
	"api-gw/pkg/server/echo_server"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
	"github.com/labstack/echo/v4"

	apinexusv1 "nexus/admin/api/build/apis/api.admin.nexus.com/v1"
	confignexusv1 "nexus/admin/api/build/apis/config.admin.nexus.com/v1"
	domain_nexus_org "nexus/admin/api/build/apis/domain.admin.nexus.com/v1"
	nexus_client "nexus/admin/api/build/nexus-client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("Echo server tests", func() {
	var e *echo_server.EchoServer

	for _, operation := range []string{"creation", "update", "deletion"} {
		operation := operation
		It("should watch for OpenApi spec file "+operation, func() {
			dir, err := os.MkdirTemp("", "nexus-openapi-watch-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(dir)
			file := filepath.Join(dir, "schema.json")
			if operation != "creation" {
				Expect(os.WriteFile(file, []byte("before"), 0600)).To(Succeed())
			}
			stopCh := make(chan struct{}, 16)
			stop := echo_server.WatchForOpenApiSpecChanges(stopCh, dir, file)
			defer stop()
			if operation == "deletion" {
				Expect(os.Remove(file)).To(Succeed())
			} else {
				Expect(os.WriteFile(file, []byte("after"), 0600)).To(Succeed())
			}
			Eventually(stopCh, 3*time.Second).Should(Receive())
		})
	}

	BeforeEach(func() {

		client.NexusClient = nexus_client.NewFakeClient()
		root, err := client.NexusClient.AddApiNexus(context.Background(), &apinexusv1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: "default"}})
		Expect(err).NotTo(HaveOccurred())
		_, err = root.AddConfig(context.Background(), &confignexusv1.Config{ObjectMeta: metav1.ObjectMeta{Name: "default"}})
		Expect(err).NotTo(HaveOccurred())

		config.Cfg = &config.Config{
			Server:             config.ServerConfig{},
			EnableNexusRuntime: true,
			BackendService:     "",
		}
		e = echo_server.NewEchoServer(config.Cfg, &kubernetes.Clientset{}, client.NexusClient)

	})

	It("should register CrdRouter for globalnamespaces.gns.vmware.org", func() {
		e.RegisterCrdRouter("globalnamespaces.gns.vmware.org")

		c := e.Echo.NewContext(nil, nil)
		e.Echo.Router().Find(http.MethodGet, "/apis/gns.vmware.org/v1/globalnamespaces/:name", c)
		Expect(c.Path()).To(Equal("/apis/gns.vmware.org/v1/globalnamespaces/:name"))

		c = e.Echo.NewContext(nil, nil)
		e.Echo.Router().Find(http.MethodGet, "/apis/gns.vmware.org/v1/globalnamespaces", c)
		Expect(c.Path()).To(Equal("/apis/gns.vmware.org/v1/globalnamespaces"))

		c = e.Echo.NewContext(nil, nil)
		e.Echo.Router().Find(http.MethodPost, "/apis/gns.vmware.org/v1/globalnamespaces", c)
		Expect(c.Path()).To(Equal("/apis/gns.vmware.org/v1/globalnamespaces"))

		c = e.Echo.NewContext(nil, nil)
		e.Echo.Router().Find(http.MethodDelete, "/apis/gns.vmware.org/v1/globalnamespaces/:name", c)
		Expect(c.Path()).To(Equal("/apis/gns.vmware.org/v1/globalnamespaces/:name"))

	})

	It("should return true when origin is present", func() {
		model.CorsConfigOrigins["default"] = []string{"http://testdomain"}

		res, err := echo_server.CheckCorsOrigin("http://localhost")
		Expect(res).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())

		res, err = echo_server.CheckCorsOrigin("http://testdomain")
		Expect(res).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
		//removing the added object
		delete(model.CorsConfigOrigins, "default")

		res, err = echo_server.CheckCorsOrigin("http://testdomain")
		Expect(res).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())

	})

	It("should register cors middleware when event is recieved", func() {

		e.RegisterCrdRouter("globalnamespaces.gns.vmware.org")
		corsObject := domain_nexus_org.CORSConfig{
			Spec: domain_nexus_org.CORSConfigSpec{
				Origins: []string{"http://tester"},
				Headers: []string{""},
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "test",
			},
		}

		//check if corsOrigin is setup
		corsUpdate := &model.CorsNodeEvent{Cors: corsObject, Type: model.Upsert}
		echo_server.HandleCorsNodeUpdate(corsUpdate, e.Echo)
		Eventually(func() bool {
			res, _ := echo_server.CheckCorsOrigin("http://tester")
			return res
		})

		//delete should remove the node
		corsUpdate = &model.CorsNodeEvent{Cors: corsObject, Type: model.Delete}
		delete(model.CorsConfigOrigins, "test")
		echo_server.HandleCorsNodeUpdate(corsUpdate, e.Echo)
		Eventually(func() bool {
			res, _ := echo_server.CheckCorsOrigin("http://tester")
			if res {
				return false
			}
			return true
		})

		err := echo_server.HandleCorsNodeUpdate(nil, e.Echo)
		Expect(err).To(HaveOccurred())

	})

	It("should start echo server", func() {
		stopCh := make(chan struct{})
		e := echo_server.InitEcho(stopCh, &config.Config{
			Server: config.ServerConfig{
				HttpPort: "0",
			},
			EnableNexusRuntime: true,
			BackendService:     "http://localhost",
		}, &kubernetes.Clientset{},
			&nexus_client.Clientset{})
		e.StopServer()
	})

	It("should start echo server and restart through channel", func() {
		stopCh := make(chan struct{})
		e := echo_server.InitEcho(stopCh, &config.Config{
			Server: config.ServerConfig{
				HttpPort: "0",
			},
			EnableNexusRuntime: true,
			BackendService:     "http://localhost",
		}, &kubernetes.Clientset{},
			&nexus_client.Clientset{})

		before := atomic.LoadInt64(&echo_server.TotalHttpServerRestartCounter)
		stopCh <- struct{}{}
		Eventually(func() int64 { return atomic.LoadInt64(&echo_server.TotalHttpServerRestartCounter) }, 3*time.Second).Should(BeNumerically(">", before))

		e.StopServer()
	})

	It("should get nexus crd context", func() {
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.Echo.NewContext(req, rec)
		nexusCrdContext := e.GetNexusCrdContext("test.vmware.org", "vmware.org", "tests")

		var actualContext *echo_server.NexusContext
		c.SetHandler(
			nexusCrdContext(
				func(c echo.Context) error {
					actualContext = c.(*echo_server.NexusContext)
					return c.NoContent(200)
				},
			),
		)
		err := c.Handler()(c)
		Expect(err).NotTo(HaveOccurred())

		Expect(actualContext.Resource).To(Equal("tests"))
		Expect(actualContext.GroupName).To(Equal("vmware.org"))
		Expect(actualContext.CrdType).To(Equal("test.vmware.org"))
	})

	It("should get nexus context", func() {
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.Echo.NewContext(req, rec)
		restUri := nexus.RestURIs{
			Uri:     "/test",
			Methods: nexus.DefaultHTTPMethodsResponses,
		}
		codes := map[nexus.ResponseCode]nexus.HTTPResponse{
			http.StatusOK: {
				Description: "description",
			},
		}

		nexusCrdContext := e.GetNexusContext(restUri, codes)
		var actualContext *echo_server.NexusContext
		c.SetHandler(
			nexusCrdContext(
				func(c echo.Context) error {
					actualContext = c.(*echo_server.NexusContext)
					return c.NoContent(200)
				},
			),
		)
		err := c.Handler()(c)
		Expect(err).NotTo(HaveOccurred())

		Expect(actualContext.NexusURI).To(Equal("/test"))
		Expect(actualContext.Codes[http.StatusOK].Description).To(Equal("description"))
	})

	It("should register TSM Routes", func() {
		e.RegisterCosmosAdminRoutes()

		c := e.Echo.NewContext(nil, nil)
		e.Echo.Router().Find(http.MethodPut, "/v0/tenants/instance", c)
		Expect(c.Path()).To(Equal("/v0/tenants/instance"))

		c = e.Echo.NewContext(nil, nil)
		e.Echo.Router().Find(http.MethodGet, "/v0/version", c)
		Expect(c.Path()).To(Equal("/v0/version"))

		c = e.Echo.NewContext(nil, nil)
		e.Echo.Router().Find(http.MethodPost, "/v0/users/login", c)
		Expect(c.Path()).To(Equal("/v0/users/login"))
	})

	It("should stop cleanly before the asynchronous listener starts", func() {
		server := echo_server.InitEcho(make(chan struct{}), &config.Config{Server: config.ServerConfig{HttpPort: "0"}}, &kubernetes.Clientset{}, &nexus_client.Clientset{})
		server.StopServer()
		server.StopServer() // Closing an already stopped server is safe.
	})

})
