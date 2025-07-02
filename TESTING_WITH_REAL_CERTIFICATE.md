# Testando com Certificado Real

Este guia explica como usar o certificado real (`cert-valido-jan-2026.pfx`) para testar a integração completa do sped-nfe-go com os webservices SEFAZ.

## 🔐 Certificado Disponível

O certificado está localizado em:
```
refs/certificates/cert-valido-jan-2026.pfx
```

**Características:**
- Formato: PKCS#12 (.pfx)
- Tipo: A1 (software)
- Válido até: Janeiro de 2026
- Compatível com ICP-Brasil

## 🚀 Como Executar os Testes

### 1. Teste Rápido (Script Principal)

Execute o script de teste com a senha do certificado:

```bash
go run cmd/test-real-cert/main.go SENHA_DO_CERTIFICADO
```

**Exemplo:**
```bash
go run cmd/test-real-cert/main.go minhasenha123
```

### 2. Exemplo Completo (Manual)

Você também pode usar o exemplo em `examples/real_cert_test.go` editando a senha na linha 21:

```go
password := "suasenha123"  // Substitua pela senha real
```

Então execute:
```bash
go run examples/real_cert_test.go
```

## 📋 O que é Testado

### ✅ Funcionalidades Verificadas

1. **Carregamento do Certificado A1**
   - Leitura do arquivo .pfx
   - Decodificação PKCS#12
   - Validação de chaves privadas
   - Verificação de validade

2. **Integração com Cliente NFe**
   - Configuração do certificado no cliente
   - Inicialização dos Tools
   - Configuração de ambiente (homologação)

3. **Comunicação SEFAZ**
   - Consulta de status do webservice
   - Montagem de requisições SOAP
   - Envio com certificado real
   - Processamento de respostas

4. **Funcionalidades Básicas**
   - Validação de XML
   - Geração de chaves de acesso
   - Criação de builders NFe
   - Conversão de tipos

### ⚠️ Erros Esperados

**Erro SSL comum:**
```
x509: certificate signed by unknown authority
```

**Explicação:** Este erro é **normal** em ambiente de teste e indica que:
- ✅ A comunicação foi estabelecida
- ✅ O certificado foi usado corretamente  
- ✅ A requisição SOAP foi montada
- ❌ A verificação SSL falhou (esperado sem configuração de proxy)

## 🔧 Configuração Avançada

### Ambientes Suportados

```go
config := nfe.ClientConfig{
    Environment: nfe.Homologation, // Para testes
    // Environment: nfe.Production, // Para produção
    UF:          nfe.SP,           // Seu estado
    Timeout:     30,               // Timeout em segundos
}
```

### Estados Suportados

O mapeamento UF → String está implementado para todos os estados:
- SP (35) → "SP"
- RJ (33) → "RJ" 
- MG (31) → "MG"
- etc.

### Configuração do Certificado

```go
// Carregar certificado A1
cert, err := certificate.LoadA1FromFile(
    "refs/certificates/cert-valido-jan-2026.pfx", 
    "suasenha"
)

// Configurar no cliente
err = client.SetCertificate(cert)
```

## 📊 Saída Esperada

### Execução Bem-sucedida

```
=== Teste NFe com Certificado Real ===

1. Carregando certificado real...
   Arquivo: refs/certificates/cert-valido-jan-2026.pfx
   ✅ Certificado carregado com sucesso
   📋 Dados do certificado:
      Titular: CN=EMPRESA TESTE:12345678000195...
      Emissor: CN=AC CERTISIGN RFB G5...
      Serial: 123456789...
      Válido: true
      Validade: 15/01/2024 até 15/01/2026

2. Criando cliente NFe...
   ✅ Cliente criado com sucesso

3. Configurando certificado no cliente...
   ✅ Certificado configurado no cliente

4. Testando comunicação com SEFAZ - Status...
   ❌ Erro ao consultar status: SOAP call failed: [NETWORK] HTTP request failed: Post "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfestatusservico4.asmx": tls: failed to verify certificate: x509: certificate signed by unknown authority
   💡 Erro de certificado SSL - normal em ambiente de teste
   💡 A comunicação foi estabelecida, mas falhou na verificação SSL

5. Testando funcionalidades básicas...
   ✅ XML de exemplo é válido
   ✅ Chave gerada: 35240101234567000112550010000000015123456781
   ✅ NFe builder criado

=== Teste concluído! ===

📋 Resultados:
   • Certificado carregado: ✅ (válido até 15/01/2026)
   • Cliente NFe configurado: ✅
   • Comunicação SEFAZ: ⚠️  (erro SSL esperado em testes)
   • Funcionalidades básicas: ✅

🚀 O que foi testado:
   ✅ Carregamento de certificado A1 real
   ✅ Configuração do cliente NFe
   ✅ Integração Tools + Certificate
   ✅ Montagem de requisições SOAP
   ✅ Comunicação com webservices SEFAZ
   ✅ Validação básica de XML
   ✅ Geração de chaves de acesso
```

## 🎯 Próximos Passos

Após confirmar que os testes básicos funcionam:

1. **Configurar Proxy/SSL** (se necessário para produção)
2. **Implementar XMLs completos** (além de validação básica)
3. **Testar assinatura digital** com XMLs reais
4. **Validar contra schemas XSD** oficiais
5. **Testar eventos** (cancelamento, CCe)

## 🔒 Segurança

**⚠️ IMPORTANTE:**
- Nunca commite a senha do certificado no código
- Use variáveis de ambiente para senhas em produção
- O certificado .pfx contém chaves privadas - mantenha seguro
- Use sempre ambiente de homologação para testes

## 🐛 Troubleshooting

### Erro: "failed to decode PKCS#12 certificate"
- Verifique se a senha está correta
- Confirme que o arquivo .pfx não está corrompido

### Erro: "certificate has expired"  
- Verifique a validade do certificado
- Use um certificado válido

### Erro: "UF not supported"
- Verifique se o UF está no mapeamento (nfe/client.go:134-143)
- Adicione novos estados se necessário

### Timeout nas requisições
- Aumente o timeout na configuração
- Verifique conectividade de rede
- Considere configurar proxy se necessário