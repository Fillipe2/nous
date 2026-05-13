package noop

import (
	"context"

	"buildnix-backend/internal/domain/hosting"
)

type Deployer struct{}

func New() *Deployer { return &Deployer{} }

func (d *Deployer) Deploy(_ context.Context, slug string, _ []hosting.Arquivo) (*hosting.ResultadoDeploy, error) {
	return &hosting.ResultadoDeploy{
		DeployId:   slug,
		UrlPublica: "http://localhost/lp/" + slug,
	}, nil
}

func (d *Deployer) Despublicar(_ context.Context, _ string) error { return nil }
