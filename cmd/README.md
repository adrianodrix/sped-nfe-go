# Comandos de Teste sped-nfe-go

Esta pasta contém scripts de teste para validar a comunicação com SEFAZ usando certificados ICP-Brasil reais.

## test-sefaz-status

**Descrição**: Testa a comunicação com SEFAZ em ambiente de PRODUÇÃO usando certificado ICP-Brasil real.

**Segurança**: ✅ TOTALMENTE SEGURO - apenas consulta status do serviço, não envia documentos.

### Uso

```bash
go run cmd/test-sefaz-status/main.go <senha_do_certificado>
```

### Exemplo

```bash
go run cmd/test-sefaz-status/main.go "minhasenha123"
```

### Pré-requisitos

1. **Certificado ICP-Brasil**: Coloque seu arquivo `.pfx` em `refs/certificates/cert-valido-jan-2026.pfx`
2. **Senha do certificado**: Forneça como argumento do comando
3. **Conectividade**: Acesso à internet para comunicação com SEFAZ

### O que o teste faz

1. ✅ Carrega certificado ICP-Brasil A1 (.pfx)
2. ✅ Configura cliente NFe para ambiente de PRODUÇÃO
3. ✅ Configura autenticação SSL/TLS com certificado
4. ✅ Envia requisição de status para SEFAZ Paraná
5. ✅ Exibe resultado da comunicação

### Resultado esperado

```
🎉 SUCESSO! Comunicação com SEFAZ PRODUÇÃO funcionou!
   📊 Status SEFAZ: 107 - Servico em Operacao
   🌐 Online: true
   📍 UF: 41 | Ambiente: 1 (1=produção)
```

### Códigos de Status SEFAZ

- **107**: Serviço em operação normal ✅
- **108**: Serviço paralisado momentaneamente ⚠️
- **109**: Serviço paralisado sem previsão ❌

### Problemas comuns

**Erro de certificado**: Instale certificados ICP-Brasil no sistema
```bash
sudo ./install-icpbrasil-certs.sh
```

**Timeout**: SEFAZ pode estar sobrecarregado, tente novamente

**403 Forbidden**: Certificado pode estar inválido ou expirado

## Notas importantes

- Este teste é 100% seguro para uso em produção
- Nenhum documento NFe é enviado ao SEFAZ
- O certificado é usado apenas para autenticação SSL
- SSL verification está desabilitada apenas para teste