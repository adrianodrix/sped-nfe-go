# Estrutura do sped-nfe-go baseada no nfephp-org/sped-nfe

## 📁 Estrutura Original PHP (nfephp-org/sped-nfe)

```
nfephp-org/sped-nfe/
├── src/
│   ├── Make.php                    # Classe principal para geração de NFe
│   ├── Tools.php                   # Comunicação com SEFAZ (extends Common/Tools)
│   ├── Complements.php             # Protocolação e complementos
│   ├── Common/                     # Classes base comuns
│   │   ├── Tools.php              # Classe base para comunicação SEFAZ
│   │   ├── Config.php             # Validação de configuração
│   │   ├── Standardize.php        # Conversão e identificação de XMLs
│   │   ├── ValidTXT.php           # Validação de arquivos TXT
│   │   └── Webservices.php        # URLs e configurações dos webservices
│   ├── Factories/                  # Fábricas e helpers
│   │   ├── Parser.php             # Conversão TXT para XML
│   │   ├── QRCode.php             # Geração de QRCode para NFCe
│   │   ├── Header.php             # Headers SOAP
│   │   ├── Contingency.php        # Contingência
│   │   └── ContingencyNFe.php     # Contingência específica NFe
│   ├── Traits/                     # Traits reutilizáveis
│   │   └── TraitEPECNfce.php      # EPEC para NFCe
│   └── Exception/                  # Exceções customizadas
│       └── DocumentsException.php  # Exceções de documentos
├── docs/                          # Documentação
├── examples/                      # Exemplos de uso
├── storage/                       # Schemas XSD e estruturas
│   ├── xmlschemes/               # Schemas XSD oficiais
│   └── txtstructure*.json        # Estruturas TXT
└── tests/                        # Testes unitários
```

## 🎯 Estrutura Proposta Go (sped-nfe-go)

```go
github.com/adrianodrix/sped-nfe-go/
├── internal/                      # Código interno (não exportado)
│   ├── schemas/                  # XSD schemas (cópia do storage/)
│   └── webservices/             # URLs e configurações SEFAZ
├── pkg/                          # Código público/exportável (opcional)
├── nfe/                         # Pacote principal (equivale ao src/)
│   ├── make.go                  # Conversão de Make.php
│   ├── tools.go                 # Conversão de Tools.php
│   ├── complements.go           # Conversão de Complements.php
│   ├── types.go                 # Estruturas de dados (structs)
│   ├── client.go                # Cliente principal (wrapper)
│   └── config.go                # Configuração
├── common/                      # Conversão de src/Common/
│   ├── tools.go                 # Conversão de Common/Tools.php
│   ├── config.go                # Conversão de Common/Config.php
│   ├── standardize.go           # Conversão de Common/Standardize.php
│   ├── validator.go             # Validações
│   └── webservices.go           # Conversão de Common/Webservices.php
├── factories/                   # Conversão de src/Factories/
│   ├── parser.go                # Conversão de Parser.php
│   ├── qrcode.go                # Conversão de QRCode.php
│   ├── header.go                # Conversão de Header.php
│   └── contingency.go           # Conversão de Contingency.php
├── certificate/                 # Certificados digitais (novo)
│   ├── certificate.go           # Interface principal
│   ├── a1.go                    # Certificados A1 (.pfx)
│   └── a3.go                    # Certificados A3 (PKCS#11)
├── soap/                        # Cliente SOAP (novo)
│   ├── client.go                # Cliente HTTP/SOAP
│   ├── envelope.go              # Envelope SOAP
│   └── security.go              # WS-Security
├── utils/                       # Utilitários brasileiros (novo)
│   ├── cnpj.go                  # Validação CNPJ
│   ├── cpf.go                   # Validação CPF
│   ├── keys.go                  # Chave de acesso NFe
│   └── xml.go                   # Helpers XML
├── examples/                    # Exemplos de uso
│   ├── nfe-simples/            # NFe básica
│   ├── nfce/                   # NFCe
│   └── eventos/                # Eventos fiscais
├── docs/                       # Documentação
├── testdata/                   # Dados para testes
└── tests/                      # Testes (se necessário)
```

## 🔄 Mapeamento de Conversão

### Principais Classes PHP → Go

| PHP | Go | Descrição |
|-----|----|--------------|
| `src/Make.php` | `nfe/make.go` | Geração de NFe/NFCe |
| `src/Tools.php` | `nfe/tools.go` | Comunicação SEFAZ |
| `src/Common/Tools.php` | `common/tools.go` | Classe base |
| `src/Complements.php` | `nfe/complements.go` | Protocolação |
| `src/Factories/Parser.php` | `factories/parser.go` | TXT→XML |
| `src/Factories/QRCode.php` | `factories/qrcode.go` | QRCode NFCe |

### Namespaces PHP → Packages Go

| Namespace PHP | Package Go | Import |
|---------------|------------|---------|
| `NFePHP\NFe\` | `nfe` | `github.com/adrianodrix/sped-nfe-go/nfe` |
| `NFePHP\NFe\Common\` | `common` | `github.com/adrianodrix/sped-nfe-go/common` |
| `NFePHP\NFe\Factories\` | `factories` | `github.com/adrianodrix/sped-nfe-go/factories` |

## 📊 Organização por Responsabilidade

### 1. **nfe/** - Core NFe
```go
// make.go - Geração de NFe
type Make struct {
    xml     string
    errors  []error
    version string
}

func (m *Make) TagIde(ide Identificacao) error
func (m *Make) TagEmit(emit Emitente) error  
func (m *Make) GetXML() ([]byte, error)

// tools.go - Comunicação SEFAZ  
type Tools struct {
    config      Config
    certificate Certificate
    soapClient  soap.Client
}

func (t *Tools) Authorize(xml []byte) (*AuthResponse, error)
func (t *Tools) Query(chave string) (*QueryResponse, error)
```

### 2. **common/** - Base Comum
```go
// tools.go - Classe base
type BaseTools struct {
    urlPortal  string
    urlVersion string
    environment int
}

// standardize.go - Conversão XML
func WhichIs(xml []byte) (string, error)
func ToStdClass(xml []byte) (*StandardResponse, error)
```

### 3. **factories/** - Helpers
```go
// parser.go - TXT para XML
type Parser struct {
    structure map[string]interface{}
    make      *Make
}

func (p *Parser) ToXML(lines []string) ([]byte, error)

// qrcode.go - QRCode NFCe
func PutQRTag(dom *xml.Document, token, idToken string) error
```

### 4. **types.go** - Estruturas de Dados
```go
// Estruturas principais
type NFe struct {
    XMLName xml.Name `xml:"NFe"`
    InfNFe  InfNFe   `xml:"infNFe"`
}

type InfNFe struct {
    ID     string        `xml:"Id,attr"`
    Versao string        `xml:"versao,attr"`
    Ide    Identificacao `xml:"ide"`
    Emit   Emitente      `xml:"emit"`
    Dest   *Destinatario `xml:"dest,omitempty"`
    Det    []Item        `xml:"det"`
    Total  Total         `xml:"total"`
}
```

## 🚀 Exemplos de Conversão

### Make.php → make.go
```php
// PHP
class Make {
    public function tagIde($std) {
        // ... lógica PHP
    }
}
```

```go
// Go
type Make struct {
    xml string
    dom *xml.Document
}

func (m *Make) TagIde(ide Identificacao) error {
    // ... lógica Go equivalente
    return nil
}
```

### Tools.php → tools.go
```php
// PHP  
class Tools extends Common\Tools {
    public function sefazAutorizar($xml, $idLote = null) {
        // ... lógica PHP
    }
}
```

```go
// Go
type Tools struct {
    *common.BaseTools // Embedded struct (herança)
}

func (t *Tools) Authorize(xml []byte, idLote string) (*AuthResponse, error) {
    // ... lógica Go equivalente
    return nil, nil
}
```

## 🎯 Prioridade de Conversão

### Fase 1: Fundação
1. **common/tools.go** - Base para tudo
2. **nfe/types.go** - Estruturas de dados
3. **nfe/config.go** - Configuração

### Fase 2: Core  
1. **nfe/make.go** - Geração de NFe
2. **certificate/** - Certificados digitais
3. **soap/** - Cliente SOAP

### Fase 3: Comunicação
1. **nfe/tools.go** - Webservices
2. **common/webservices.go** - URLs SEFAZ
3. **nfe/complements.go** - Protocolação

### Fase 4: Extras
1. **factories/parser.go** - TXT→XML
2. **factories/qrcode.go** - QRCode
3. **utils/** - Utilitários brasileiros

## 📋 Comando para Criar Estrutura

```bash
# Criar estrutura de pastas
mkdir -p nfe common factories certificate soap utils examples docs testdata

# Arquivos iniciais essenciais
touch nfe/{make.go,tools.go,types.go,config.go,client.go}
touch common/{tools.go,config.go,standardize.go,webservices.go}
touch factories/{parser.go,qrcode.go,contingency.go}
touch certificate/{certificate.go,a1.go,a3.go}
touch soap/{client.go,envelope.go}
touch utils/{cnpj.go,cpf.go,keys.go,xml.go}

# go.mod
echo 'module github.com/adrianodrix/sped-nfe-go

go 1.21

require (
    github.com/lafriks/go-xmldsig v0.2.0
    github.com/ThalesIgnite/crypto11 v1.2.5
    software.sslmate.com/src/go-pkcs12 v0.2.0
    github.com/beevik/etree v1.1.0
    github.com/go-playground/validator/v10 v10.16.0
)' > go.mod
```

Esta estrutura mantém **100% de compatibilidade** com a organização do projeto PHP original, facilitando a conversão incremental e manutenção da funcionalidade! 🎯