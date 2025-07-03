# Correção da Estrutura XMLDSig - Reference/Transforms

## Problema Identificado

O erro do SEFAZ estava relacionado à **estrutura incorreta do elemento Reference** na assinatura XMLDSig. O schema XMLDSig exige que o elemento `Transforms` apareça **antes** do elemento `DigestMethod` dentro de `Reference`.

### Erro Original
```xml
<ds:Reference URI="#NFe...">
  <ds:DigestMethod Algorithm="..."/>  <!-- ERRADO: DigestMethod sem Transforms -->
  <ds:DigestValue>...</ds:DigestValue>
</ds:Reference>
```

### Estrutura Corrigida
```xml
<ds:Reference URI="#NFe...">
  <ds:Transforms>                     <!-- CORRETO: Transforms primeiro -->
    <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
    <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
  </ds:Transforms>
  <ds:DigestMethod Algorithm="..."/>   <!-- Depois DigestMethod -->
  <ds:DigestValue>...</ds:DigestValue>
</ds:Reference>
```

## Mudanças Implementadas

### 1. XMLSigner (`signer.go`)

**Função corrigida:** `createSignatureTemplate()`
- Adicionado elemento `<ds:Transforms>` antes de `<ds:DigestMethod>`
- Incluídos os transforms obrigatórios para NFe:
  - `enveloped-signature`: Remove a assinatura do cálculo do digest
  - `xml-exc-c14n`: Canonicalização exclusiva

```go
// Antes:
<ds:Reference URI="%s">
  <ds:DigestMethod Algorithm="%s"/>
  <ds:DigestValue>%s</ds:DigestValue>
</ds:Reference>

// Depois:
<ds:Reference URI="%s">
  <ds:Transforms>
    <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
    <ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
  </ds:Transforms>
  <ds:DigestMethod Algorithm="%s"/>
  <ds:DigestValue>%s</ds:DigestValue>
</ds:Reference>
```

### 2. XMLDSigSigner (`xmldsig_signer.go`)

**Função verificada:** `createSignatureElement()`
- ✅ Já estava correto com elemento `Transforms`
- ✅ Já usava configuração `TransformMethods` do `DefaultXMLDSigConfig()`

### 3. Mudança na Assinatura do Teste

**Arquivo:** `/cmd/teste-completo-nfe/main.go`
- Substituído `CreateXMLSigner()` por `NewXMLDSigSigner()`
- Motivo: XMLDSigSigner tem melhor conformidade com padrões XMLDSig

```go
// Antes:
signer := certificate.CreateXMLSigner(cert)
xmlAssinado, err := signer.SignNFeXML(xml)

// Depois:
signer := certificate.NewXMLDSigSigner(cert, certificate.DefaultXMLDSigConfig())
result, err := signer.SignNFeXML(xml)
xmlAssinado := result.SignedXML
```

## Conformidade com Schema XMLDSig

A correção garante que a assinatura XMLDSig esteja em conformidade com:

- ✅ **W3C XML Signature Syntax**: https://www.w3.org/TR/xmldsig-core/
- ✅ **Schema XSD XMLDSig**: Ordem correta dos elementos
- ✅ **Padrões SEFAZ**: Transforms obrigatórios para NFe
- ✅ **Canonicalização**: Exclusiva C14N para compatibilidade

## Transforms Incluídos

### 1. Enveloped Signature Transform
- **Algorithm**: `http://www.w3.org/2000/09/xmldsig#enveloped-signature`
- **Propósito**: Remove a própria assinatura do cálculo do digest
- **Obrigatório**: Sim, para assinaturas envelopadas

### 2. Exclusive Canonicalization Transform
- **Algorithm**: `http://www.w3.org/2001/10/xml-exc-c14n#`
- **Propósito**: Normaliza o XML para cálculo consistente
- **Obrigatório**: Sim, padrão SEFAZ

## Resultado Esperado

Com esta correção, o erro do SEFAZ:
```
"The element 'Reference' in namespace 'http://www.w3.org/2000/09/xmldsig#' 
has invalid child element 'DigestMethod'... List of possible elements expected: 'Transforms'"
```

Deve ser resolvido, pois agora a estrutura XMLDSig segue exatamente o padrão W3C e SEFAZ.

## Como Testar

Execute o teste completo novamente:
```bash
go run main.go <senha_certificado>
```

A assinatura XMLDSig agora deve ser aceita pelo SEFAZ sem erros de schema.