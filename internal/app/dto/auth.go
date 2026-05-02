package dto

type LoginInput struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type LoginOutput struct {
	Token        string                 `json:"token"`
	RefreshToken string                 `json:"refresh_token"`
	Usuario      interface{}            `json:"usuario"`
	Clientes     []*ClienteBasicoOutput `json:"clientes,omitempty"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutInput struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}
