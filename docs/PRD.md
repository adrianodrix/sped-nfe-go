# Product Requirements Document (PRD)
# sped-nfe-go: Biblioteca de Nota Fiscal Eletrônica para Go

**Versão:** 1.0  
**Data:** 02 de Julho de 2025  
**Autor:** Adriano Santos (@adrianodrix)  
**Status:** Draft  

---

## 1. Visão Geral do Produto

### 1.1 Resumo Executivo

O **sped-nfe-go** é um pacote Go para geração, assinatura, validação e transmissão de Notas Fiscais Eletrônicas (NFe/NFCe) brasileiras. O projeto visa criar uma biblioteca robusta, performática e idiomática que desenvolvedores possam importar diretamente em suas aplicações Go com `go get github.com/adrianodrix/sped-nfe-go`, oferecendo uma alternativa ao consolidado **nfephp-org/sped-nfe** (PHP).

### 1.2 Problema

O ecossistema Go brasileiro carece de um pacote completo para NFe que atenda aos padrões oficiais da SEFAZ. As alternativas existentes são limitadas:
- **frones/nfe**: Foca apenas em consultas, sem geração completa
- **webmaniabr/NFe-Go**: Dependente de API comercial paga
- Ausência de bibliotecas para geração de DANFE, eventos fiscais e contingência

### 1.3 Oportunidade

Criar o primeiro pacote Go completo para NFe representa uma oportunidade significativa de:
- Modernizar sistemas fiscais brasileiros com performance superior
- Oferecer alternativa open source robusta para desenvolvedores Go
- Estabelecer padrão de excelência para pacotes fiscais em Go
- Facilitar adoção do Go em fintechs e ERPs brasileiros

---

## 2. Objetivos do Produto

### 2.1 Objetivos Primários

1. **Paridade Funcional**: Replicar 100% das funcionalidades do sped-nfe PHP
2. **Performance Superior**: Alcançar 5-10x melhor performance que a versão PHP
3. **Type Safety**: Garantir validação em tempo de compilação
4. **Facilidade de Uso**: API intuitiva e documentação abrangente

### 2.2 Objetivos Secundários

1. **Concorrência Nativa**: Processamento paralelo de múltiplas NFe
2. **Facilidade de Importação**: Instalação simples com `go get github.com/adrianodrix/sped-nfe-go`
3. **Comunidade Ativa**: Estabelecer base de contribuidores brasileiros
4. **Integração CI/CD**: Testes automatizados contra ambientes SEFAZ

### 2.3 Métricas de Sucesso

- **Performance**: Geração de NFe 5x mais rápida que PHP
- **Adoção**: 100+ stars no GitHub em 6 meses, 1000+ downloads via `go get`
- **Qualidade**: Cobertura de testes >90%
- **Compatibilidade**: 100% conformidade com schemas oficiais SEFAZ

---

## 3. Análise de Mercado

### 3.1 Landscape Atual

#### Soluções PHP (Dominantes)
- **nfephp-org/sped-nfe**: 3.2k stars, padrão de mercado
- **nfephp-org/nfephp**: 1.8k stars, biblioteca mais ampla

#### Soluções Go (Limitadas)
- **frones/nfe**: Consultas apenas, arquitetura sólida, importável via `go get`
- **webmaniabr/NFe-Go**: Cliente para API comercial, funcionalidade completa mas dependente

#### Lacunas Identificadas
- Geração nativa de XML NFe/NFCe em pacote Go puro
- Biblioteca de assinatura digital brasileira standalone
- Geração de DANFE em Go puro
- Suporte completo a eventos fiscais em pacote importável

### 3.2 Vantagem Competitiva

1. **Performance**: Go naturalmente 5-10x mais rápido que PHP
2. **Concorrência**: Goroutines para processamento paralelo
3. **Type Safety**: Prevenção de erros em tempo de compilação
4. **Simplicidade**: Importação direta com `go get`, sem dependências externas complexas
5. **Memória**: Garbage collector eficiente

---

## 4. Personas e Casos de Uso

### 4.1 Personas Primárias

#### Desenvolvedor Backend (85% dos usuários)
- **Perfil**: Desenvolvedor pleno/sênior, 3-8 anos experiência
- **Necessidades**: Pacote simples de importar, documentação clara, performance
- **Dores**: Complexidade fiscal brasileira, certificados digitais, dependências

#### Arquiteto de Software (10% dos usuários)
- **Perfil**: Sênior/especialista, decisor técnico
- **Necessidades**: Escalabilidade, manutenibilidade, arquitetura limpa
- **Dores**: Vendor lock-in, performance em alto volume

#### Freelancer/Consultor (5% dos usuários)
- **Perfil**: Desenvolvedor independente, múltiplos projetos
- **Necessidades**: Importação rápida com `go get`, documentação completa
- **Dores**: Tempo de implementação, suporte técnico, setup complexo

### 4.2 Casos de Uso Principais

1. **Emissão de NFe/NFCe**
   - Geração de XML conforme layout oficial
   - Assinatura digital com certificados A1/A3
   - Transmissão para SEFAZ estadual

2. **Consulta e Validação**
   - Status de serviços SEFAZ
   - Situação de NFe específica
   - Validação de chave de acesso

3. **Eventos Fiscais**
   - Carta de Correção Eletrônica (CCe)
   - Cancelamento de NFe
   - Manifestação do Destinatário

4. **Contingência**
   - EPEC (Evento Prévio de Emissão em Contingência)
   - FS-IA (Formulário de Segurança para Impressão Auxiliar)
   - SVC-AN/SVC-RS (SEFAZ Virtual de Contingência)

---

## 5. Especificações Funcionais

### 5.1 Módulos Core

#### 5.1.1 Certificate (Certificados Digitais)
```go
type Certificate interface {
    LoadA1(pfxPath, password string) error
    LoadA3(pkcs11Config PKCS11Config) error
    GetPrivateKey() crypto.PrivateKey
    GetCertificate() *x509.Certificate
    GetChain() []*x509.Certificate
}
```

**Funcionalidades:**
- Carregamento de certificados A1 (.pfx/.p12)
- Integração com tokens/HSM A3 via PKCS#11
- Validação de vencimento e cadeia de certificação
- Cache inteligente de certificados

#### 5.1.2 Make (Geração de NFe)
```go
type Make interface {
    TagInfNFe(chave, versao string) error
    TagIde(ide Identificacao) error
    TagEmit(emit Emitente) error
    TagDest(dest Destinatario) error
    TagDet(item Item) error
    TagTotal(total Total) error
    GetXML() ([]byte, error)
}
```

**Funcionalidades:**
- Builder pattern para construção de NFe
- Validação automática de campos obrigatórios
- Suporte a múltiplos layouts (3.10, 4.00)
- Geração de chave de acesso automática

#### 5.1.3 Sign (Assinatura Digital)
```go
type Signer interface {
    SignXML(xml []byte, cert Certificate) ([]byte, error)
    VerifySignature(xml []byte) error
    AddTimestamp(xml []byte) ([]byte, error)
}
```

**Funcionalidades:**
- Assinatura XML conforme padrão ICP-Brasil
- Suporte a múltiplos algoritmos (RSA-SHA1, RSA-SHA256)
- Validação de integridade de assinatura
- Timestamp opcional

#### 5.1.4 Tools (Ferramentas Auxiliares)
```go
type Tools interface {
    ValidateXML(xml []byte, xsdPath string) error
    CalculateDigitVerifier(chave string) string
    GenerateAccessKey(uf, ano, cnpj, modelo, serie, numero, tipoEmissao, codigo string) string
    AddProtocol(xml []byte, protocol Protocol) ([]byte, error)
}
```

#### 5.1.5 Webservices (Comunicação SEFAZ)
```go
type WebService interface {
    Authorize(xml []byte) (*AuthResponse, error)
    Query(chave string) (*QueryResponse, error)
    Cancel(xml []byte, justificativa string) (*CancelResponse, error)
    SendEvent(xml []byte) (*EventResponse, error)
    CheckService() (*StatusResponse, error)
}
```

### 5.2 Estruturas de Dados

#### 5.2.1 NFe Principal
```go
type NFe struct {
    XMLName xml.Name `xml:"NFe" json:"-"`
    InfNFe  InfNFe   `xml:"infNFe" json:"infNFe" validate:"required"`
    Signature *Signature `xml:"Signature,omitempty" json:"signature,omitempty"`
}

type InfNFe struct {
    ID       string        `xml:"Id,attr" json:"id" validate:"required,len=47"`
    Versao   string        `xml:"versao,attr" json:"versao" validate:"required"`
    Ide      Identificacao `xml:"ide" json:"ide" validate:"required"`
    Emit     Emitente      `xml:"emit" json:"emit" validate:"required"`
    Dest     *Destinatario `xml:"dest,omitempty" json:"dest,omitempty"`
    Det      []Item        `xml:"det" json:"det" validate:"required,min=1,max=990"`
    Total    Total         `xml:"total" json:"total" validate:"required"`
    Transp   *Transporte   `xml:"transp,omitempty" json:"transp,omitempty"`
    Cobr     *Cobranca     `xml:"cobr,omitempty" json:"cobr,omitempty"`
    Pag      *Pagamento    `xml:"pag,omitempty" json:"pag,omitempty"`
    InfAdic  *InfoAdic     `xml:"infAdic,omitempty" json:"infAdic,omitempty"`
}
```

### 5.3 API Pública

#### 5.3.1 Interface Principal
```go
package nfe

// Client é a interface principal para operações NFe
type Client struct {
    cert      Certificate
    signer    Signer
    validator Validator
    wsClient  WebServiceClient
    config    Config
}

func New(config Config) (*Client, error)
func (c *Client) LoadCertificate(certConfig CertConfig) error
func (c *Client) CreateNFe() *Make
func (c *Client) Authorize(nfe *NFe) (*AuthResponse, error)
func (c *Client) Query(chave string) (*QueryResponse, error)
func (c *Client) Cancel(chave, justificativa string) (*CancelResponse, error)
```

#### 5.3.2 Exemplo de Uso
```go
package main

import (
    "log"
    "time"
    
    "github.com/adrianodrix/sped-nfe-go/nfe"
)

func main() {
    // Configuração
    config := nfe.Config{
        Environment: nfe.Production,
        UF:         nfe.SP,
        Timeout:    30 * time.Second,
    }

    client, err := nfe.New(config)
    if err != nil {
        log.Fatal(err)
    }

    // Carrega certificado A1
    err = client.LoadCertificate(nfe.CertConfig{
        Type:     nfe.A1,
        Path:     "/path/to/cert.pfx",
        Password: "senha123",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Cria NFe
    make := client.CreateNFe()
    make.TagInfNFe(chave, "4.00")
    make.TagIde(ide)
    make.TagEmit(emitente)
    make.TagDest(destinatario)
    make.TagDet(item)
    make.TagTotal(total)

    nfeXML, err := make.GetXML()
    if err != nil {
        log.Fatal(err)
    }

    // Autoriza
    response, err := client.Authorize(nfeXML)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("NFe autorizada: %s", response.ProtocolNumber)
}

---

## 6. Especificações Técnicas

### 6.1 Arquitetura

#### 6.1.1 Estrutura de Pastas
```
github.com/adrianodrix/sped-nfe-go/
├── cmd/                          # CLI tools (opcional)
│   └── nfecli/                  # Ferramenta linha de comando
├── pkg/
│   ├── nfe/                     # Core NFe
│   │   ├── make.go             # Geração de XML
│   │   ├── sign.go             # Assinatura digital
│   │   ├── tools.go            # Ferramentas auxiliares
│   │   ├── webservices.go      # Comunicação SEFAZ
│   │   └── types.go            # Estruturas de dados
│   ├── certificate/             # Gerenciamento certificados
│   │   ├── a1.go               # Certificados A1
│   │   ├── a3.go               # Certificados A3
│   │   └── validator.go        # Validação certificados
│   ├── soap/                   # Cliente SOAP
│   │   ├── client.go           # HTTP client
│   │   ├── envelope.go         # SOAP envelope
│   │   └── security.go         # WS-Security
│   └── utils/                  # Utilitários
│       ├── xml.go              # Helpers XML
│       ├── validation.go       # Validações
│       └── cnpj.go            # Utilitários brasileiros
├── internal/                   # Código interno
│   ├── schemas/               # XSD schemas
│   └── testdata/             # Dados para testes
├── examples/                  # Exemplos de uso
├── docs/                     # Documentação
└── scripts/                  # Scripts de build/deploy
```

#### 6.1.2 Dependências
```go
// Principais dependências externas
require (
    github.com/lafriks/go-xmldsig v0.2.0        // Assinatura XML
    github.com/ThalesIgnite/crypto11 v1.2.5    // PKCS#11 (A3)
    software.sslmate.com/src/go-pkcs12 v0.2.0  // PKCS#12 (A1)
    github.com/beevik/etree v1.1.0              // XML processing
    github.com/go-playground/validator/v10      // Validation
    golang.org/x/net v0.17.0                    // HTTP/SOAP
    golang.org/x/crypto v0.14.0                 // Cryptography
)
```

### 6.2 Performance

#### 6.2.1 Benchmarks Alvo
- **Geração NFe**: < 10ms por documento
- **Assinatura Digital**: < 50ms por documento
- **Transmissão SEFAZ**: < 2s por lote
- **Memória**: < 50MB para 1000 NFe simultâneas

#### 6.2.2 Otimizações
- Pool de objetos para XML parsing
- Cache de certificados e schemas XSD
- Goroutines para operações I/O
- Context para timeout e cancelamento

### 6.3 Segurança

#### 6.3.1 Certificados Digitais
- Suporte a certificados ICP-Brasil A1 e A3
- Validação de cadeia de certificação
- Proteção de chaves privadas em memória
- Integração segura com HSM/tokens

#### 6.3.2 Comunicação
- TLS 1.2+ obrigatório para SEFAZ
- Validação de certificados do servidor
- Retry com backoff exponencial
- Rate limiting configurável

---

## 7. Roadmap de Desenvolvimento

### 7.1 Milestone 1: Fundação (4 semanas)
**Sprint 1-2: Setup e Certificados**
- [ ] Setup do projeto e estrutura de pastas
- [ ] Implementação de Certificate (A1/A3)
- [ ] Testes unitários para certificados
- [ ] CI/CD básico (GitHub Actions)

**Sprint 3-4: Estruturas Base**
- [ ] Definição de todas as structs NFe
- [ ] Validação com struct tags
- [ ] Utilitários básicos (CNPJ, chave de acesso)
- [ ] Documentação das structs

### 7.2 Milestone 2: Core NFe (6 semanas)
**Sprint 5-7: Make (Geração)**
- [ ] Implementação completa do Make
- [ ] Builder pattern para NFe/NFCe
- [ ] Geração de XML conforme schemas
- [ ] Testes com XMLs reais

**Sprint 8-10: Sign e Tools**
- [ ] Assinatura digital XML
- [ ] Validação XSD
- [ ] Ferramentas auxiliares
- [ ] Integração completa Make + Sign

### 7.3 Milestone 3: Webservices (4 semanas)
**Sprint 11-12: Cliente SOAP**
- [ ] Cliente HTTP robusto
- [ ] Envelope SOAP com WS-Security
- [ ] Retry e error handling
- [ ] Timeout e context

**Sprint 13-14: Serviços SEFAZ**
- [ ] Autorização (NFeAutorizacao4)
- [ ] Consulta (NFeConsultaProtocolo4)
- [ ] Status do serviço
- [ ] Testes em homologação

### 7.4 Milestone 4: Eventos e Contingência (4 semanas)
**Sprint 15-16: Eventos Fiscais**
- [ ] Carta de Correção (CCe)
- [ ] Cancelamento
- [ ] Manifestação do Destinatário
- [ ] EPEC (contingência)

**Sprint 17-18: Finalização**
- [ ] Contingência completa (FS-IA, SVC)
- [ ] Otimizações de performance
- [ ] Documentação completa
- [ ] Exemplos práticos

### 7.5 Milestone 5: Release (2 semanas)
**Sprint 19-20: Release v1.0**
- [ ] Testes de carga e stress
- [ ] Benchmark contra versão PHP
- [ ] Release notes e migração
- [ ] Marketing para comunidade

---

## 8. Casos de Teste

### 8.1 Testes Unitários
- **Cobertura**: Mínimo 90% de code coverage
- **Mocking**: Interfaces mockadas para external dependencies
- **Golden Files**: XML de referência para validação
- **Property-Based**: Testes com dados gerados automaticamente

### 8.2 Testes de Integração
- **SEFAZ Homologação**: Testes contra ambiente real
- **Certificados**: Testes com diferentes tipos de certificado
- **Múltiplos UF**: Validação com todos os estados
- **Cenários de Erro**: Tratamento de falhas de rede/SEFAZ

### 8.3 Testes de Performance
- **Benchmark**: Comparação com sped-nfe PHP
- **Load Testing**: 1000+ NFe simultâneas
- **Memory Profiling**: Detecção de vazamentos
- **Stress Testing**: Limites de sistema

---

## 9. Documentação

### 9.1 Documentação Técnica
- **GoDoc**: Documentação inline completa
- **README**: Getting started e exemplos básicos
- **CHANGELOG**: Histórico de versões
- **CONTRIBUTING**: Guia para contribuidores

### 9.2 Guias e Tutoriais
- **Quickstart**: Instalação com `go get` e primeiro uso em 5 minutos
- **Migration Guide**: Migração do sped-nfe PHP para Go
- **Best Practices**: Padrões recomendados para uso do pacote
- **Troubleshooting**: Solução de problemas comuns de importação e uso

### 9.3 Exemplos Práticos
- **NFe Simples**: Produto sem ST/II
- **NFe Completa**: Todos os campos preenchidos
- **NFCe**: Nota Fiscal de Consumidor
- **Eventos**: CCe, cancelamento, manifestação
- **Contingência**: EPEC e FS-IA

---

## 10. Critérios de Aceite

### 10.1 Funcionalidade
- [ ] Geração de NFe/NFCe conforme layout 4.00
- [ ] Assinatura digital com certificados A1/A3
- [ ] Comunicação com todos os SEFAZ estaduais
- [ ] Eventos fiscais completos
- [ ] Modes de contingência

### 10.2 Qualidade
- [ ] Cobertura de testes > 90%
- [ ] Zero vulnerabilidades de segurança
- [ ] Performance 5x superior ao PHP
- [ ] Documentação completa

### 10.3 Usabilidade
- [ ] API intuitiva e consistente
- [ ] Mensagens de erro claras
- [ ] Exemplos funcionais
- [ ] Setup simplificado

---

## 11. Métricas e KPIs

### 11.1 Métricas Técnicas
- **Performance**: Tempo de geração/assinatura
- **Qualidade**: Code coverage, bugs reportados
- **Reliability**: Uptime, success rate SEFAZ
- **Security**: Vulnerabilidades, certificate validation

### 11.2 Métricas de Produto
- **Adoção**: GitHub stars, downloads via `go get`, forks, dependências
- **Engagement**: Issues, PRs, discussions, importações
- **Satisfação**: Survey de usuários, feedback, reviews
- **Market Share**: Comparação com alternativas Go existentes

### 11.3 Métricas de Negócio
- **Community Growth**: Contribuidores ativos, importações do pacote
- **Enterprise Adoption**: Empresas usando em produção via `go get`
- **Ecosystem Impact**: Projetos que dependem do pacote
- **Market Position**: Reconhecimento na comunidade Go brasileira

---

## 12. Riscos e Mitigações

### 12.1 Riscos Técnicos

#### Complexidade dos Schemas XSD
- **Risco**: Layouts NFe são complexos e mudam frequentemente
- **Mitigação**: Parsing automático de XSD, versionamento claro
- **Contingência**: Manter compatibilidade com versões anteriores

#### Compatibilidade SEFAZ
- **Risco**: Diferenças entre SEFAZ estaduais
- **Mitigação**: Testes abrangentes, configuração por UF
- **Contingência**: Fallback para comportamento padrão

#### Performance de Assinatura Digital
- **Risco**: Operações criptográficas podem ser lentas
- **Mitigação**: Cache de certificados, otimização de algoritmos
- **Contingência**: Assinatura assíncrona opcional

### 12.2 Riscos de Produto

#### Adoção da Comunidade
- **Risco**: Resistência à migração do PHP
- **Mitigação**: Documentação de migração, performance clara
- **Contingência**: Wrapper para facilitar transição

#### Manutenção a Longo Prazo
- **Risco**: Projeto pode ficar sem manutenção
- **Mitigação**: Documentação extensiva, múltiplos maintainers
- **Contingência**: Transfer para organização da comunidade

### 12.3 Riscos Regulatórios

#### Mudanças na Legislação
- **Risco**: SEFAZ pode alterar especificações
- **Mitigação**: Monitoramento ativo, arquitetura flexível
- **Contingência**: Releases emergenciais para compliance

---

## 13. Conclusão

O **sped-nfe-go** representa uma oportunidade única de modernizar o ecossistema fiscal brasileiro com uma solução nativa Go de alta performance. Com base na sólida fundação do **nfephp-org/sped-nfe** e aproveitando insights de projetos como **frones/nfe**, este projeto pode estabelecer novo padrão de excelência para bibliotecas fiscais.

A estratégia de conversão incremental, começando com a estrutura PHP existente e evoluindo para aproveitar as características únicas do Go, oferece um caminho de baixo risco para entregar valor significativo à comunidade de desenvolvedores brasileiros.

O sucesso deste projeto não apenas preencherá uma lacuna importante no ecossistema Go, mas também demonstrará a viabilidade do Go para aplicações empresariais críticas no mercado brasileiro.

---

## 14. Referências

### 14.1 Projetos de Referência
- **nfephp-org/sped-nfe**: https://github.com/nfephp-org/sped-nfe
- **frones/nfe**: https://github.com/frones/nfe
- **webmaniabr/NFe-Go**: https://github.com/webmaniabr/NFe-Go
- **lafriks/go-xmldsig**: https://github.com/lafriks/go-xmldsig

### 14.2 Documentação Oficial
- **Portal NFe**: http://www.nfe.fazenda.gov.br/
- **Schemas XSD**: https://www.nfe.fazenda.gov.br/portal/exibirArquivo.aspx?conteudo=+YAUb0WfgB4=
- **Manual de Orientação**: https://www.nfe.fazenda.gov.br/portal/exibirArquivo.aspx?conteudo=Xhpw4x0HEyM=

### 14.3 Especificações Técnicas
- **ICP-Brasil**: https://www.gov.br/iti/pt-br/centrais-de-conteudo/doc-icp-01-v-5-3-1-pdf
- **XML-DSig**: https://www.w3.org/TR/xmldsig-core/
- **SOAP 1.2**: https://www.w3.org/TR/soap12/

---

**Documento aprovado por:** Adriano Santos (@adrianodrix)  
**Data de aprovação:** 02 de Julho de 2025  
**Próxima revisão:** 02 de Agosto de 2025