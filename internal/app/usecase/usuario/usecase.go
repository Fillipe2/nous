package usuario

import (
	"context"
	"fmt"
	"time"

	"buildnix-backend/internal/app/dto"
	domain "buildnix-backend/internal/domain/usuario"

	"golang.org/x/crypto/bcrypt"
)

type UseCase struct {
	repository domain.Repository
}

func NewUseCase(repository domain.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (u *UseCase) CriarUsuario(ctx context.Context, input dto.CriarUsuarioInput) (*dto.UsuarioOutput, error) {
	if input.Nome == "" || input.Email == "" || input.Senha == "" {
		return nil, fmt.Errorf("nome, email e senha sao obrigatorios")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("criar usuario gerar senha: %w", err)
	}

	us := &domain.Usuario{
		Nome:         input.Nome,
		Email:        input.Email,
		SenhaHash:    string(hash),
		Role:         domain.RoleUsuario,
		Status:       domain.StatusAtivo,
		CriadoEm:     time.Now(),
		AtualizadoEm: time.Now(),
		Version:      1,
	}

	if err := u.repository.Salvar(ctx, us); err != nil {
		return nil, fmt.Errorf("criar usuario: %w", err)
	}

	return mapearUsuarioOutput(us), nil
}

func (u *UseCase) ObterUsuario(ctx context.Context, id string) (*dto.UsuarioOutput, error) {
	us, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obter usuario: %w", err)
	}
	return mapearUsuarioOutput(us), nil
}

func (u *UseCase) ObterUsuarioAutenticado(ctx context.Context, id string) (*dto.UsuarioOutput, error) {
	return u.ObterUsuario(ctx, id)
}

func (u *UseCase) AtualizarUsuario(ctx context.Context, id string, input dto.AtualizarUsuarioInput) (*dto.UsuarioOutput, error) {
	us, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("atualizar usuario: %w", err)
	}
	if input.Version != us.Version {
		return nil, domain.ErrVersaoConflito
	}

	if input.Nome != "" {
		us.Nome = input.Nome
	}
	if input.Email != "" {
		us.Email = input.Email
	}
	if input.Senha != "" {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(input.Senha), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, fmt.Errorf("atualizar usuario gerar senha: %w", hashErr)
		}
		us.SenhaHash = string(hash)
	}
	us.AtualizadoEm = time.Now()
	us.Version++

	if err := u.repository.Atualizar(ctx, us); err != nil {
		return nil, fmt.Errorf("atualizar usuario salvar: %w", err)
	}

	return mapearUsuarioOutput(us), nil
}

func (u *UseCase) ExcluirUsuario(ctx context.Context, id string) error {
	if err := u.repository.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("excluir usuario: %w", err)
	}
	return nil
}

func (u *UseCase) ListarUsuarios(ctx context.Context, filtro domain.FiltroListagem) (*dto.ListaUsuariosOutput, error) {
	usuarios, total, err := u.repository.Listar(ctx, filtro)
	if err != nil {
		return nil, fmt.Errorf("listar usuarios: %w", err)
	}

	var saida []*dto.UsuarioOutput
	for _, us := range usuarios {
		saida = append(saida, mapearUsuarioOutput(us))
	}

	return &dto.ListaUsuariosOutput{
		Dados:     saida,
		Total:     total,
		Pagina:    filtro.Pagina,
		PorPagina: filtro.PorPagina,
	}, nil
}

func mapearUsuarioOutput(us *domain.Usuario) *dto.UsuarioOutput {
	return &dto.UsuarioOutput{
		Id:           us.Id,
		Nome:         us.Nome,
		Email:        us.Email,
		Role:         string(us.Role),
		Status:       string(us.Status),
		CriadoEm:     us.CriadoEm,
		AtualizadoEm: us.AtualizadoEm,
		Version:      us.Version,
	}
}
