package failover

import (
	"context"
	"fmt"

	"github.com/pace/bricks/backend/k8sapi"
)

const Label = "github.com.pace.bricks.activepassive"

type StateSetter interface {
	SetState(ctx context.Context, state string) error
}

type podStateSetter struct {
	k8sClient *k8sapi.Client
}

func NewPodStateSetter() (*podStateSetter, error) {
	k8sClient, err := k8sapi.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return &podStateSetter{k8sClient: k8sClient}, nil
}

func (p *podStateSetter) SetState(ctx context.Context, state string) error {
	return p.k8sClient.SetCurrentPodLabel(ctx, Label, state)
}

type CustomStateSetter struct {
	fn func(ctx context.Context, state string) error
}

func NewCustomStateSetter(fn func(ctx context.Context, state string) error) (*CustomStateSetter, error) {
	if fn == nil {
		return nil, fmt.Errorf("fn must not be nil")
	}

	return &CustomStateSetter{fn: fn}, nil
}

func (c *CustomStateSetter) SetState(ctx context.Context, state string) error {
	return c.fn(ctx, state)
}

type NoopStateSetter struct{}

func (n *NoopStateSetter) SetState(ctx context.Context, state string) error {
	return nil
}
