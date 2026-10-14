# Playground TL;DR

Welcome to the Nexus playground TL;DR version.

This tutorial will walk through a minimalistic workflow to try out Nexus playground.

The goal is to get you going with Nexus in the shortest possible time.

## 1. Clone Nexus Repo
```
git clone https://github.com/dilipmighty245/graph-framework-for-microservices.git
cd applications.development.framework.nexus/
export NEXUS_REPO_DIR=${PWD}
```

## 2. Build & Install Nexus CLI

### For Linux
```
make cli.build.linux; sudo make cli.install.linux
```
For MacOS
```
make cli.build.darwin; sudo make cli.install.darwin
```

## 3. Install Nexus Runtime

```
make runtime.install
export HOST_KUBECONFIG=$PWD/nexus-runtime-manifests/k8s/kubeconfig
```

## 4 Setup Workspace for our playground application: Sock Shop

```
export NEXUS_REPO_DIR=${PWD};
mkdir sock-shop-saas-app;cd sock-shop-saas-app;
go mod init demo;
go mod tidy;
go mod edit -replace sockshop=./datamodel;
mkdir datamodel; cd datamodel
```

## 5 Declare data model specification for Sock Shop

### 5.1 Initialize data model 

So we want to call our data model, sockshop.

```
# NOTE: NEXUS_REPO_DIR is the local directory where Nexus repo is cloned to.
nexus datamodel init --name sockshop --group sockshop.com --local-dir $NEXUS_REPO_DIR
```

With the above init, the entire code setup needed to define a fully functional data model is now ready.

### 5.2 Write the model specification fo Sock Shop

Copy & paste the below code snippet to file root.go in the datamodel directory

#### File: root.go

```
package root

import (
	"github.com/dilipmighty245/graph-framework-for-microservices/nexus/nexus"
)

type SockShop struct {
	nexus.SingletonNode

	OrgName  string
	Location string
	Website  string

	Inventory      Socks    `nexus:"children"`
	PO             Orders   `nexus:"children"`
	ShippingLedger Shipping `nexus:"children"`
}

var SocksRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/sock/{root.Socks}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
		{
			Uri:     "/socks",
			Methods: nexus.HTTPListResponse,
		},
	},
}

// nexus-rest-api-gen:SocksRestAPISpec
type Socks struct {
	nexus.Node

	Brand string
	Color string
	Size  int
}

var OrderRestAPISpec = nexus.RestAPISpec{
	Uris: []nexus.RestURIs{
		{
			Uri:     "/order/{root.Orders}",
			Methods: nexus.DefaultHTTPMethodsResponses,
		},
	},
}

// nexus-rest-api-gen:OrderRestAPISpec
type Orders struct {
	nexus.Node

	SockName string
	Address  string

	Cart     Socks    `nexus:"link"`
	Shipping Shipping `nexus:"link"`
}

type Shipping struct {
	nexus.Node

	TrackingId int
}
```
### 5.3 Build data model for Sock Shop

Nexus compiler can be invoked to build the datamodel with the following command:

```
DATAMODEL_DOCKER_REGISTRY=<container-registry-for-datamodel> TAG=<datamodel-tag> make docker.build
```

## 6 Install Sock Shop data model on Nexus Runtime

```
DATAMODEL_DOCKER_REGISTRY=<container-registry-for-datamodel> TAG=<datamodel-tag> make dm.install
```

REST API is available [here](http://localhost:8082/sockshop.com/docs#/)

## 7 Let's write our business logic

The business logic is quite simple: initiate shipping for an order, when the order is placed.

### Step 7.1 Go to top workdir
```
cd ..
```

### Step 7.2 Copy paste this code to file: main.go

```
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	sockshopv1 "sockshop/build/apis/root.sockshop.com/v1"
	nexus_client "sockshop/build/nexus-client"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func generateTrackingId() int {
	return rand.Intn(10000-1) + 1
}

func ProcessOrder(order *nexus_client.RootOrders) {

	// Check if the Order has a Shipping object associated with it, already.
	shipping, err := order.GetShipping(context.TODO())

	if nexus_client.IsLinkNotFound(err) {

		// No shipping object was found associated with this order.

		fmt.Printf("Hurray...Order %v received!\n", order.DisplayName())

		// Construct a Shipping object.
		shippingObj := &sockshopv1.Shipping{
			ObjectMeta: metav1.ObjectMeta{
				Name: order.DisplayName(),
			},
			Spec: sockshopv1.ShippingSpec{
				TrackingId: generateTrackingId(),
			},
		}

		// Create a shipping object.
		shippingInfo, _ := nexusClient.RootSockShop().AddShippingLedger(context.TODO(), shippingObj)

		// Associate the shipping object to the Order.
		order.LinkShipping(context.TODO(), shippingInfo)

		fmt.Printf("Order %v shipped to %v! Tracking Id: %v\n", order.DisplayName(), order.Spec.Address, shippingInfo.Spec.TrackingId)
	} else {

		// A shipping object was found associated with this order, already. Nothing to be done.

		fmt.Printf("Order %v has been shipped. Tracking Id: %v\n", order.DisplayName(), shipping.Spec.TrackingId)
	}
}

var nexusClient *nexus_client.Clientset

func main() {

	rand.Seed(time.Now().UnixNano())
	var kubeconfig string
	flag.StringVar(&kubeconfig, "k", "", "Absolute path to the kubeconfig file. Defaults to ~/.kube/config.")
	flag.Parse()

	var config *rest.Config
	if len(kubeconfig) != 0 {
		var err error
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}
	} else {
		config = &rest.Config{Host: "localhost:9000"}
	}

	nexusClient, _ = nexus_client.NewForConfig(config)

	nexusClient.RootSockShop().PO("*").RegisterAddCallback(ProcessOrder)

	nexusClient.AddRootSockShop(context.TODO(), &sockshopv1.SockShop{
		Spec: sockshopv1.SockShopSpec{
			OrgName:  "Unicorn",
			Location: "Seattle",
			Website:  "website.com",
		},
	})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		<-sigs
		done <- true
	}()
	<-done
	fmt.Println("exiting")
}
```

## Step 7.3 Let's run the Application

```
go mod tidy; go run main.go
```

Our application is now production ready !!!

## 8 Let's test our application

### 8.1 Add a Socks to our inventory. We gotta sell something right !

```
HTTP Method: PUT

URI: /sock/{root.Socks}

Name of the root.Socks node: Polo

Spec:
{
  "brand": "polo,inc",
  "color": "white",
  "size": 9
}
```

### 8.2 Let's place an Order for a Socks through REST API

Place an Order.

```
HTTP Method: PUT

URI: /order/{root.Orders}

Name of the root.Orders node: MyFirstOrder

Spec:

{
  "address": "SFO",
  "sockName": "polo"
}
```

## 9.3 Application will be notified of new order and will create a shipping request.

```
➜ go run main.go

Hurray...Order MyFirstOrder received!
Order MyFirstOrder shipped to SFO! Tracking Id: 9815
```
