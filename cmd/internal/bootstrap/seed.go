package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nous/internal/domain/usuario"

	"golang.org/x/crypto/bcrypt"
)

func (c *Container) seedAdmin(ctx context.Context) error {
	if c.Config.AdminEmail == "" || c.Config.AdminSenha == "" {
		return nil
	}

	admin, err := c.UsuarioRepository.BuscarPorEmail(ctx, c.Config.AdminEmail)
	if err != nil && !errors.Is(err, usuario.ErrNaoEncontrado) {
		return fmt.Errorf("buscar admin inicial: %w", err)
	}

	if admin != nil {
		alterado := false
		if admin.Role != usuario.RoleAdmin {
			admin.Role = usuario.RoleAdmin
			alterado = true
		}
		if admin.Status != usuario.StatusAtivo {
			admin.Status = usuario.StatusAtivo
			alterado = true
		}
		if !admin.EmailVerified {
			admin.EmailVerified = true
			alterado = true
		}
		if alterado {
			admin.AtualizadoEm = time.Now()
			admin.Version++
			if err := c.UsuarioRepository.Atualizar(ctx, admin); err != nil {
				return fmt.Errorf("atualizar admin inicial: %w", err)
			}
		}
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(c.Config.AdminSenha), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gerar senha admin inicial: %w", err)
	}

	now := time.Now()
	admin = &usuario.Usuario{
		Nome:          c.Config.AdminNome,
		Email:         c.Config.AdminEmail,
		SenhaHash:     string(hash),
		Role:          usuario.RoleAdmin,
		Status:        usuario.StatusAtivo,
		EmailVerified: true,
		CriadoEm:      now,
		AtualizadoEm:  now,
		Version:       1,
	}

	if err := c.UsuarioRepository.Salvar(ctx, admin); err != nil {
		return fmt.Errorf("criar admin inicial: %w", err)
	}

	return nil
}
