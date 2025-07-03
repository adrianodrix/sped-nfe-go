# Estrutura de Pagamento Corrigida - NFe

## Problema Identificado

O erro do SEFAZ estava relacionado à estrutura XML incorreta para pagamentos. A estrutura anterior não seguia o schema XSD oficial da NFe versão 4.00.

### Erro Original
```
Status=225, Protocolo=, StatusText=Falha no Schema XML do lote de NFe. 
org.xml.sax.SAXParseException; lineNumber: 1; columnNumber: 3227; 
cvc-complex-type.2.4.a: Invalid content was found starting with element 'tPag'. 
One of '{"http://www.portalfiscal.inf.br/nfe":detPag}' is expected.
```

## Solução Implementada

### Estrutura Anterior (Incorreta)
```xml
<pag>
    <tPag>01</tPag>
    <vPag>100.00</vPag>
</pag>
```

### Estrutura Nova (Correta)
```xml
<pag>
    <detPag>
        <tPag>01</tPag>
        <vPag>100.00</vPag>
    </detPag>
    <vTroco>5.00</vTroco> <!-- Opcional: valor do troco -->
</pag>
```

## Mudanças no Código

### 1. Estrutura de Dados Go

**Antes:**
```go
type Pagamento struct {
    XMLName xml.Name `xml:"pag"`
    TPag    string   `xml:"tPag"`
    VPag    string   `xml:"vPag"`
    // ... outros campos ...
}
```

**Depois:**
```go
type Pagamento struct {
    XMLName xml.Name `xml:"pag"`
    DetPag  []DetPag `xml:"detPag"`
    VTroco  string   `xml:"vTroco,omitempty"`
}

type DetPag struct {
    XMLName xml.Name `xml:"detPag"`
    TPag    string   `xml:"tPag"`
    VPag    string   `xml:"vPag"`
    // ... outros campos ...
}
```

### 2. Exemplo de Uso

**Antes:**
```go
pagamento := &nfe.Pagamento{
    TPag: "01",
    VPag: "100.00",
}
```

**Depois:**
```go
pagamento := &nfe.Pagamento{
    DetPag: []nfe.DetPag{
        {
            TPag: "01",
            VPag: "100.00",
        },
    },
}
```

### 3. Usando os Builders

```go
// Criar pagamento individual
detPag := nfe.NewDetPagBuilder().
    SetType(nfe.PaymentTypeMoney).
    SetValue("100.00").
    Build()

// Criar grupo de pagamentos
pagamento := nfe.NewPaymentBuilder().
    AddDetPag(detPag).
    Build()

// Ou usando funções de conveniência
pagamento := nfe.CreateCashPayment("100.00")
```

### 4. Múltiplos Pagamentos

Agora é possível ter até 100 formas de pagamento diferentes:

```go
pagamento := &nfe.Pagamento{
    DetPag: []nfe.DetPag{
        {
            TPag: "01",      // Dinheiro
            VPag: "50.00",
        },
        {
            TPag: "03",      // Cartão de Crédito
            VPag: "50.00",
            Card: &nfe.Card{
                TpIntegra: "2",
                TBand:     "01",
                CAut:      "123456",
            },
        },
    },
    VTroco: "5.00", // Troco
}
```

## Conformidade com Schema XSD

A nova estrutura segue exatamente o schema XSD oficial da NFe versão 4.00:

- ✅ Elemento `<pag>` como container
- ✅ Múltiplos `<detPag>` (1 a 100)
- ✅ Ordem correta dos elementos
- ✅ Campos opcionais como `vTroco`
- ✅ Estrutura `<card>` para pagamentos com cartão
- ✅ Validação de tipos de pagamento

## Resultado

Esta correção resolve o erro do SEFAZ e permite que as NFe sejam processadas corretamente pelos webservices oficiais.