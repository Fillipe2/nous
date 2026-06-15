package dto

type LoginInput struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type LoginOutput struct {
	Token        string                 `json:"token"`
	RefreshToken string                 `json:"refresh_token"`
	Usuario      interface{}            `json:"usuario"`
	Negocios     []*NegocioBasicoOutput `json:"negocios,omitempty"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutInput struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

type EsqueciSenhaInput struct {
	Email string `json:"email"`
}

type EsqueciSenhaOutput struct {
	Mensagem   string `json:"mensagem"`
	ResetToken string `json:"reset_token,omitempty"`
}

type RedefinirSenhaInput struct {
	Token     string `json:"token"`
	NovaSenha string `json:"nova_senha"`
}

type SolicitarVerificacaoEmailInput struct {
	Email string `json:"email"`
}

type SolicitarVerificacaoEmailOutput struct {
	Mensagem         string `json:"mensagem"`
	VerificacaoToken string `json:"verificacao_token,omitempty"`
}

type VerificarEmailInput struct {
	Token string `json:"token"`
}
