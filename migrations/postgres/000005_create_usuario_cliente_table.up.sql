CREATE TABLE usuario_cliente (
    usuario_id UUID NOT NULL REFERENCES usuario(id) ON DELETE CASCADE,
    cliente_id UUID NOT NULL REFERENCES cliente(id) ON DELETE CASCADE,
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (usuario_id, cliente_id)
);
