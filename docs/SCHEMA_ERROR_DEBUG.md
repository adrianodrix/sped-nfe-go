# Debug do Erro de Schema XML NFe

## Erro Atual
```
Status=225, Protocolo=, StatusText=Falha no Schema XML do lote de NFe. 
org.xml.sax.SAXParseException; lineNumber: 1; columnNumber: 3557; 
cvc-complex-type.2.4.b: The content of element 'NFe' is not complete. 
One of '{"http://www.portalfiscal.inf.br/nfe":infNFeSupl, "http://www.w3.org/2000/0...
```

## Análise do Problema

### Posição do Erro: 3557
- O erro está na posição 3557 do XML
- Isso sugere que o XML vai até próximo do final mas está incompleto

### Elemento NFe Incompleto
O schema XSD espera um dos seguintes elementos após `infNFe`:
1. `infNFeSupl` (Informações suplementares da NFe)
2. `Signature` (namespace XMLDSig `http://www.w3.org/2000/09/xmldsig#`)

### Possíveis Causas

#### 1. Assinatura Ausente ou Mal Formada
- A assinatura XMLDSig não está sendo inserida
- A assinatura está com namespace incorreto
- A assinatura está mal posicionada

#### 2. Elemento infNFeSupl Ausente
- Para NFCe pode ser obrigatório o `infNFeSupl`
- Contém informações como QR Code

## Soluções a Testar

### 1. Verificar XML Gerado
Adicionar debug para ver o XML final:
```go
fmt.Printf("   ✅ XML: %s\n", xmlAssinado)
```

### 2. Verificar Modelo da NFe
- Se for NFCe (modelo 65), pode precisar de `infNFeSupl`
- Verificar se está usando modelo correto (55 = NFe, 65 = NFCe)

### 3. Validar Estrutura da Assinatura
A assinatura deve estar exatamente assim:
```xml
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe>...</infNFe>
  <Signature xmlns="http://www.w3.org/2000/09/xmldsig#">
    <!-- conteúdo da assinatura -->
  </Signature>
</NFe>
```

### 4. Verificar Implementação do XMLSigner
- Garantir que `signNFeSpecifically()` está inserindo assinatura corretamente
- Verificar se não há problemas de namespace
- Confirmar posicionamento como irmã de `infNFe`

## Debug Recomendado

1. **Examinar XML final** para ver se assinatura está presente
2. **Verificar posição 3557** no XML para identificar o que está faltando
3. **Validar namespace** da assinatura XMLDSig
4. **Confirmar modelo da NFe** (55 vs 65)

## Próximos Passos

Analisar o XML completo gerado para identificar exatamente o que está causando o erro de schema na posição 3557.