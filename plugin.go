package plugin

import (
	"context"
	"errors"

	"github.com/opdev/operator-certification/internal/flags"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	"github.com/opdev/knex"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Assert that we implement the Plugin interface.
var _ knex.Plugin = NewPlugin()

var vers = semver.MustParse("0.0.1")

func init() {
	knex.Register("check-operator", NewPlugin())
}

type operatorEngine interface {
	ExecuteChecks(context.Context) error
	Results(context.Context) knex.Results
}

type plug struct {
	image  string
	engine operatorEngine
}

func NewPlugin() *plug {
	p := plug{}
	// plugin-related things may happen here.
	return &p
}

func (p *plug) Register() error {
	return nil
}

func (p *plug) Name() string {
	return "Operator Certification"
}

func (p *plug) Init(ctx context.Context, cfg *viper.Viper, args []string) error {
	l := logr.FromContextOrDiscard(ctx)
	l.Info("Initializing Container Certification")
	if len(args) != 1 {
		return errors.New("a single argument is required (the container image to test)")
	}

	return nil
}

func (p *plug) BindFlags(f *pflag.FlagSet) *pflag.FlagSet {
	flags.BindFlagDockerConfigFilePath(f)
	return f
}

func (p *plug) Version() semver.Version {
	return *vers
}

func (p *plug) ExecuteChecks(ctx context.Context) error {
	l := logr.FromContextOrDiscard(ctx)
	l.Info("Execute Checks Called")
	return p.engine.ExecuteChecks(ctx)
}

func (p *plug) Results(ctx context.Context) knex.Results {
	return p.engine.Results(ctx)
}

func (p *plug) Submit(ctx context.Context) error {
	l := logr.FromContextOrDiscard(ctx)
	l.Info("Submit called")
	return nil
}
