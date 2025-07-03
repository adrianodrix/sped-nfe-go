# Correção do Erro 795 - ICMS Desonerado

## Problema Identificado

Erro **795 - Total do ICMS Desonerado VICMSDeson difere do somatório dos itens** indica inconsistência entre o valor total de ICMS desonerado e a soma dos valores dos itens individuais.

### Erro Original
```
795 - Rejeição: Total do ICMS Desonerado VICMSDeson difere do somatório dos itens
```

## Análise do Problema

### No XML Gerado:

**Totalizador (correto):**
```xml
<ICMSTot>
  <vICMSDeson>0.00</vICMSDeson>
  <!-- outros totais -->
</ICMSTot>
```

**Item 1 (correto):**
```xml
<ICMS00>
  <!-- Sem vICMSDeson, pois CST=00 (tributado) -->
</ICMS00>
```

**Item 2 (incorreto):**
```xml
<ICMS40>
  <orig>0</orig>
  <CST>40</CST>
  <vICMSDeson>52.00</vICMSDeson>  <!-- ❌ ERRO: Valor inconsistente -->
  <motDesICMS>9</motDesICMS>
</ICMS40>
```

## Validação SEFAZ

A validação do SEFAZ verifica:
```
Σ(vICMSDeson dos itens) = vICMSDeson do totalizador

Item 1: 0.00 (não informado)
Item 2: 52.00
Soma: 52.00

Totalizador: 0.00

52.00 ≠ 0.00 → ERRO 795
```

## Solução Implementada

### Opção 1: Zerar ICMS Desonerado (Implementada)
```go
// Correção no item 2:
ICMS40: &nfe.ICMS40{
    Orig:       "0",
    CST:        "40",
    VICMSDeson: "0.00",  // ✅ Corrigido para 0.00
    MotDesICMS: "9",
}
```

**Resultado:**
- Item 1: 0.00
- Item 2: 0.00  
- Soma: 0.00
- Totalizador: 0.00
- ✅ **0.00 = 0.00 → VALIDAÇÃO OK**

### Opção 2: Ajustar Totalizador (Alternativa)
Se realmente houvesse ICMS desonerado:
```xml
<ICMSTot>
  <vICMSDeson>52.00</vICMSDeson>  <!-- Ajustar para somar os itens -->
</ICMSTot>
```

## Conceitos Fiscais

### ICMS Desonerado (vICMSDeson)
- **Conceito**: Valor do ICMS que deixou de ser cobrado por benefício fiscal
- **Quando usar**: Em operações com isenção, não incidência ou redução de base de cálculo
- **CST aplicáveis**: 
  - `40` (Isenta)
  - `41` (Não tributada) 
  - `50` (Suspensão)

### Motivo da Desoneração (motDesICMS)
- **Código 9**: Outros motivos de desoneração
- **Obrigatório**: Quando há valor em `vICMSDeson`
- **Lista completa**: Conforme tabela SEFAZ

## Estrutura Corrigida

### XML Final Esperado:
```xml
<!-- Item 1 - Tributado -->
<det nItem="1">
  <imposto>
    <ICMS>
      <ICMS00>
        <orig>0</orig>
        <CST>00</CST>
        <vBC>5000.00</vBC>
        <pICMS>18.00</pICMS>
        <vICMS>900.00</vICMS>
        <!-- Sem vICMSDeson: tributado normalmente -->
      </ICMS00>
    </ICMS>
  </imposto>
</det>

<!-- Item 2 - Isento -->
<det nItem="2">
  <imposto>
    <ICMS>
      <ICMS40>
        <orig>0</orig>
        <CST>40</CST>
        <vICMSDeson>0.00</vICMSDeson>  <!-- Corrigido -->
        <motDesICMS>9</motDesICMS>
      </ICMS40>
    </ICMS>
  </imposto>
</det>

<!-- Totalizador -->
<total>
  <ICMSTot>
    <vICMSDeson>0.00</vICMSDeson>  <!-- Soma: 0.00 + 0.00 = 0.00 -->
  </ICMSTot>
</total>
```

## Validações Relacionadas

### Outras validações de consistência:
- **794**: vBC ICMS ST difere do somatório
- **796**: vICMS difere do somatório  
- **797**: vBCST difere do somatório
- **798**: vST difere do somatório

### Fórmula geral:
```
Campo_Total = Σ(Campo_Item_i) para i=1 até n
```

## Como Testar

Execute o teste novamente:
```bash
go run main.go <senha_certificado>
```

A NFe agora deve passar na validação 795 do SEFAZ, pois os valores de ICMS desonerado estarão consistentes entre itens e totalizador.