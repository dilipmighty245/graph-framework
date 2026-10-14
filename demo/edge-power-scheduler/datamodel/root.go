package root

import (
	"powerschedulermodel/config"
	"powerschedulermodel/inventory"
	"powerschedulermodel/runtimedesiredconfig"

	"github.com/dilipmighty245/graph-framework-for-microservices/nexus"
)

// Simulation scenario for DM demo :
// Problem 1: run app on the edge device to make sure it does not consume more than
// x Watts globally and is taking < y% of the power available on the edge
// Problem 2: Jobs are coming in with x amount of power needed per job,
// schedule the jobs across multiple edges ...
// track the completion of the jobs based on power availability and schedule the next job.
// dm layout
// Inventory -> Edges -> Power -> freePowerAvailable
//                     -> JobsRunning -> Progress
// Config -> Jobs -> PowerNeeded / Status->Running on and progress

// Setup for the Env. -
// Deploy kind clusters C1, E1, E2, E3, E4
// On C1 install the full data model and the controller for the power scheduler
// on E1..4 create the sync agent and specific parts of the data model
// On C1 create the inventory nodes manually and init the sync agents for runtime DM nodes to the specific edge cluster.
// The agent on the edge node will cycle power available randomly between the set of power points available and update crd
//               - it will accept jobs to run and based on power requirement will update progress every sec
//               - will garbage collect jobs that are older than an hour.
// The controller will create a backlog of jobs ... queue of 2-3 jobs that are always there with random power requirements
// and schedule them for execution and collect stats.
// Dynamic-Scale-EaseOfOnboarding: Edges can join and leave at any time, part of the demo we can show this as elastic / dynamic system.
// HA/Resilient: Edges can crash and restart / disconnect and reconnect without loss of system functionality.

// Datamodel graph root
type PowerScheduler struct {

	// Tags "Root" as a node in datamodel graph
	nexus.SingletonNode

	// Spec Fields
	Inventory         inventory.Inventory                    `nexus:"child"`
	Config            config.Config                          `nexus:"child"`
	DesiredSiteConfig runtimedesiredconfig.DesiredSiteConfig `nexus:"child"`
}
