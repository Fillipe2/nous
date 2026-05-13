package hosting

import "context"

type Arquivo struct {
	Caminho  string
	Conteudo []byte
}

type ResultadoDeploy struct {
	DeployId   string
	UrlPublica string
}

type Deployer interface {
	Deploy(ctx context.Context, slug string, arquivos []Arquivo) (*ResultadoDeploy, error)
	Despublicar(ctx context.Context, slug string) error
}
