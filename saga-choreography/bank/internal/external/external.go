package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ezotrank/playground/saga-choreography/bank/internal/repository"
)

func NewExternal(addr string) *External {
	return &External{
		addr: addr,
	}
}

type External struct {
	addr string
}

func (e External) CreateAccount(ctx context.Context, account *repository.Account) error {
	body, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %v", err)
	}
	u := e.addr + "/accounts"
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	client := http.Client{
		Timeout: time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	switch {
	case resp.StatusCode == http.StatusOK:
		account.Status = repository.AccountStatusRegistered
	case resp.StatusCode == http.StatusBadRequest:
		account.Status = repository.AccountStatusRejected
	default:
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}
