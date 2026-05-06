package jwt

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
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

// JWK representa uma chave pública no formato JSON Web Key.
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSResponse é o envelope do endpoint /.well-known/jwks.json.
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type Servico struct {
	secretAccess  []byte
	secretRefresh []byte
	rsaKey        *rsa.PrivateKey
	rsaKid        string
}

// NewServico cria um Servico JWT com assinatura HS256.
func NewServico(secretAccess, secretRefresh string) *Servico {
	return &Servico{
		secretAccess:  []byte(secretAccess),
		secretRefresh: []byte(secretRefresh),
	}
}

// NewServicoRSA cria um Servico JWT com assinatura RS256.
// Se rsaKey for nil, comporta-se igual ao NewServico (HS256).
func NewServicoRSA(secretAccess, secretRefresh string, rsaKey *rsa.PrivateKey) *Servico {
	s := NewServico(secretAccess, secretRefresh)
	if rsaKey != nil {
		s.rsaKey = rsaKey
		s.rsaKid = derivarKeyID(rsaKey)
	}
	return s
}

// KidRSA retorna o Key ID da chave RSA configurada, ou string vazia se HS256.
func (s *Servico) KidRSA() string {
	return s.rsaKid
}

// JWKS retorna o conjunto de chaves públicas para validação externa de tokens RS256.
// Retorna false quando o serviço usa HS256 (sem chave RSA configurada).
func (s *Servico) JWKS() (JWKSResponse, bool) {
	if s.rsaKey == nil {
		return JWKSResponse{}, false
	}
	pub := &s.rsaKey.PublicKey

	eBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(eBytes, uint32(pub.E))
	for len(eBytes) > 1 && eBytes[0] == 0 {
		eBytes = eBytes[1:]
	}

	return JWKSResponse{
		Keys: []JWK{{
			Kty: "RSA",
			Use: "sig",
			Alg: "RS256",
			Kid: s.rsaKid,
			N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(eBytes),
		}},
	}, true
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

func (s *Servico) gerarToken(usuarioId, email, role string, duracao time.Duration, hmacSecret []byte) (string, error) {
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

	if s.rsaKey != nil {
		token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
		token.Header["kid"] = s.rsaKid
		return token.SignedString(s.rsaKey)
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(hmacSecret)
}

func (s *Servico) validarToken(tokenString string, hmacSecret []byte) (*Claims, error) {
	var keyFunc jwtlib.Keyfunc
	if s.rsaKey != nil {
		keyFunc = func(t *jwtlib.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwtlib.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("metodo de assinatura inesperado: %v", t.Header["alg"])
			}
			return &s.rsaKey.PublicKey, nil
		}
	} else {
		keyFunc = func(t *jwtlib.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metodo de assinatura inesperado: %v", t.Header["alg"])
			}
			return hmacSecret, nil
		}
	}

	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, keyFunc)
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

func derivarKeyID(key *rsa.PrivateKey) string {
	h := sha256.Sum256(key.PublicKey.N.Bytes())
	return fmt.Sprintf("%x", h[:8])
}
