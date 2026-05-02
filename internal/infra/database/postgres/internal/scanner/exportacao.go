package scanner

import (
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/exportacao"
)

// Exportacao escaneia uma row do banco para a entidade de domínio Exportacao.
func Exportacao(row pgx.Row) (*exportacao.Exportacao, error) {
	var (
		id            string
		landingPageID string
		usuarioID     string
		nomeArquivo   string
		contentType   string
		tamanhoBytes  int64
		arquivo       []byte
		hash          string
		criadoEm      time.Time
	)
	if err := row.Scan(&id, &landingPageID, &usuarioID, &nomeArquivo, &contentType, &tamanhoBytes, &arquivo, &hash, &criadoEm); err != nil {
		return nil, err
	}
	return &exportacao.Exportacao{
		Id:            id,
		LandingPageId: landingPageID,
		UsuarioId:     usuarioID,
		NomeArquivo:   nomeArquivo,
		ContentType:   contentType,
		TamanhoBytes:  tamanhoBytes,
		Arquivo:       arquivo,
		Hash:          hash,
		CriadoEm:      criadoEm,
	}, nil
}
