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
func (u *UseCase) RegistrarUso(ctx context.Context, ev *faturamento.UsoClienteEvento) {
	if err := u.uso.Registrar(ctx, ev); err != nil {
		log.Printf("faturamento registrar uso tipo=%s cliente=%s: %v", ev.Tipo, ev.ClienteId, err)
		return
	}

	agora := time.Now().UTC()

	if ev.ValorUnitarioCentavos > 0 {
		centavos := int(ev.Quantidade) * ev.ValorUnitarioCentavos
		if err := u.fat.AdicionarUso(ctx, ev.ClienteId, agora, centavos); err != nil {
			log.Printf("faturamento adicionar uso cliente=%s: %v", ev.ClienteId, err)
		}
	}

	if err := u.bi.IncrementarMetrica(ctx, ev.ClienteId, agora, ev.Tipo, ev.Quantidade, map[string]string{}); err != nil {
		log.Printf("bi incrementar metrica tipo=%s cliente=%s: %v", ev.Tipo, ev.ClienteId, err)
	}
}

// FecharCompetencia fecha o faturamento mensal de um cliente para a competência fornecida.
func (u *UseCase) FecharCompetencia(ctx context.Context, clienteId string, competencia time.Time) error {
	if err := u.fat.FecharCompetencia(ctx, clienteId, competencia); err != nil {
		return fmt.Errorf("fechar competencia cliente=%s: %w", clienteId, err)
	}
	return nil
}
