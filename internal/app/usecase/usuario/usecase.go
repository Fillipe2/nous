package usuario

import (
	"context"
	"fmt"
	"time"

	"nous/internal/app/dto"
	domain "nous/internal/domain/usuario"

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
	if err := domain.ValidarSenha(input.Senha); err != nil {
		return nil, err
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
	if input.Email != "" && input.Email != us.Email {
		us.Email = input.Email
		// E-mail alterado deixa de ser verificado até nova confirmação.
		us.EmailVerified = false
	}
	us.AtualizadoEm = time.Now()
	us.Version++

	if err := u.repository.Atualizar(ctx, us); err != nil {
		return nil, fmt.Errorf("atualizar usuario salvar: %w", err)
	}

	return mapearUsuarioOutput(us), nil
}

func (u *UseCase) AlterarSenha(ctx context.Context, id string, input dto.AlterarSenhaInput) error {
	if input.SenhaAtual == "" || input.NovaSenha == "" {
		return fmt.Errorf("senha atual e nova senha sao obrigatorias")
	}
	if err := domain.ValidarSenha(input.NovaSenha); err != nil {
		return err
	}

	us, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return fmt.Errorf("alterar senha: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(us.SenhaHash), []byte(input.SenhaAtual)); err != nil {
		return domain.ErrNaoAutorizado
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NovaSenha), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("alterar senha gerar hash: %w", err)
	}
	us.SenhaHash = string(hash)
	us.AtualizadoEm = time.Now()
	us.Version++

	if err := u.repository.Atualizar(ctx, us); err != nil {
		return fmt.Errorf("alterar senha salvar: %w", err)
	}
	return nil
}

// AlterarRole promove ou rebaixa um usuário (ação restrita a admin). Impede
// rebaixar o último administrador ativo, evitando lockout do sistema.
func (u *UseCase) AlterarRole(ctx context.Context, id string, novoRole string) (*dto.UsuarioOutput, error) {
	role := domain.Role(novoRole)
	if !domain.RoleValida(role) {
		return nil, domain.ErrRoleInvalida
	}

	us, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("alterar role: %w", err)
	}
	if us.Role == role {
		return mapearUsuarioOutput(us), nil
	}

	if us.Role == domain.RoleAdmin && role != domain.RoleAdmin {
		total, err := u.repository.ContarAdministradoresAtivos(ctx)
		if err != nil {
			return nil, fmt.Errorf("alterar role contar admins: %w", err)
		}
		if total <= 1 {
			return nil, domain.ErrUltimoAdmin
		}
	}

	us.Role = role
	us.AtualizadoEm = time.Now()
	us.Version++
	if err := u.repository.Atualizar(ctx, us); err != nil {
		return nil, fmt.Errorf("alterar role salvar: %w", err)
	}
	return mapearUsuarioOutput(us), nil
}

// AlterarStatus ativa ou inativa um usuário (ação restrita a admin). A exclusão
// é feita pelo soft delete dedicado. Impede inativar o último admin ativo.
func (u *UseCase) AlterarStatus(ctx context.Context, id string, novoStatus string) (*dto.UsuarioOutput, error) {
	status := domain.Status(novoStatus)
	if !domain.StatusGerenciavelValido(status) {
		return nil, domain.ErrStatusInvalido
	}

	us, err := u.repository.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("alterar status: %w", err)
	}
	if us.Status == status {
		return mapearUsuarioOutput(us), nil
	}

	if us.Role == domain.RoleAdmin && us.Status == domain.StatusAtivo && status != domain.StatusAtivo {
		total, err := u.repository.ContarAdministradoresAtivos(ctx)
		if err != nil {
			return nil, fmt.Errorf("alterar status contar admins: %w", err)
		}
		if total <= 1 {
			return nil, domain.ErrUltimoAdmin
		}
	}

	us.Status = status
	us.AtualizadoEm = time.Now()
	us.Version++
	if err := u.repository.Atualizar(ctx, us); err != nil {
		return nil, fmt.Errorf("alterar status salvar: %w", err)
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
		Id:            us.Id,
		Nome:          us.Nome,
		Email:         us.Email,
		Role:          string(us.Role),
		Status:        string(us.Status),
		EmailVerified: us.EmailVerified,
		UltimoLoginEm: us.UltimoLoginEm,
		CriadoEm:      us.CriadoEm,
		AtualizadoEm:  us.AtualizadoEm,
		Version:       us.Version,
	}
}
