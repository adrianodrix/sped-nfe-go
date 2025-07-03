# SOAP Envelope Fix - Correção para Compatibilidade SEFAZ

## Problema Identificado
Durante os testes de inutilização, alguns servidores SEFAZ rejeitaram requisições SOAP com WS-Security timestamp, retornando erro:
```
org.apache.axis2.databinding.ADBException: Unexpected subelement Timestamp
```

## Servidores Afetados
- **Amazonas (AM)**: `sefaz.am.gov.br` - Rejeita WS-Security timestamp
- Outros servidores podem ter comportamento similar

## Solução Implementada

### 1. Envelope Simples
Criada função `CreateSimpleNFeSOAPEnvelope()` que gera envelope SOAP sem WS-Security:

```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" 
               xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
               xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <soap:Body>
        <!-- Conteúdo do corpo -->
    </soap:Body>
</soap:Envelope>
```

### 2. Envelope Normal
Mantido `CreateNFeSOAPEnvelope()` com WS-Security para compatibilidade:

```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" 
               xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
               xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <soap:Header>
        <wsse:Security xmlns:wsse="..." xmlns:wsu="...">
            <wsu:Timestamp wsu:Id="timestamp-1">
                <wsu:Created>2025-07-03T21:59:57.300Z</wsu:Created>
                <wsu:Expires>2025-07-03T22:04:57.300Z</wsu:Expires>
            </wsu:Timestamp>
        </wsse:Security>
    </soap:Header>
    <soap:Body>
        <!-- Conteúdo do corpo -->
    </soap:Body>
</soap:Envelope>
```

### 3. Seleção Automática
Função `needsSimpleEnvelope()` determina qual envelope usar baseado na URL:

```go
func needsSimpleEnvelope(url string) bool {
    problematicServers := []string{
        "sefaz.am.gov.br",        // Amazonas
        "nfe.sefaz.am.gov.br",   // Amazonas alternativo
        "nfce.sefaz.am.gov.br",  // Amazonas NFCe
    }
    
    for _, server := range problematicServers {
        if strings.Contains(url, server) {
            return true
        }
    }
    
    return false
}
```

### 4. Integração Transparente
A função `CreateNFeSOAPRequest()` seleciona automaticamente o envelope apropriado:

```go
// Alguns servidores SEFAZ não aceitam WS-Security timestamp
if needsSimpleEnvelope(url) {
    envelope, err = CreateSimpleNFeSOAPEnvelope(bodyContent)
} else {
    envelope, err = CreateNFeSOAPEnvelope(bodyContent)
}
```

## Arquivos Modificados
- `soap/envelope.go`: Adicionadas funções para envelope simples e seleção automática

## Teste e Validação
Execute o teste completo para verificar a correção:

```bash
./cmd/teste-completo-inutilizacao/teste-completo-inutilizacao <senha_certificado>
```

## Resultado Esperado
- **AM**: Deve aceitar envelope sem WS-Security
- **Outros estados**: Continuam funcionando com WS-Security
- **Compatibilidade**: Mantida para todos os servidores SEFAZ

## Próximos Passos
1. Executar teste completo para validar correção
2. Adicionar outros servidores problemáticos conforme identificados
3. Implementar assinatura digital (XMLDSig) para estados que exigem

---
*Documentação atualizada em: 2025-07-03*