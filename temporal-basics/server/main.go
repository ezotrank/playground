package main

import (
	"fmt"
	"net/http"

	temporal_basics "temporal-basics"

	tclient "go.temporal.io/sdk/client"
)

var client tclient.Client

func p2p(w http.ResponseWriter, r *http.Request) {
	args := temporal_basics.P2PWorkflowArguments{
		Sender:    r.URL.Query().Get("sender"),
		Recipient: r.URL.Query().Get("receiver"),
		Amount:    r.URL.Query().Get("amount"),
	}

	wf, err := client.ExecuteWorkflow(r.Context(), tclient.StartWorkflowOptions{
		TaskQueue: "default",
	}, temporal_basics.P2PWorkflow, args)
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(w, "Workflow ID: %s", wf.GetID())
}

func p2pAsync(w http.ResponseWriter, r *http.Request) {
	args := temporal_basics.P2PAsyncWorkflowArguments{
		Sender:    r.URL.Query().Get("sender"),
		Recipient: r.URL.Query().Get("receiver"),
		Amount:    r.URL.Query().Get("amount"),
	}

	wf, err := client.ExecuteWorkflow(r.Context(), tclient.StartWorkflowOptions{
		TaskQueue: "default",
	}, temporal_basics.P2PAsyncWorkflow, args)
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(w, "Workflow ID: %s", wf.GetID())
}

func confirm(w http.ResponseWriter, r *http.Request) {
	wfID := r.URL.Query().Get("workflow-id")
	signal := r.URL.Query().Get("signal")
	err := client.SignalWorkflow(r.Context(), wfID, "", signal, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	var err error
	if client, err = tclient.Dial(tclient.Options{}); err != nil {
		panic(err)
	}

	http.HandleFunc("/p2p", p2p)
	http.HandleFunc("/p2p-async", p2pAsync)
	http.HandleFunc("/confirm", confirm)

	if err := http.ListenAndServe(":8888", nil); err != nil {
		panic(err)
	}
}
