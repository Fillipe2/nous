package scanner

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"buildnix-backend/internal/domain/landingpage"
)

// LandingPage escaneia uma row do banco para a entidade de domínio LandingPage.
// Espera as colunas na ordem: id, cliente_id, negocio_id, categoria_id, template_id, tipo,
// nome, conteudo, status, is_draft, autosave_data, last_autosave_em, locked_by, locked_at,
// criado_em, atualizado_em, version.
func LandingPage(row pgx.Row) (*landingpage.LandingPage, error) {
	var (
		id             string
		clienteID      string
		negocioID      *string
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
	)
	if err := row.Scan(
		&id, &clienteID, &negocioID, &categoriaID, &templateID,
		&tipo, &nome, &conteudo, &status, &isDraft, &autosaveData,
		&lastAutosaveEm, &lockedBy, &lockedAt,
		&criadoEm, &atualizadoEm, &version,
	); err != nil {
		return nil, err
	}
	lp := &landingpage.LandingPage{
		Id:             id,
		ClienteId:      clienteID,
		Tipo:           landingpage.Tipo(tipo),
		Nome:           nome,
		Status:         landingpage.Status(status),
		IsDraft:        isDraft,
		AutosaveData:   json.RawMessage(autosaveData),
		LastAutosaveEm: lastAutosaveEm,
		LockedBy:       lockedBy,
		LockedAt:       lockedAt,
		CriadoEm:       criadoEm,
		AtualizadoEm:   atualizadoEm,
		Version:        version,
	}
	if negocioID != nil {
		lp.NegocioId = *negocioID
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
	return lp, nil
}
