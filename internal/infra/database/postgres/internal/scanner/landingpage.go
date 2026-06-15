package scanner

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"nous/internal/domain/landingpage"
)

// LandingPage escaneia uma row do banco para a entidade de domínio LandingPage.
// Espera as colunas na ordem: id, negocio_id, unidade_id, categoria_id, template_id, tipo,
// nome, conteudo, status, is_draft, autosave_data, last_autosave_em, locked_by, locked_at,
// criado_em, atualizado_em, version, slug, url_publica, deploy_id, publicada_em.
func LandingPage(row pgx.Row) (*landingpage.LandingPage, error) {
	var (
		id             string
		negocioID      string
		unidadeID      *string
		categoriaID    *string
		templateID     *string
		tipo           string
		nome           string
		conteudo       *string
		status         string
		isDraft        bool
		autosaveData   []byte
		lastAutosaveEm *time.Time
		lockedBy       *string
		lockedAt       *time.Time
		criadoEm       time.Time
		atualizadoEm   time.Time
		version        int
		slug           *string
		urlPublica     *string
		deployId       *string
		publicadaEm    *time.Time
	)
	if err := row.Scan(
		&id, &negocioID, &unidadeID, &categoriaID, &templateID,
		&tipo, &nome, &conteudo, &status, &isDraft, &autosaveData,
		&lastAutosaveEm, &lockedBy, &lockedAt,
		&criadoEm, &atualizadoEm, &version,
		&slug, &urlPublica, &deployId, &publicadaEm,
	); err != nil {
		return nil, err
	}
	lp := &landingpage.LandingPage{
		Id:             id,
		NegocioId:      negocioID,
		Tipo:           landingpage.Tipo(tipo),
		Nome:           nome,
		Status:         landingpage.Status(status),
		IsDraft:        isDraft,
		AutosaveData:   json.RawMessage(autosaveData),
		LastAutosaveEm: lastAutosaveEm,
		LockedBy:       lockedBy,
		LockedAt:       lockedAt,
		UrlPublica:     urlPublica,
		DeployId:       deployId,
		PublicadaEm:    publicadaEm,
		CriadoEm:       criadoEm,
		AtualizadoEm:   atualizadoEm,
		Version:        version,
	}
	if unidadeID != nil {
		lp.UnidadeId = *unidadeID
	}
	if categoriaID != nil {
		lp.CategoriaId = *categoriaID
	}
	if templateID != nil {
		lp.TemplateId = *templateID
	}
	if conteudo != nil {
		lp.Conteudo = *conteudo
	}
	if slug != nil {
		lp.Slug = *slug
	}
	return lp, nil
}
