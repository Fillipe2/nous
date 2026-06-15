package usuario

import (
	"context"
	"errors"
	"testing"
	"time"

	"nous/internal/app/dto"
	domain "nous/internal/domain/usuario"

	"golang.org/x/crypto/bcrypt"
)

type usuarioRepositoryFake struct {
	usuario      *domain.Usuario
	atualizado   *domain.Usuario
	salvo        *domain.Usuario
	adminsAtivos int
	salvarErr    error
	atualizarErr error
}

func (r *usuarioRepositoryFake) Salvar(_ context.Context, u *domain.Usuario) error {
	if r.salvarErr != nil {
		return r.salvarErr
	}
	u.Id = "usuario-novo"
	r.salvo = u
	return nil
}

func (r *usuarioRepositoryFake) BuscarPorID(context.Context, string) (*domain.Usuario, error) {
	if r.usuario == nil {
		return nil, domain.ErrNaoEncontrado
	}
	return r.usuario, nil
}

func (r *usuarioRepositoryFake) BuscarPorEmail(context.Context, string) (*domain.Usuario, error) {
	return nil, domain.ErrNaoEncontrado
}

func (r *usuarioRepositoryFake) Atualizar(_ context.Context, u *domain.Usuario) error {
	if r.atualizarErr != nil {
		return r.atualizarErr
	}
	r.atualizado = u
	return nil
}

func (r *usuarioRepositoryFake) AtualizarUltimoLogin(context.Context, string, time.Time) error {
	return nil
}

func (r *usuarioRepositoryFake) ContarAdministradoresAtivos(context.Context) (int, error) {
	return r.adminsAtivos, nil
}

func (r *usuarioRepositoryFake) SoftDelete(context.Context, string) error { return nil }
func (r *usuarioRepositoryFake) Listar(context.Context, domain.FiltroListagem) ([]*domain.Usuario, int, error) {
	return nil, 0, nil
}
func (r *usuarioRepositoryFake) SalvarAuth(context.Context, *domain.UsuarioAuth) error { return nil }
func (r *usuarioRepositoryFake) BuscarAuthPorProvider(context.Context, string, string) (*domain.UsuarioAuth, error) {
	return nil, domain.ErrNaoEncontrado
}
func (r *usuarioRepositoryFake) SalvarRefreshToken(context.Context, *domain.RefreshToken) error {
	return nil
}
func (r *usuarioRepositoryFake) BuscarRefreshTokenPorHash(context.Context, string) (*domain.RefreshToken, error) {
	return nil, domain.ErrNaoAutorizado
}
func (r *usuarioRepositoryFake) RevogarRefreshToken(context.Context, string, time.Time) error {
	return nil
}
func (r *usuarioRepositoryFake) RevogarRefreshTokensDoUsuario(context.Context, string, time.Time) error {
	return nil
}
func (r *usuarioRepositoryFake) ExpurgarRefreshTokens(context.Context, time.Time) error { return nil }
func (r *usuarioRepositoryFake) SalvarResetSenhaToken(context.Context, *domain.ResetSenhaToken) error {
	return nil
}
func (r *usuarioRepositoryFake) BuscarResetSenhaTokenPorHash(context.Context, string) (*domain.ResetSenhaToken, error) {
	return nil, domain.ErrNaoEncontrado
}
func (r *usuarioRepositoryFake) MarcarResetSenhaTokenUsado(context.Context, string, time.Time) error {
	return nil
}
func (r *usuarioRepositoryFake) SalvarVerificacaoEmailToken(context.Context, *domain.VerificacaoEmailToken) error {
	return nil
}
func (r *usuarioRepositoryFake) BuscarVerificacaoEmailTokenPorHash(context.Context, string) (*domain.VerificacaoEmailToken, error) {
	return nil, domain.ErrNaoEncontrado
}
func (r *usuarioRepositoryFake) MarcarVerificacaoEmailTokenUsado(context.Context, string, time.Time) error {
	return nil
}

func TestCriarUsuarioValidaObrigatorios(t *testing.T) {
	useCase := NewUseCase(&usuarioRepositoryFake{})

	_, err := useCase.CriarUsuario(context.Background(), dto.CriarUsuarioInput{Email: "user@sognix.com.br", Senha: "senha"})

	if err == nil {
		t.Fatal("erro esperado para nome vazio")
	}
}

func TestCriarUsuarioSalvaSenhaComHash(t *testing.T) {
	repository := &usuarioRepositoryFake{}
	useCase := NewUseCase(repository)

	output, err := useCase.CriarUsuario(context.Background(), dto.CriarUsuarioInput{Nome: "Usuario", Email: "user@sognix.com.br", Senha: "senha-segura"})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.Id != "usuario-novo" || output.Role != string(domain.RoleUsuario) || output.Status != string(domain.StatusAtivo) {
		t.Fatalf("saida inesperada: %+v", output)
	}
	if repository.salvo.SenhaHash == "senha-segura" {
		t.Fatal("senha nao deveria ser salva em texto puro")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.salvo.SenhaHash), []byte("senha-segura")); err != nil {
		t.Fatalf("hash salvo nao corresponde a senha: %v", err)
	}
}

func TestAtualizarUsuarioIncrementaVersaoSemAlterarSenha(t *testing.T) {
	senhaAntiga, err := bcrypt.GenerateFromPassword([]byte("antiga"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("gerar hash: %v", err)
	}
	repository := &usuarioRepositoryFake{usuario: &domain.Usuario{Id: "usuario-1", Nome: "Antigo", Email: "old@sognix.com.br", SenhaHash: string(senhaAntiga), Role: domain.RoleUsuario, Status: domain.StatusAtivo, Version: 3}}
	useCase := NewUseCase(repository)

	output, err := useCase.AtualizarUsuario(context.Background(), "usuario-1", dto.AtualizarUsuarioInput{Nome: "Novo", Email: "new@sognix.com.br", Version: 3})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if output.Version != 4 || output.Nome != "Novo" || output.Email != "new@sognix.com.br" {
		t.Fatalf("saida inesperada: %+v", output)
	}
	if repository.atualizado == nil || repository.atualizado.SenhaHash != string(senhaAntiga) {
		t.Fatal("senha nao deveria ser alterada no update de perfil")
	}
}

func TestAlterarSenhaValidaAtualERehashNova(t *testing.T) {
	senhaAntiga, err := bcrypt.GenerateFromPassword([]byte("antiga"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("gerar hash: %v", err)
	}
	repository := &usuarioRepositoryFake{usuario: &domain.Usuario{Id: "usuario-1", SenhaHash: string(senhaAntiga), Role: domain.RoleUsuario, Status: domain.StatusAtivo, Version: 3}}
	useCase := NewUseCase(repository)

	if err := useCase.AlterarSenha(context.Background(), "usuario-1", dto.AlterarSenhaInput{SenhaAtual: "antiga", NovaSenha: "nova-senha-segura"}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repository.atualizado == nil || repository.atualizado.Version != 4 {
		t.Fatalf("usuario atualizado inesperado: %+v", repository.atualizado)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.atualizado.SenhaHash), []byte("nova-senha-segura")); err != nil {
		t.Fatalf("hash atualizado nao corresponde a senha nova: %v", err)
	}
}

func TestAlterarSenhaRecusaSenhaAtualInvalida(t *testing.T) {
	senhaAntiga, err := bcrypt.GenerateFromPassword([]byte("antiga"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("gerar hash: %v", err)
	}
	repository := &usuarioRepositoryFake{usuario: &domain.Usuario{Id: "usuario-1", SenhaHash: string(senhaAntiga), Version: 3}}
	useCase := NewUseCase(repository)

	err = useCase.AlterarSenha(context.Background(), "usuario-1", dto.AlterarSenhaInput{SenhaAtual: "errada", NovaSenha: "nova-senha-segura"})

	if !errors.Is(err, domain.ErrNaoAutorizado) {
		t.Fatalf("erro esperado ErrNaoAutorizado, recebido %v", err)
	}
	if repository.atualizado != nil {
		t.Fatal("senha nao deveria ser atualizada com senha atual invalida")
	}
}

func TestCriarUsuarioEmailDuplicado(t *testing.T) {
	repository := &usuarioRepositoryFake{salvarErr: domain.ErrEmailDuplicado}
	useCase := NewUseCase(repository)

	_, err := useCase.CriarUsuario(context.Background(), dto.CriarUsuarioInput{Nome: "Maria", Email: "m@x.com", Senha: "uma-senha-segura"})
	if !errors.Is(err, domain.ErrEmailDuplicado) {
		t.Fatalf("esperava ErrEmailDuplicado, recebeu %v", err)
	}
}

func TestAlterarRolePromove(t *testing.T) {
	repository := &usuarioRepositoryFake{usuario: &domain.Usuario{Id: "u1", Role: domain.RoleUsuario, Status: domain.StatusAtivo, Version: 2}}
	useCase := NewUseCase(repository)

	out, err := useCase.AlterarRole(context.Background(), "u1", "admin")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Role != "admin" {
		t.Fatalf("esperava role admin, recebeu %s", out.Role)
	}
	if repository.atualizado == nil || repository.atualizado.Version != 3 {
		t.Fatalf("atualizacao inesperada: %+v", repository.atualizado)
	}
}

func TestAlterarRoleRoleInvalida(t *testing.T) {
	repository := &usuarioRepositoryFake{usuario: &domain.Usuario{Id: "u1", Role: domain.RoleUsuario, Status: domain.StatusAtivo, Version: 1}}
	useCase := NewUseCase(repository)

	_, err := useCase.AlterarRole(context.Background(), "u1", "superuser")
	if !errors.Is(err, domain.ErrRoleInvalida) {
		t.Fatalf("esperava ErrRoleInvalida, recebeu %v", err)
	}
}

func TestAlterarRoleBloqueiaUltimoAdmin(t *testing.T) {
	repository := &usuarioRepositoryFake{
		usuario:      &domain.Usuario{Id: "u1", Role: domain.RoleAdmin, Status: domain.StatusAtivo, Version: 1},
		adminsAtivos: 1,
	}
	useCase := NewUseCase(repository)

	_, err := useCase.AlterarRole(context.Background(), "u1", "usuario")
	if !errors.Is(err, domain.ErrUltimoAdmin) {
		t.Fatalf("esperava ErrUltimoAdmin, recebeu %v", err)
	}
	if repository.atualizado != nil {
		t.Fatal("nao deveria rebaixar o ultimo admin")
	}
}

func TestAlterarRoleRebaixaComOutrosAdmins(t *testing.T) {
	repository := &usuarioRepositoryFake{
		usuario:      &domain.Usuario{Id: "u1", Role: domain.RoleAdmin, Status: domain.StatusAtivo, Version: 1},
		adminsAtivos: 2,
	}
	useCase := NewUseCase(repository)

	out, err := useCase.AlterarRole(context.Background(), "u1", "usuario")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Role != "usuario" {
		t.Fatalf("esperava role usuario, recebeu %s", out.Role)
	}
}

func TestAlterarStatusInativa(t *testing.T) {
	repository := &usuarioRepositoryFake{usuario: &domain.Usuario{Id: "u1", Role: domain.RoleUsuario, Status: domain.StatusAtivo, Version: 1}}
	useCase := NewUseCase(repository)

	out, err := useCase.AlterarStatus(context.Background(), "u1", "inativo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Status != "inativo" || repository.atualizado == nil {
		t.Fatalf("status nao atualizado: %+v", out)
	}
}

func TestAlterarStatusInvalidoRejeitaExcluido(t *testing.T) {
	repository := &usuarioRepositoryFake{usuario: &domain.Usuario{Id: "u1", Status: domain.StatusAtivo, Version: 1}}
	useCase := NewUseCase(repository)

	if _, err := useCase.AlterarStatus(context.Background(), "u1", "excluido"); !errors.Is(err, domain.ErrStatusInvalido) {
		t.Fatalf("esperava ErrStatusInvalido, recebeu %v", err)
	}
}

func TestAlterarStatusBloqueiaInativarUltimoAdmin(t *testing.T) {
	repository := &usuarioRepositoryFake{
		usuario:      &domain.Usuario{Id: "u1", Role: domain.RoleAdmin, Status: domain.StatusAtivo, Version: 1},
		adminsAtivos: 1,
	}
	useCase := NewUseCase(repository)

	if _, err := useCase.AlterarStatus(context.Background(), "u1", "inativo"); !errors.Is(err, domain.ErrUltimoAdmin) {
		t.Fatalf("esperava ErrUltimoAdmin, recebeu %v", err)
	}
	if repository.atualizado != nil {
		t.Fatal("nao deveria inativar o ultimo admin")
	}
}
