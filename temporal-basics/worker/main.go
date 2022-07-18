package main

import (
	temporal_basics "temporal-basics"

	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var client tclient.Client

func main() {
	var err error
	if client, err = tclient.Dial(tclient.Options{}); err != nil {
		panic(err)
	}

	w := worker.New(client, "default", worker.Options{})

	w.RegisterWorkflow(temporal_basics.P2PWorkflow)
	w.RegisterActivity(&temporal_basics.P2PActivity{})

	w.RegisterWorkflow(temporal_basics.P2PAsyncWorkflow)
	w.RegisterActivity(&temporal_basics.P2PAsyncActivity{})

	if err = w.Run(worker.InterruptCh()); err != nil {
		panic(err)
	}
}
