package engine

import (
	"context"
	"fmt"

	"github.com/opdev/knex"
)

type engine struct {
	image        string
	dockerConfig string
}

type OperatorOption func(*engine)

func New(ctx context.Context, options ...OperatorOption) *engine {
	return &engine{}
}

func (engine) ExecuteChecks(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (engine) Results(ctx context.Context) knex.Results {
	return knex.Results{}
}

func WithImage(image string) OperatorOption {
	return func(e *engine) {
		e.image = image
	}
}

func WithDockerConfig(dockerConfig string) OperatorOption {
	return func(e *engine) {
		e.dockerConfig = dockerConfig
	}
}
