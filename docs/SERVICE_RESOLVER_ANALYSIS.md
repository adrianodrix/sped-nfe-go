# Análise do Problema: QueryStatus (Funcionando) vs Authorize (Falhando)

## Problema Identificado

O `QueryStatus` funciona normalmente, mas o `Authorize` retorna resposta vazia da SEFAZ. A análise revelou que **diferentes sistemas de resolução de URLs** estão sendo usados.

## Diferenças Encontradas

### 1. QueryStatus (Funcionando)
```go
// usa t.getStatusServiceInfo() que chama:
return t.resolver.GetStatusServiceURL(uf, isProduction, t.model)
```

**Sistema usado**: `webservices/resolver.go` → `GetWebserviceURL()` → `webservices/urls.go`

**Configuração**: Usa `NFe55Config` do arquivo `webservices/urls.go`

### 2. Authorize (Falhando)
```go
// usa diretamente:
serviceInfo, err := t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeAutorizacao, env, t.model)
```

**Sistema usado**: `common/webservices.go` → `initializePRServices()`

**Configuração**: Usa configuração hardcoded do arquivo `common/webservices.go`

## URLs Diferentes Sendo Usadas

### QueryStatus (webservices/urls.go)
- **Status Homolog**: `https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeStatusServico4?wsdl`
- **Authorize Homolog**: `https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeAutorizacao4?wsdl`

### Authorize (common/webservices.go)  
- **Status Homolog**: `https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeStatusServico4?wsdl`
- **Authorize Homolog**: `https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeAutorizacao4?wsdl`

## Observação Crítica

**AS URLs SÃO IDÊNTICAS!** O problema NÃO está nas URLs, mas sim na **inconsistência dos sistemas de resolução**.

## Possíveis Causas do Problema

### 1. Diferenças na Estrutura de Dados
- `webservices/urls.go` usa struct `Service`
- `common/webservices.go` usa struct `WebServiceInfo`
- Campos diferentes podem estar sendo perdidos na conversão

### 2. Action Headers Diferentes
- QueryStatus: Action vem de `service.Operation` 
- Authorize: Action vem de `serviceInfo.Action`
- Pode haver diferença no header SOAPAction

### 3. Configurações SSL/TLS Diferentes
- Os dois sistemas podem ter configurações de cliente HTTP diferentes
- Certificados ou timeouts podem estar configurados de forma inconsistente

## Código Relevante

### QueryStatus Chain:
```
nfe/client.go:QueryStatus() 
  → nfe/tools.go:SefazStatus()
    → nfe/tools.go:getStatusServiceInfo()
      → webservices/resolver.go:GetStatusServiceURL()
        → webservices/urls.go:GetWebserviceURL()
```

### Authorize Chain:
```
nfe/client.go:Authorize() 
  → nfe/tools.go:sefazEnviaLoteInternal()
    → common/webservices.go:GetServiceURL()
      → common/webservices.go:initializePRServices()
```

## Solução Recomendada

**Unificar o sistema de resolução**: Fazer ambos os serviços usarem o mesmo método de resolução de URLs para garantir consistência total.

### Opção 1: Authorize usar o resolver
```go
// Em sefazEnviaLoteInternal, trocar:
serviceInfo, err := t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeAutorizacao, env, t.model)

// Por:
serviceInfo, err := t.resolver.GetAuthorizationServiceURL(uf, isProduction, t.model)
```

### Opção 2: QueryStatus usar webservices  
```go
// Em getStatusServiceInfo, trocar:
return t.resolver.GetStatusServiceURL(uf, isProduction, t.model)

// Por:
return t.webservices.GetServiceURL(uf, common.NFeStatusServico, env, t.model)
```

## Prioridade de Investigação

1. **ALTA**: Verificar diferenças nos headers SOAPAction
2. **ALTA**: Comparar configurações de client HTTP/TLS  
3. **MÉDIA**: Validar estruturas de dados e campos
4. **BAIXA**: Diferenças nas URLs (confirmado que são idênticas)

A inconsistência entre sistemas de resolução é a causa mais provável da resposta vazia do serviço de autorização.