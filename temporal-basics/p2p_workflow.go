package temporal_basics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type P2PWorkflowArguments struct {
	Sender    string
	Recipient string
	Amount    string
}

func P2PWorkflow(ctx workflow.Context, args P2PWorkflowArguments) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1,
			MaximumInterval:    time.Second,
		},
	})
	var activity *P2PActivity
	var err error

	var code int

	err = workflow.ExecuteActivity(ctx, activity.TakeMoney, args.Sender, args.Amount).Get(ctx, &code)
	if err != nil {
		return fmt.Errorf("failed to take money: %w", err)
	}

	err = workflow.ExecuteActivity(ctx, activity.AddMoney, args.Recipient, args.Amount).Get(ctx, &code)
	if err != nil {
		return fmt.Errorf("failed to add money: %w", err)
	}

	return nil
}

type P2PActivity struct {
}

func (p *P2PActivity) TakeMoney(_ context.Context, account, amount string) (int, error) {
	fmt.Println("Start executing TakeMoney activity")
	defer fmt.Println("End executing TakeMoney activity")
	resp, err := http.Get("https://httpstat.us/200?" + "account=" + account + "&amount=" + amount)
	if err != nil {
		return 0, fmt.Errorf("failed to get http status: %w", err)
	}

	return resp.StatusCode, nil
}

func (p *P2PActivity) AddMoney(_ context.Context, account, amount string) (int, error) {
	fmt.Println("Start executing AddMoney activity")
	defer fmt.Println("End executing AddMoney activity")
	resp, err := http.Get("https://httpstat.us/200?" + "account=" + account + "&amount=" + amount)
	if err != nil {
		return 0, fmt.Errorf("failed to get http status: %w", err)
	}

	return resp.StatusCode, nil
}
