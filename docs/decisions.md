# Decisions (ADR)

Architecture Decision Records — registro das decisões técnicas e o raciocínio por trás de cada uma.

Formato: **contexto** (qual problema havia), **decisão** (o que foi escolhido), **consequências** (o que isso implica).

---

## ADR-001 — UUID em vez de serial/integer como PK

**Contexto**
Precisávamos escolher o tipo do identificador primário de todas as entidades.

**Decisão**
Usar `uuid` (v4) gerado na camada de aplicação, não pelo banco.

**Consequências**
- IDs não são previsíveis — um atacante que vê o ID 42 não consegue adivinhar que o ID 41 existe
- Geração no lado da aplicação significa que podemos montar o objeto completo antes do INSERT, simplificando o código
- Facilita migração para arquitetura distribuída no futuro (sem colisão de IDs entre instâncias)
- Tradeoff: UUIDs ocupam mais espaço que inteiros e são ligeiramente piores para índices B-tree — aceitável para o porte deste projeto

---

## ADR-002 — sqlc em vez de GORM

**Contexto**
Precisávamos de uma estratégia de acesso ao banco de dados.

**Decisão**
Usar `sqlc` para gerar código Go tipado a partir de queries SQL escritas manualmente.

**Consequências**
- Escrevemos SQL real — `SELECT FOR UPDATE`, CTEs, window functions são possíveis sem gambiarras
- O compilador Go valida os tipos gerados pelo sqlc — erros de tipo aparecem em tempo de compilação, não em runtime
- Queries explícitas: sempre sabemos exatamente o que está sendo executado no banco
- A comunidade Go tem cultura anti-ORM — sqlc está alinhado com o idioma da linguagem
- Tradeoff: mais verboso para CRUDs simples. Compensa pela transparência e controle
- Não é útil se precisar trocar de banco (Postgres → MySQL) — mas este projeto usa Postgres exclusivamente

---

## ADR-003 — chi em vez de Gin ou Echo

**Contexto**
Precisávamos de um roteador/framework HTTP.

**Decisão**
Usar `go-chi/chi`, que é um roteador leve construído sobre o `net/http` padrão.

**Consequências**
- Compatível 100% com `net/http` — qualquer middleware padrão da comunidade funciona
- Menos mágica: handlers são `http.HandlerFunc` normais, sem tipos proprietários do framework
- Obriga a aprender o `net/http` de verdade, o que é valioso para um desenvolvedor iniciando em Go
- Não tem "batteries included" como Gin — serialização, validação e erros são responsabilidade nossa (o que é bom para aprendizado)
- Tradeoff: mais código boilerplate nos handlers comparado ao Gin. Vale pelo aprendizado

---

## ADR-004 — Inventory separado de product_variants

**Contexto**
Precisávamos decidir onde armazenar a quantidade em estoque de cada variante.

**Decisão**
Criar uma tabela `inventory` separada com relação 1:1 com `product_variants`.

**Consequências**
- Estoque muda com altíssima frequência (a cada venda, reserva, reposição). Separar permite `SELECT FOR UPDATE` só na linha de inventory sem bloquear a leitura da variante
- Queries de catálogo (listar produtos) nunca precisam tocar na tabela de estoque — melhor performance
- Mais claro semanticamente: variante descreve o produto, inventory descreve disponibilidade
- Tradeoff: um JOIN a mais quando precisamos exibir estoque junto com a variante — custo mínimo

---

## ADR-005 — Copiar preço em order_items (não referenciar)

**Contexto**
Precisávamos decidir como registrar o preço de cada item no pedido.

**Decisão**
O campo `unit_price` em `order_items` é uma cópia do preço no momento da compra, não uma FK para o preço atual da variante.

**Consequências**
- Pedidos históricos nunca mudam de valor — essencial para relatórios financeiros e notas fiscais
- Se um produto muda de preço, pedidos antigos continuam corretos
- Implementa o padrão "event sourcing leve": o pedido registra o que aconteceu, não o estado atual
- Tradeoff: desnormalização intencional — os dados ficam duplicados entre `product_variants.price` e `order_items.unit_price`. É o comportamento correto

---

## ADR-006 — order_status_history como append-only

**Contexto**
Precisávamos rastrear as mudanças de status de um pedido ao longo do tempo.

**Decisão**
Nunca atualizar registros em `order_status_history` — apenas inserir. Toda mudança de status cria uma nova linha.

**Consequências**
- Histórico completo e imutável: quando cada estado foi atingido, quanto tempo levou cada etapa
- Essencial para suporte ao cliente ("quando meu pedido foi confirmado?")
- Permite relatórios operacionais: tempo médio entre confirmed e shipped, por exemplo
- Tradeoff: a tabela cresce com o tempo — índice em `order_id` é obrigatório para performance

---

## ADR-007 — Soft delete em products e users

**Contexto**
Precisávamos decidir como lidar com a remoção de produtos e usuários.

**Decisão**
Usar soft delete via flag `is_active` (products) — nunca DELETE físico em entidades com histórico de pedidos.

**Consequências**
- Pedidos antigos mantêm referência válida para o produto e para o usuário
- Produtos descontinuados ficam invisíveis no catálogo mas presentes nos relatórios
- Toda query pública deve filtrar `is_active = true` — responsabilidade do repository
- Tradeoff: dados "deletados" continuam ocupando espaço. Para este porte, irrelevante

---

## ADR-008 — Injeção de dependência manual (sem framework de DI)

**Contexto**
Precisávamos de uma estratégia de composição das dependências (service precisa de repository, handler precisa de service, etc.).

**Decisão**
Injeção de dependência manual no `main.go`, sem framework como Wire ou Fx.

**Consequências**
- Explícito: o `main.go` mostra exatamente como a aplicação é montada
- Sem mágica de reflection ou geração de código extra
- Alinhado com o idioma Go: "explicit is better than implicit"
- Tradeoff: o `main.go` fica mais longo à medida que o projeto cresce. Se isso virar problema, migrar para Wire é uma decisão reversível
