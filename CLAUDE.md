# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Projeto: sped-nfe-go

Biblioteca Go para geração, assinatura e transmissão de Notas Fiscais Eletrônicas (NFe/NFCe) brasileiras. É uma alternativa ao projeto PHP nfephp-org/sped-nfe, com foco em performance, type safety e concorrência nativa do Go.

## Contexto do Projeto
Você é um especialista em desenvolvimento Go que vai me ajudar a criar o sped-nfe-go, um pacote Go para Nota Fiscal Eletrônica brasileira. Este projeto vai converter o projeto PHP nfephp-org/sped-nfe para Go, mantendo 100% da funcionalidade.

### Objetivo Principal
Criar um pacote Go que desenvolvedores possam instalar com:
bashgo get github.com/adrianodrix/sped-nfe-go
E usar em suas aplicações Go para:

- Gerar XMLs de NFe/NFCe conforme layouts oficiais SEFAZ
- Assinar digitalmente com certificados A1/A3 ICP-Brasil
- Transmitir para webservices SEFAZ
- Processar eventos fiscais (cancelamento, carta de correção)

## Estrutura de Referência (PHP Original)
Você deve usar a estrutura e arquivos PHP do projeto nfephp-org/sped-nfe como referência. O projeto se encontra em `refs/ssped-nfe`. Sua tarefa é converter a funcionalidade mantendo:

1. Mesma lógica de negócio
2. Mesma estrutura de dados
3. Mesma validação
4. Compatibilidade total com schemas XSD

## Comandos de Desenvolvimento

### Testes
```bash
# Executar todos os testes
go test ./...

# Testes com saída verbosa
go test -v ./...

# Testes com coverage
go test -cover ./...

# Executar testes de benchmark
go test -bench=. ./...
```

### Build e Linting
```bash
# Build do projeto
go build ./...

# Verificar dependências
go mod download
go mod tidy

# Verificar formatting
go fmt ./...

# Verificar vet (análise estática básica)
go vet ./...
```

## Arquitetura do Código

### Estrutura Principal
O projeto está organizado seguindo o padrão Go idiomático, baseado na estrutura do projeto PHP original:

- **`nfe/`** - Pacote principal com funcionalidades core NFe (make.go, tools.go, types.go)
- **`common/`** - Código base compartilhado (tools.go, config.go, webservices.go)  
- **`factories/`** - Fábricas e helpers (parser.go, qrcode.go, contingency.go)
- **`certificate/`** - Gestão de certificados digitais A1/A3
- **`soap/`** - Cliente SOAP para comunicação com SEFAZ
- **`utils/`** - Utilitários brasileiros (CNPJ, CPF, chaves de acesso)
- **`examples/`** - Exemplos de uso da biblioteca
- **`docs/`** - Documentação técnica e estrutural
- **`testdata/`** - Dados para testes
- **`refs/sped-nfe/`** - Referência do projeto PHP original

### Padrões de Conversão PHP → Go
O projeto mantém compatibilidade conceitual com o nfephp-org/sped-nfe:

- Classes PHP → Types/Structs Go
- Namespaces PHP → Packages Go  
- Métodos PHP → Functions/Methods Go com error handling idiomático
- Arrays associativos PHP → Structs tipadas Go
- Exceções PHP → Error returns Go

### Estado Atual
**⚠️ PROJETO EM DESENVOLVIMENTO INICIAL**
- Apenas estrutura básica implementada em `nfe/nfe.go`
- Funcionalidades principais ainda não implementadas
- Não está pronto para uso em produção

### Prioridades de Implementação
1. **Fase 1**: Estruturas de dados (types.go) e configuração básica
2. **Fase 2**: Geração de NFe (make.go) e certificados digitais  
3. **Fase 3**: Comunicação SEFAZ (tools.go, soap/)
4. **Fase 4**: Recursos extras (parser, QRCode, utilidades)

## Convenções e Padrões

### Go Idiomático
- Seguir convenções padrão Go (gofmt, go vet)
- Error handling explícito com returns
- Interfaces pequenas e composáveis
- Estruturas de dados tipadas
- Documentação em comentários GoDoc

### Certificados e Segurança
- Suporte a certificados ICP-Brasil A1 (.pfx) e A3 (PKCS#11)
- Assinatura digital XML com xmldsig
- Comunicação segura TLS com webservices SEFAZ

### Testes
- Testes unitários em arquivos `*_test.go`
- Dados de teste em `testdata/`
- Benchmarks para performance critical paths
- Mocks para webservices SEFAZ em testes

### Compliance SEFAZ
- Aderência total aos schemas XSD oficiais (em `refs/sped-nfe/schemes/`)
- Suporte aos layouts NFe 3.10 e 4.00
- Validação completa de dados antes da transmissão
- Tratamento correto de ambientes (homologação/produção)

## Regras de Conversão PHP → Go

### 1. Classes PHP → Structs/Interfaces Go
```php
// PHP
class Make {
    protected $nfe;
    public function tagInfNFe($chave) { }
}
```

```go
// Go
type Make struct {
    nfe *NFe
}

func (m *Make) TagInfNFe(chave string) error { }
```

### 2. Arrays PHP → Slices Go
```php
// PHP
protected $det = [];
```

```go
// Go
type InfNFe struct {
    Det []Item `xml:"det" json:"det"`
}
```

### 3. Validação PHP → Struct Tags Go
```php
// PHP
if (empty($chave) || strlen($chave) != 44) {
    throw new Exception();
}
```

```go
// Go
type InfNFe struct {
    ID string `xml:"Id,attr" json:"id" validate:"required,len=44"`
}
```

### 4. XML PHP → encoding/xml Go
```php
// PHP
$this->dom->createElement("infNFe");
```

```go
// Go
type NFe struct {
    XMLName xml.Name `xml:"NFe"`
    InfNFe  InfNFe   `xml:"infNFe"`
}
```

## Dependências Obrigatórias

```go
require (
    github.com/lafriks/go-xmldsig v0.2.0        // Assinatura XML
    github.com/ThalesIgnite/crypto11 v1.2.5    // PKCS#11 (A3)
    software.sslmate.com/src/go-pkcs12 v0.2.0  // PKCS#12 (A1)
    github.com/beevik/etree v1.1.0              // XML processing
    github.com/go-playground/validator/v10      // Validation
    golang.org/x/net v0.17.0                    // HTTP/SOAP
)
```

## API Final Esperada

```go
package main

import (
    "github.com/adrianodrix/sped-nfe-go/nfe"
    "github.com/adrianodrix/sped-nfe-go/certificate"
)

func main() {
    // Configurar cliente
    client, err := nfe.New(nfe.Config{
        Environment: nfe.Production,
        UF:         nfe.SP,
    })
    
    // Carregar certificado
    cert, err := certificate.LoadA1("/path/cert.pfx", "senha")
    client.SetCertificate(cert)
    
    // Criar NFe
    make := client.CreateNFe()
    make.TagInfNFe(chave, "4.00")
    make.TagIde(identificacao)
    make.TagEmit(emitente)
    make.TagDest(destinatario)
    make.TagDet(item)
    make.TagTotal(total)
    
    // Gerar XML e transmitir
    xml, err := make.GetXML()
    response, err := client.Authorize(xml)
}
```

## Instruções Específicas para o Agente

### Como escolher o proximo arquivo para converter?
Leia a estrutura do projeto e rastreie as dependencias de cada arquivo, e comece pelo arquivo que as dependencias ja estao prontas para importar.

### Quando for converter um arquivo PHP:

1. **Analise a estrutura** da classe PHP
2. **Identifique métodos públicos** → viram funções exportadas
3. **Identifique propriedades** → viram campos de struct
4. **Converta validações** para struct tags
5. **Mantenha nomes** similares (camelCase em Go)
6. **Preserve lógica** exata de validação e processamento

### Estruturas XML:

1. **Use struct tags** para mapeamento XML exato
2. **Mantenha namespaces** quando necessário
3. **Preserve atributos** XML como `xml:",attr"`
4. **Use ponteiros** para campos opcionais

### Tratamento de Erros:

1. **Retorne erros** explícitos em todas as funções
2. **Use fmt.Errorf** para mensagens descritivas
3. **Valide inputs** antes de processar
4. **Mantenha mensagens** de erro similares ao PHP

### Performance:

1. **Use sync.Pool** para objetos reutilizáveis
2. **Evite alocações** desnecessárias
3. **Use context.Context** para operações longas
4. **Implemente timeouts** em requisições HTTP

## Exemplo de Tarefa

**Input**: Arquivo `Make.php` do sped-nfe
**Output Esperado**: 
- `nfe/make.go` com todas as funções convertidas
- `nfe/types.go` com structs correspondentes
- Testes unitários básicos
- Exemplo de uso

## Prioridades

1. **Funcionalidade correta** (100% compatível)
2. **Type safety** (validação em compile time)
3. **Performance** (aproveitar concorrência Go)
4. **Facilidade de uso** (API intuitiva)
5. **Documentação** (GoDoc em todas as funções)

## Workflow de Desenvolvimento das Issues

### 🔄 Git Flow e Versionamento

Para cada issue implementada, seguir rigorosamente este fluxo:

#### 1. Iniciar Feature
```bash
# Criar feature branch para a issue
git flow feature start issue-{numero}

# Exemplo para issue #1:
git flow feature start issue-1
```

#### 2. Assignar Issue
- Assignar a issue para **adrianodrix** no GitHub
- Mover issue para "In Progress" se usando Projects

#### 3. Desenvolvimento
- Implementar um arquivo por vez
- **Commit individual** para cada arquivo finalizado
- **Formato do commit**: `#{numero}: {descrição do arquivo/funcionalidade}`

```bash
# Exemplos de commits para issue #1:
git commit -m "#1: Add basic error types and validation"
git commit -m "#1: Implement Config struct with validation"
git commit -m "#1: Add constants for environments and UF codes"
git commit -m "#1: Add unit tests for config validation"
```

#### 4. Finalizar Feature
```bash
# Finalizar feature branch
git flow feature finish issue-{numero}

# Exemplo:
git flow feature finish issue-1
```

#### 5. Fechar Issue
- Fechar a issue no GitHub
- Marcar como concluída com comentário do que foi implementado

### 📋 Padrões de Commit

#### Estrutura Obrigatória:
```
#{issue_number}: {descrição_clara}

Exemplos:
#1: Add Config struct with environment validation
#1: Implement error types for better error handling
#2: Add CNPJ validation function with unit tests
#5: Create NFe struct with XML tags
```

#### Tipos de Commit:
- **Arquivos principais**: `#{issue}: Add {filename} with {functionality}`
- **Testes**: `#{issue}: Add tests for {functionality}`
- **Documentação**: `#{issue}: Add documentation for {component}`
- **Correções**: `#{issue}: Fix {problem} in {component}`

### 🎯 Ordem de Implementação

As issues devem ser implementadas na ordem de dependências:

1. **Issue #1**: Base e configuração (sem dependências)
2. **Issue #2**: Utilitários brasileiros (depende #1)
3. **Issue #3**: Cliente SOAP (depende #1)
4. **Issue #4**: Certificados digitais (depende #1)
5. **Issue #5**: Estruturas de dados (depende #1, #2)
6. **Issue #6**: Geração XML (depende #2, #5)
7. **Issue #7**: Comunicação SEFAZ (depende #3, #4, #6)
8. **Issue #8**: Conversão TXT (depende #5, #6)
9. **Issue #9**: Funcionalidades complementares (depende #5, #7)
10. **Issue #10**: Cliente principal (depende todas anteriores)

### ✅ Critérios para Finalizar Issue

Uma issue só pode ser finalizada quando:

- [ ] Todos os arquivos especificados foram implementados
- [ ] Todos os testes unitários passam
- [ ] Cobertura de testes ≥ 80%
- [ ] Código formatado com `go fmt`
- [ ] Sem warnings de `go vet`
- [ ] Documentação GoDoc completa
- [ ] Commits seguem o padrão estabelecido

### 🚀 Comandos de Verificação

Antes de finalizar qualquer feature:

```bash
# Formatar código
go fmt ./...

# Verificar análise estática
go vet ./...

# Executar todos os testes
go test ./...

# Verificar cobertura
go test -cover ./...

# Build para verificar compilação
go build ./...
```

## Formato de Resposta

Para cada arquivo convertido, forneça:

```
## Arquivo: nfe/make.go

```go
// Código Go aqui
```

## Explicação das Conversões:
- Método X virou função Y
- Validação Z virou struct tag W
- etc.

## Testes:
```go
// Testes unitários básicos
```

## Uso:
```go
// Exemplo prático de uso
```
```

---
