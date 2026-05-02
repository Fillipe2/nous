// Package migrations exporta o sistema de arquivos embarcado com todos os
// scripts de migração do banco de dados. Ao importar este pacote, a aplicação
// pode rodar as migrações sem depender de arquivos no disco em produção.
package migrations

import "embed"

// Files contem todos os arquivos .sql desta pasta, disponiveis em tempo de execucao.
//
//go:embed postgres/*.sql
var Files embed.FS
