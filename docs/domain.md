# Domain model

E-commerce com três domínios principais: **identidade**, **catálogo** e **vendas**.

---

## Domínio de identidade

### users
Entidade central. Todo recurso do sistema (pedido, endereço, avaliação) pertence a um usuário.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | gerado no lado da aplicação |
| name | varchar(255) | |
| email | varchar(255) UNIQUE | |
| password_hash | varchar(255) | bcrypt, nunca armazenar senha em texto |
| role | enum('customer','admin') | sem tabela de permissões separada nesse estágio |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### addresses
Separado de `users` porque um cliente pode ter vários endereços e o endereço de um pedido precisa ser preservado mesmo que o usuário atualize o cadastro depois.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK → users | ON DELETE CASCADE |
| street | varchar(255) | |
| city | varchar(100) | |
| state | varchar(100) | |
| zip_code | varchar(20) | |
| country | varchar(100) | |
| is_default | bool | apenas um endereço padrão por usuário — enforçado na camada de serviço |

---

## Domínio de catálogo

Hierarquia: `categories → products → product_variants → inventory`

Cada camada tem responsabilidade e taxa de mudança diferentes. Não misturar.

### categories
Árvore auto-referenciada via `parent_id`. Permite hierarquias como Roupas > Masculino > Camisetas.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| parent_id | uuid FK → categories | NULL = categoria raiz |
| name | varchar(255) | |
| slug | varchar(255) UNIQUE | usado na URL: /categorias/camisetas |
| description | text | |

### products
O que o cliente vê na vitrine. Representa o conceito comercial, não o item físico.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| category_id | uuid FK → categories | |
| name | varchar(255) | |
| slug | varchar(255) UNIQUE | usado na URL: /produtos/camiseta-polo |
| description | text | |
| base_price | numeric(12,2) | fallback quando não há variantes |
| is_active | bool | soft delete — nunca deletar produto com pedidos |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### product_variants
O item concreto e vendável. Tem SKU único, atributos (cor, tamanho) e preço definitivo que sobrescreve o `base_price` do produto.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| product_id | uuid FK → products | ON DELETE RESTRICT |
| sku | varchar(100) UNIQUE | código de estoque — único no sistema inteiro |
| name | varchar(255) | ex: "Azul · M" |
| price | numeric(12,2) | sobrescreve base_price se preenchido |
| is_active | bool | |

### product_images
| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| product_id | uuid FK → products | ON DELETE CASCADE |
| url | varchar(500) | URL pública da imagem |
| display_order | int | ordem de exibição |
| is_primary | bool | imagem principal do produto |

### inventory
Separado de `product_variants` porque muda com altíssima frequência (a cada venda, reserva ou reposição). Relação 1:1 com variante.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| variant_id | uuid FK UNIQUE → product_variants | ON DELETE CASCADE |
| quantity | int | estoque físico total |
| reserved | int | comprometido em pedidos com status pending/confirmed |
| updated_at | timestamptz | |

**Regra:** `disponível = quantity - reserved`. Nunca exibir `quantity` diretamente ao cliente.

**Operações atômicas:**
- Pedido criado (pending): `reserved += qty`
- Pedido confirmado (pending → confirmed): sem mudança no estoque ainda
- Pedido cancelado de pending: `reserved -= qty`
- Pedido cancelado de confirmed: `quantity -= qty` + `reserved -= qty` (ou `quantity` permanece e só `reserved` cai, dependendo da regra de negócio)
- Pedido entregue: `quantity -= qty` e `reserved -= qty`

---

## Domínio de vendas

### orders
Snapshot de uma compra. Uma vez criado, os valores (total, endereço, itens) são imutáveis.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK → users | ON DELETE RESTRICT |
| address_id | uuid FK → addresses | ON DELETE RESTRICT — não permitir deletar endereço com pedido |
| status | enum | ver máquina de estados abaixo |
| subtotal | numeric(12,2) | soma dos order_items.subtotal |
| shipping_fee | numeric(12,2) | calculado no momento do pedido |
| total | numeric(12,2) | subtotal + shipping_fee |
| created_at | timestamptz | |
| updated_at | timestamptz | |

**Máquina de estados do status:**
```
pending → confirmed → shipped → delivered
    └──────────────→ cancelled (somente de pending ou confirmed)
```

- `pending`: pedido criado, estoque reservado, aguardando pagamento
- `confirmed`: pagamento aprovado, estoque debitado, em separação
- `shipped`: enviado à transportadora, cancelamento não permitido
- `delivered`: entregue, cliente pode avaliar o produto
- `cancelled`: liberação de reserva (de pending) ou devolução ao estoque (de confirmed)

### order_items
Cada linha do pedido. O `unit_price` é uma cópia do preço no momento da compra — nunca uma FK para o preço atual.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| order_id | uuid FK → orders | ON DELETE CASCADE |
| product_id | uuid FK → products | ON DELETE RESTRICT — referência histórica |
| variant_id | uuid FK → product_variants | ON DELETE RESTRICT |
| quantity | int | |
| unit_price | numeric(12,2) | cópia imutável do preço na compra |
| subtotal | numeric(12,2) | quantity × unit_price |

### order_status_history
Trilha de auditoria imutável. Toda mudança de status gera uma nova linha — nunca atualizar, só inserir.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| order_id | uuid FK → orders | ON DELETE CASCADE |
| status | enum | estado no momento do registro |
| note | text | observação do admin (opcional) |
| created_at | timestamptz | |

### reviews
Avaliação pós-entrega. A regra "só quem comprou pode avaliar" é verificada no `order_service`.

| campo | tipo | observação |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK → users | ON DELETE CASCADE |
| product_id | uuid FK → products | ON DELETE CASCADE |
| rating | int | 1 a 5, validado na camada de handler |
| comment | text | |
| is_approved | bool | default false — requer aprovação do admin |
| created_at | timestamptz | |

---

## Relacionamentos resumidos

| de | para | cardinalidade | observação |
|---|---|---|---|
| users | addresses | 1:N | um usuário tem vários endereços |
| users | orders | 1:N | um usuário faz vários pedidos |
| users | reviews | 1:N | um usuário escreve várias avaliações |
| categories | categories | auto-ref | hierarquia pai → filho |
| categories | products | 1:N | uma categoria agrupa produtos |
| products | product_variants | 1:N | um produto tem várias variantes |
| products | product_images | 1:N | um produto tem várias imagens |
| product_variants | inventory | 1:1 | cada variante tem exatamente um registro de estoque |
| orders | order_items | 1:N | um pedido contém vários itens |
| orders | order_status_history | 1:N | trilha de estados do pedido |
| addresses | orders | 1:N | endereço referenciado no pedido |
