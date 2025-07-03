# Investigação dos Endpoints SOAP do SEFAZ Paraná

## Problema Original

O SEFAZ Paraná estava retornando HTTP 200 com Content-Length: 0 (resposta vazia), impedindo a autorização de NFe.

## Análise Realizada

### Hipótese Inicial (INCORRETA)
Inicialmente suspeitei que o problema era o sufixo `?wsdl` nas URLs, que normalmente é usado apenas para descoberta de WSDL.

### Descoberta da Fonte Oficial
Consultando o arquivo HTML oficial da Receita Federal (`url-oficiais.html`), confirma-se que:

**O SEFAZ Paraná É UMA EXCEÇÃO e oficialmente usa `?wsdl` nas URLs:**

```
Sefaz Paraná - (PR)
NFeAutorizacao: https://nfe.sefa.pr.gov.br/nfe/NFeAutorizacao4?wsdl
NFeRetAutorizacao: https://nfe.sefa.pr.gov.br/nfe/NFeRetAutorizacao4?wsdl  
NfeStatusServico: https://nfe.sefa.pr.gov.br/nfe/NFeStatusServico4?wsdl
```

## URLs Oficiais Confirmadas

### Ambiente de Produção (PR)
- ✅ `https://nfe.sefa.pr.gov.br/nfe/NFeAutorizacao4?wsdl`
- ✅ `https://nfe.sefa.pr.gov.br/nfe/NFeRetAutorizacao4?wsdl`
- ✅ `https://nfe.sefa.pr.gov.br/nfe/NFeStatusServico4?wsdl`
- ✅ `https://nfe.sefa.pr.gov.br/nfe/NFeConsultaProtocolo4?wsdl`
- ✅ `https://nfe.sefa.pr.gov.br/nfe/NFeInutilizacao4?wsdl`
- ✅ `https://nfe.sefa.pr.gov.br/nfe/NFeRecepcaoEvento4?wsdl`

### Ambiente de Homologação (PR)
- ✅ `https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeAutorizacao4?wsdl`
- ✅ `https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeRetAutorizacao4?wsdl`
- ✅ `https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeStatusServico4?wsdl`

## Outros Estados (Para Comparação)

A maioria dos outros estados NÃO usa `?wsdl`:

- **São Paulo**: `https://nfe.fazenda.sp.gov.br/ws/nfeautorizacao4.asmx`
- **Rio Grande do Sul**: `https://nfe.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx`
- **Minas Gerais**: `https://nfe.fazenda.mg.gov.br/nfe2/services/NFeAutorizacao4`

Estados que também usam `?wsdl`:
- **Goiás**: `https://nfe.sefaz.go.gov.br/nfe/services/NFeAutorizacao4?wsdl`
- **Mato Grosso**: `https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeAutorizacao4?wsdl`

## Status da Investigação

✅ **CONFIRMADO**: URLs oficiais do Paraná incluem `?wsdl`
❌ **PROBLEMA PERSISTE**: Resposta vazia não é devido às URLs
🔍 **PRÓXIMA INVESTIGAÇÃO**: Verificar implementação SOAP client ou headers

## Próximos Passos

O problema de resposta vazia não é devido às URLs. Possíveis causas:

1. **Headers SOAP incorretos** (SOAPAction, Content-Type)
2. **Estrutura do envelope SOAP** incompatível
3. **Certificado SSL/TLS** não aceito pelo Paraná
4. **User-Agent ou outros headers** bloqueados
5. **Timeout ou configuração de rede**

A assinatura XMLDSig (erros 298/297) foi resolvida anteriormente, mas ainda há problemas na comunicação SOAP com o Paraná.