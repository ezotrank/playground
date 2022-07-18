package temporal_basics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type P2PAsyncWorkflowArguments struct {
	Sender    string
	Recipient string
	Amount    string
}

func P2PAsyncWorkflow(ctx workflow.Context, args P2PAsyncWorkflowArguments) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1,
			MaximumInterval:    time.Second,
		},
	})
	var activity *P2PAsyncActivity
	var err error

	var code int

	err = workflow.ExecuteActivity(ctx, activity.TakeMoneyAsync, args.Sender, args.Amount).Get(ctx, &code)
	if err != nil {
		return fmt.Errorf("failed to take money: %w", err)
	}

	var done bool

	fmt.Println("Start waiting webhook for TakeMoney activity")
	signalChan := workflow.GetSignalChannel(ctx, "money-taken")
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &done)
	})
	selector.Select(ctx)

	err = workflow.ExecuteActivity(ctx, activity.AddMoneyAsync, args.Recipient, args.Amount).Get(ctx, &code)
	if err != nil {
		return fmt.Errorf("failed to add money: %w", err)
	}

	fmt.Println("Start waiting webhook for AddMoney activity")
	signalChan = workflow.GetSignalChannel(ctx, "money-added")
	selector = workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &done)
	})
	selector.Select(ctx)

	return nil
}

type P2PAsyncActivity struct {
}

func (p *P2PAsyncActivity) TakeMoneyAsync(_ context.Context, account, amount string) (int, error) {
	fmt.Println("Start executing TakeMoney activity")
	defer fmt.Println("End executing TakeMoney activity")
	resp, err := http.Get("https://httpstat.us/200?" + "account=" + account + "&amount=" + amount)
	if err != nil {
		return 0, fmt.Errorf("failed to get http status: %w", err)
	}

	return resp.StatusCode, nil
}

func (p *P2PAsyncActivity) AddMoneyAsync(_ context.Context, account, amount string) (int, error) {
	//return 0, errors.New("some unexpected error")
	fmt.Println("Start executing AddMoney activity")
	defer fmt.Println("End executing AddMoney activity")
	resp, err := http.Get("https://httpstat.us/200?" + "account=" + account + "&amount=" + amount)
	if err != nil {
		return 0, fmt.Errorf("failed to get http status: %w", err)
	}

	return resp.StatusCode, nil
}
