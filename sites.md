# Sites

Defina aqui os sites que o scraper deve consultar.

Formato:

```text
| name | url | pattern | enabled |
| --- | --- | --- | --- |
| Loja Exemplo | https://example.com/promocoes | (?i)(promocao\|oferta\|desconto).{0,120} | true |
```

Regras:

- `name`: nome curto do site.
- `url`: pagina que sera coletada.
- `pattern`: regex opcional para localizar promocoes no texto da pagina.
- Use `\|` quando a regex precisar de alternancia dentro da tabela Markdown.
- `enabled`: use `true` para ativar e `false` para ignorar.
- `headers`: coluna opcional com pares `Nome=Valor` separados por `;`. Use `\;` para ponto e virgula literal dentro do valor.

## Lista

| name | url | pattern | enabled | headers |
| --- | --- | --- | --- | --- |
| Kabum | https://www.kabum.com.br | (?i)(promocao\|promo\|oferta\|desconto\|cupom\|black friday\|frete gratis).{0,160} | true | |
| Pichau | https://www.pichau.com.br | (?i)(promocao\|promo\|oferta\|desconto\|cupom\|black friday\|frete gratis).{0,160} | true | |
| Terabyte Shop | https://www.terabyteshop.com.br | (?i)(promocao\|promo\|oferta\|desconto\|cupom\|black friday\|frete gratis).{0,160} | true | Accept-Language=pt-BR,pt\;q=0.9 |

