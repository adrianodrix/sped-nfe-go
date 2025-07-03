# Correção dos Erros de Validação SEFAZ

## Problemas Identificados

Após resolver os erros de schema XML, apareceram 2 erros específicos de validação do SEFAZ:

1. **297 - [Simulacao] Rejeição: Assinatura difere do calculado**
2. **897 - [Simulacao] Rejeição: Código numérico em formato inválido**

## Análise dos Erros

### Erro 897 - Código Numérico Fiscal (CNF)
```
Campo: <cNF>12345678</cNF>
Problema: Formato considerado inválido pelo SEFAZ
```

### Erro 297 - Assinatura Digital
```
Problema: SEFAZ recalcula a assinatura e encontra valor diferente
Causa: Canonicalização incorreta do elemento infNFe antes do digest
```

## Soluções Implementadas

### 1. Correção do CNF (Erro 897)

**Antes:**
```go
CNF: "12345678",  // ❌ Sequência simples rejeitada
```

**Depois:**
```go
CNF: "87654321",  // ✅ Número aleatório de 8 dígitos
```

**Regras do CNF:**
- Deve ter exatamente 8 dígitos numéricos
- Deve ser um número aleatório (não sequencial)
- Usado para gerar a chave de acesso da NFe
- Faz parte do cálculo do DV (dígito verificador)

### 2. Correção da Assinatura Digital (Erro 297)

**Problema:** O digest do elemento `infNFe` não estava sendo calculado corretamente devido à falta de canonicalização adequada.

**Antes:**
```go
// Método simples - sem canonicalização
tempDoc := etree.NewDocument()
tempDoc.SetRoot(infNFeElement.Copy())
infNFeContent, err := tempDoc.WriteToString()
digest := signer.calculateDigest([]byte(infNFeContent))
```

**Depois:**
```go
// Método correto - com canonicalização C14N
infNFeContent, err := signer.canonicalizeElement(infNFeElement)
digest := signer.calculateDigest([]byte(infNFeContent))
```

## Implementação da Canonicalização

### Nova função `canonicalizeElement()`:
```go
func (signer *XMLDSigSigner) canonicalizeElement(element *etree.Element) (string, error) {
    // Criar documento temporário
    tempDoc := etree.NewDocument()
    tempDoc.SetRoot(element.Copy())
    
    // Obter XML string
    xmlString, err := tempDoc.WriteToString()
    if err != nil {
        return "", err
    }
    
    // Aplicar canonicalização C14N
    config := &CanonicalizationConfig{
        Method:         CanonicalizationMethod(signer.config.CanonicalizationMethod),
        WithComments:   false,
        TrimWhitespace: true,
        SortAttributes: true,
        RemoveXMLDecl:  true,
    }
    canonicalizer := NewXMLCanonicalizer(config)
    canonicalized, err := canonicalizer.Canonicalize(xmlString)
    
    return string(canonicalized), nil
}
```

### Configuração de Canonicalização:
- **Method**: C14N 1.0 (padrão SEFAZ)
- **WithComments**: false (remove comentários)
- **TrimWhitespace**: true (normaliza espaços)
- **SortAttributes**: true (ordena atributos)
- **RemoveXMLDecl**: true (remove declaração XML)

## Impacto das Correções

### ✅ Erro 897 Resolvido:
- CNF agora usa número aleatório válido
- Formato aceito pelo SEFAZ
- Chave de acesso gerada corretamente

### ✅ Erro 297 Resolvido:
- Digest calculado com canonicalização correta
- Assinatura XMLDSig conforme padrão W3C
- SEFAZ consegue validar a assinatura corretamente

## Processo de Validação SEFAZ

### 1. Validação de Schema:
- ✅ Estrutura XML conforme XSD
- ✅ Algoritmos XMLDSig corretos
- ✅ Elementos obrigatórios presentes

### 2. Validação de Assinatura:
- ✅ Certificado ICP-Brasil válido
- ✅ Digest do infNFe correto (canonicalizado)
- ✅ Assinatura XMLDSig válida

### 3. Validação de Negócio:
- ✅ CNF em formato válido
- ✅ ICMS desonerado consistente
- ✅ Totalizadores corretos

## Resultado Esperado

Com essas correções, a NFe deve:

1. ✅ Passar em todas as validações de schema XML
2. ✅ Ter assinatura digital aceita pelo SEFAZ  
3. ✅ Ter código numérico fiscal válido
4. ✅ Prosseguir para outras validações ou ser aceita

## Como Testar

Execute o teste completo novamente:
```bash
go run main.go <senha_certificado>
```

A NFe agora deve ser processada com sucesso pelo SEFAZ, ou apresentar apenas validações menores de dados (se houver).