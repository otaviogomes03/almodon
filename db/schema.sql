-- Esquema do Banco de Dados - Projeto Almodon
-- Dialeto: PostgreSQL

-- 1. Tabela de Usuários
CREATE TABLE usuarios (
    siape VARCHAR(20) PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    senha_hash VARCHAR(255) NOT NULL,
    perfil VARCHAR(50) NOT NULL, -- Ex: 'ADMIN', 'BOLSISTA'
    data_criacao TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2. Entidades Externas
CREATE TABLE fornecedores (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    cnpj VARCHAR(20) NOT NULL UNIQUE,
    contato VARCHAR(255)
);

CREATE TABLE clinicas (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL
);

CREATE TABLE laboratorios (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL
);

-- 3. Produtos
CREATE TABLE produtos (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    codigo_ecampus VARCHAR(50),
    siads VARCHAR(50),
    catmat VARCHAR(50),
    estoque_minimo INTEGER NOT NULL DEFAULT 0,
    unidade VARCHAR(20) NOT NULL -- Ex: 'CX', 'UN', 'LITRO'
);

-- 4. Lotes
CREATE TABLE lotes (
    id SERIAL PRIMARY KEY,
    id_produto INTEGER NOT NULL REFERENCES produtos(id),
    id_fornecedor INTEGER NOT NULL REFERENCES fornecedores(id),
    codigo_lote VARCHAR(100) NOT NULL,
    data_validade DATE NOT NULL,
    valor_unitario DECIMAL(10, 2) NOT NULL,
    quantidade_atual INTEGER NOT NULL DEFAULT 0 CHECK (quantidade_atual >= 0),
    data_entrada TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 5. Solicitações
CREATE TABLE solicitacoes (
    id SERIAL PRIMARY KEY,
    id_clinica INTEGER REFERENCES clinicas(id),
    id_laboratorio INTEGER REFERENCES laboratorios(id),
    id_usuario_aprovador VARCHAR(20) REFERENCES usuarios(siape),
    data_solicitacao TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'PENDENTE',
    motivo TEXT
);

-- Regra de Integridade: A solicitação deve pertencer a uma Clínica OU a um Laboratório (XOR)
ALTER TABLE solicitacoes ADD CONSTRAINT check_origem_solicitacao
CHECK (
    (id_clinica IS NOT NULL AND id_laboratorio IS NULL) OR 
    (id_clinica IS NULL AND id_laboratorio IS NOT NULL)
);

CREATE TABLE itens_solicitacao (
    id SERIAL PRIMARY KEY,
    id_solicitacao INTEGER NOT NULL REFERENCES solicitacoes(id) ON DELETE CASCADE,
    id_produto INTEGER NOT NULL REFERENCES produtos(id),
    quantidade_solicitada INTEGER NOT NULL CHECK (quantidade_solicitada > 0)
);

-- 6. Transações (Log de Auditoria e Movimentação)
CREATE TABLE transacoes (
    id SERIAL PRIMARY KEY,
    id_lote INTEGER NOT NULL REFERENCES lotes(id),
    id_usuario VARCHAR(20) NOT NULL REFERENCES usuarios(siape),
    tipo_transacao VARCHAR(20) NOT NULL, -- 'ENTRADA', 'SAIDA', 'AJUSTE', 'PERDA'
    quantidade INTEGER NOT NULL, 
    data_hora TIMESTAMP NOT NULL DEFAULT NOW()
);