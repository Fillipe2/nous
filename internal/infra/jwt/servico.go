package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenInvalido = errors.New("token invalido")
	ErrTokenExpirado = errors.New("token expirado")
)

const (
	duracaoAccess  = 2 * time.Hour
	duracaoRefresh = 7 * 24 * time.Hour
)

type Claims struct {
	UsuarioId string `json:"usuario_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	jwtlib.RegisteredClaims
}

type Servico struct {
	secretAccess  []byte
	secretRefresh []byte
}

func NewServico(secretAccess, secretRefresh string) *Servico {
	return &Servico{
		secretAccess:  []byte(secretAccess),
		secretRefresh: []byte(secretRefresh),
	}
}

func (s *Servico) GerarTokens(usuarioId, email, role string) (accessToken, refreshToken string, err error) {
	accessToken, err = s.gerarToken(usuarioId, email, role, duracaoAccess, s.secretAccess)
	if err != nil {
		return "", "", fmt.Errorf("gerar access token: %w", err)
	}

	refreshToken, err = s.gerarToken(usuarioId, email, role, duracaoRefresh, s.secretRefresh)
	if err != nil {
		return "", "", fmt.Errorf("gerar refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *Servico) ValidarToken(tokenString string) (*Claims, error) {
	return s.validarToken(tokenString, s.secretAccess)
}

func (s *Servico) ValidarRefreshToken(tokenString string) (usuarioId string, err error) {
	claims, err := s.validarToken(tokenString, s.secretRefresh)
	if err != nil {
		return "", err
	}
	return claims.UsuarioId, nil
}

func (s *Servico) gerarToken(usuarioId, email, role string, duracao time.Duration, secret []byte) (string, error) {
	claims := &Claims{
		UsuarioId: usuarioId,
		Email:     email,
		Role:      role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(duracao)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			Subject:   usuarioId,
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (s *Servico) validarToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(t *jwtlib.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo de assinatura inesperado: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrTokenExpirado
		}
		return nil, ErrTokenInvalido
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalido
	}

	return claims, nil
}
