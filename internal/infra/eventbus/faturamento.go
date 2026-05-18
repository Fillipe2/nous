package eventbus

import (
	"context"
	"log"
	"time"

	"buildnix-backend/internal/domain/evento"
	"buildnix-backend/internal/domain/faturamento"
)

// FaturamentoPublisher é um decorator de Publisher que intercepta eventos conhecidos
// e registra automaticamente o uso correspondente no módulo de faturamento/BI.
// Erros de faturamento nunca interrompem o fluxo principal.
type FaturamentoPublisher struct {
	next faturamento.UsoRepository
	bi   faturamento.BiRepository
	pub  evento.Publisher
}

func NewFaturamentoPublisher(pub evento.Publisher, uso faturamento.UsoRepository, bi faturamento.BiRepository) evento.Publisher {
	return &FaturamentoPublisher{next: uso, bi: bi, pub: pub}
}

func (p *FaturamentoPublisher) Publicar(ctx context.Context, ev evento.Evento) error {
	err := p.pub.Publicar(ctx, ev)

	// Registra uso independentemente do resultado da publicação.
	go p.registrarUso(ev)

	return err
}

func (p *FaturamentoPublisher) registrarUso(ev evento.Evento) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	uso := usoParaEvento(ev)
	if uso == nil {
		return
	}

	if err := p.next.Registrar(ctx, uso); err != nil {
		log.Printf("[faturamento] registrar uso evento=%s: %v", ev.EventType, err)
		return
	}

	dimensoes := dimensoesParaEvento(ev)
	if err := p.bi.IncrementarMetrica(ctx, uso.ClienteId, time.Now().UTC(), ev.EventType, 1, dimensoes); err != nil {
		log.Printf("[faturamento] bi metrica evento=%s: %v", ev.EventType, err)
	}
}

// usoParaEvento mapeia um evento de domínio para um UsoClienteEvento,
// extraindo cliente_id e entidade do payload quando disponíveis.
func usoParaEvento(ev evento.Evento) *faturamento.UsoClienteEvento {
	switch ev.EventType {
	case "landingpage.criada", "landingpage.publicada", "landingpage.exportada":
		return &faturamento.UsoClienteEvento{
			ClienteId:  strPayload(ev.Payload, "cliente_id"),
			EntidadeId: strPayload(ev.Payload, "landingpage_id"),
			Entidade:   "landingpage",
			Tipo:       ev.EventType,
			Quantidade: 1,
			Unidade:    "unidade",
			OcorridoEm: ev.OccurredAt,
		}
	case "lead.capturado":
		return &faturamento.UsoClienteEvento{
			ClienteId:  strPayload(ev.Payload, "cliente_id"),
			EntidadeId: strPayload(ev.Payload, "lead_id"),
			Entidade:   "lead",
			Tipo:       ev.EventType,
			Quantidade: 1,
			Unidade:    "unidade",
			OcorridoEm: ev.OccurredAt,
		}
	case "lead_aquisicao.registrado":
		return &faturamento.UsoClienteEvento{
			ClienteId:  "",
			EntidadeId: strPayload(ev.Payload, "lead_aquisicao_id"),
			Entidade:   "lead_aquisicao",
			Tipo:       ev.EventType,
			Quantidade: 1,
			Unidade:    "unidade",
			OcorridoEm: ev.OccurredAt,
		}
	}
	return nil
}

func dimensoesParaEvento(ev evento.Evento) map[string]string {
	dim := map[string]string{"event_type": ev.EventType}
	if source := ev.Source; source != "" {
		dim["source"] = source
	}
	return dim
}

func strPayload(payload map[string]interface{}, key string) string {
	if v, ok := payload[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
