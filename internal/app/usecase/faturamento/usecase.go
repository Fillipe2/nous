package faturamento

import (
	"context"
	"fmt"
	"log"
	"time"

	"nous/internal/domain/faturamento"
)

// UseCase centraliza operações de registro de uso, BI e fechamento de competência.
type UseCase struct {
	uso         faturamento.UsoRepository
	fat         faturamento.FaturamentoRepository
	bi          faturamento.BiRepository
}

func NewUseCase(uso faturamento.UsoRepository, fat faturamento.FaturamentoRepository, bi faturamento.BiRepository) *UseCase {
	return &UseCase{uso: uso, fat: fat, bi: bi}
}

// RegistrarUso persiste um evento de uso e atualiza o faturamento/BI do mês.
// Erros não interrompem o fluxo principal (logados, não propagados ao chamador).
func (u *UseCase) RegistrarUso(ctx context.Context, ev *faturamento.UsoNegocioEvento) {
	if err := u.uso.Registrar(ctx, ev); err != nil {
		log.Printf("faturamento registrar uso tipo=%s negocio=%s: %v", ev.Tipo, ev.NegocioId, err)
		return
	}

	agora := time.Now().UTC()

	if ev.ValorUnitarioCentavos > 0 {
		centavos := int(ev.Quantidade) * ev.ValorUnitarioCentavos
		if err := u.fat.AdicionarUso(ctx, ev.NegocioId, agora, centavos); err != nil {
			log.Printf("faturamento adicionar uso negocio=%s: %v", ev.NegocioId, err)
		}
	}

	if err := u.bi.IncrementarMetrica(ctx, ev.NegocioId, agora, ev.Tipo, ev.Quantidade, map[string]string{}); err != nil {
		log.Printf("bi incrementar metrica tipo=%s negocio=%s: %v", ev.Tipo, ev.NegocioId, err)
	}
}

// FecharCompetencia fecha o faturamento mensal de um negocio para a competência fornecida.
func (u *UseCase) FecharCompetencia(ctx context.Context, negocioId string, competencia time.Time) error {
	if err := u.fat.FecharCompetencia(ctx, negocioId, competencia); err != nil {
		return fmt.Errorf("fechar competencia negocio=%s: %w", negocioId, err)
	}
	return nil
}
