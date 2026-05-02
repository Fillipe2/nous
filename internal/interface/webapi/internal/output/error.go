// Package output define estruturas de resposta HTTP da API REST.
package output

// ErrorResponse encapsula uma resposta de erro da API.
type ErrorResponse struct {
	Message string `json:"message,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// InvalidRequest cria resposta padrao para requisições inválidas.
func InvalidRequest(err error) ErrorResponse {
	response := ErrorResponse{Message: "requisicao invalida"}
	if err != nil {
		response.Error = err.Error()
	}
	return response
}
