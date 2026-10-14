package app

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"

	handler "connector/pkg/handler"

	"connector/controllers"
	"connector/pkg/config"
	"connector/pkg/handlers"
	"connector/pkg/utils"
)

func Start(cache *controllers.GvrCache) {
	conf, err := config.LoadConfig("/config/connector-config")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	for {
		stopCh := make(chan struct{})
		go StartControllers(conf, cache, stopCh)
		go StartReplicationConfigController(conf, stopCh)
		<-stopCh
		log.Info("Stop signal received, restarting crd watcher")
	}
}

func StartControllers(conf *config.Config, cache *controllers.GvrCache, stopCh chan struct{}) {
	log.Infoln("Starting Controllers...")
	_, localDynamicAPI, err := utils.SetUpAPIs()
	if err != nil {
		log.Fatalf("Error creating APIs: %v", err)
		return
	}

	var remoteDynamicClient dynamic.Interface
	if conf.StatusReplicationEnabled {
		remoteDynamicClient, err = utils.BuildRemoteClientAPI(conf.RemoteEndpointHost, conf.RemoteEndpointPort, conf.RemoteEndpointCert, localDynamicAPI)
		if err != nil {
			log.Fatalf("Error creating remote-client APIs: %v", err)
			return
		}
	}

	for {
		select {
		case <-stopCh:
			log.Errorln("StartController() terminated..")
			return
		case gvr := <-controllers.GvrCh:
			// Skipping watcher creation for nexus datamodel CRDs.
			if utils.NexusDatamodelCRDs(gvr.Group) {
				continue
			}

			informer := getInformer(gvr)
			handler := handlers.NewRemoteHandler(gvr, localDynamicAPI, remoteDynamicClient, conf)
			c := newController(
				fmt.Sprintf("controller-%s", gvr),
				informer,
				handler,
				conf)

			if cache.Get(gvr) == nil {
				go c.Run(stopCh)
			}
			cache.UpsertController(gvr, c)
		}
	}
}

func StartReplicationConfigController(conf *config.Config, stopCh chan struct{}) {
	log.Info("Starting ReplicationConfig Controller...")
	_, localDynamicAPI, err := utils.SetUpAPIs()
	if err != nil {
		log.Fatalf("Error creating APIs: %v", err)
		return
	}

	gvr := utils.GetGVRFromCrdType(utils.ReplicationConfigCRD, utils.V1Version)
	informer := getInformer(gvr)
	handler := handlers.NewReplicationConfigHandler(gvr, conf, localDynamicAPI)
	c := newController(
		fmt.Sprintf("controller-%s", utils.ReplicationConfigCRD),
		informer,
		handler,
		conf)
	go c.Run(stopCh)
}

func newController(name string, informer cache.SharedIndexInformer, handl handler.Handler,
	conf *config.Config) *handler.Controller {
	return handler.NewController(
		name,
		informer,
		handl,
		conf.Dispatcher.MaxWorkerCount,
		conf.IgnoredNamespaces.MatchNames)
}

func getInformer(gvr schema.GroupVersionResource) cache.SharedIndexInformer {
	conf, err := utils.GetRestConfig()
	if err != nil {
		log.WithError(err).Fatal("could not get config")
	}
	dynamicAPI, err := dynamic.NewForConfig(conf)
	if err != nil {
		log.WithError(err).Fatal("could not generate dynamic client for config")
	}
	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicAPI, 0, v1.NamespaceAll, nil)
	return f.ForResource(gvr).Informer()
}
