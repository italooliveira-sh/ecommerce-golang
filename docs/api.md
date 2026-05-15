# API reference

Base URL: `/api/v1`

Todas as respostas seguem o envelope:
```json
{ "data": {}, "error": null }
{ "data": null, "error": { "code": "INVALID_INPUT", "message": "..." } }
```

Autenticação via Bearer token no header: `Authorization: Bearer <jwt>`

---

## Auth

### POST /auth/register
Cria um novo usuário com role `customer`.

**Request**
```json
{
  "name": "João Silva",
  "email": "joao@email.com",
  "password": "senhaSegura123"
}
```

**Response 201**
```json
{
  "data": {
    "id": "uuid",
    "name": "João Silva",
    "email": "joao@email.com",
    "role": "customer",
    "created_at": "2025-05-10T14:32:00Z"
  }
}
```

---

### POST /auth/login

**Request**
```json
{ "email": "joao@email.com", "password": "senhaSegura123" }
```

**Response 200**
```json
{
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "expires_in": 3600
  }
}
```

---

### POST /auth/refresh
Renova o access token usando o refresh token.

**Request**
```json
{ "refresh_token": "eyJ..." }
```

---

### POST /auth/logout
Invalida o refresh token. Requer auth.

---

## Users (requer auth)

### GET /me
Retorna o perfil do usuário autenticado.

### PUT /me
Atualiza nome ou email.

**Request**
```json
{ "name": "João S.", "email": "novo@email.com" }
```

### DELETE /me
Soft delete — desativa a conta.

---

## Addresses (requer auth)

### GET /me/addresses
Lista todos os endereços do usuário autenticado.

### POST /me/addresses
**Request**
```json
{
  "street": "Rua das Flores, 123",
  "city": "Fortaleza",
  "state": "CE",
  "zip_code": "60000-000",
  "country": "BR",
  "is_default": true
}
```

### PUT /me/addresses/:id
Atualiza um endereço. Não permitido se houver pedido vinculado em aberto.

### DELETE /me/addresses/:id
Remove endereço. Retorna 409 se houver pedido vinculado.

---

## Categories

### GET /categories
Lista todas as categorias. Query params: `?parent_id=uuid` para filtrar por pai.

**Response 200**
```json
{
  "data": [
    {
      "id": "uuid",
      "parent_id": null,
      "name": "Roupas",
      "slug": "roupas",
      "description": "..."
    }
  ]
}
```

### GET /categories/:slug
Retorna uma categoria pelo slug, incluindo subcategorias.

### POST /categories (admin)
```json
{ "parent_id": "uuid|null", "name": "Camisetas", "slug": "camisetas", "description": "..." }
```

### PUT /categories/:id (admin)
### DELETE /categories/:id (admin)
Retorna 409 se houver produtos vinculados.

---

## Products

### GET /products
Lista produtos ativos com paginação e filtros.

**Query params**
| param | tipo | exemplo |
|---|---|---|
| page | int | ?page=1 |
| limit | int | ?limit=20 (max 100) |
| category | string | ?category=camisetas (slug) |
| q | string | ?q=polo (busca por nome) |
| sort | string | ?sort=price_asc, price_desc, newest |

**Response 200**
```json
{
  "data": {
    "items": [...],
    "total": 120,
    "page": 1,
    "limit": 20,
    "pages": 6
  }
}
```

### GET /products/:slug
Retorna produto completo com variantes, imagens e média de avaliações.

**Response 200**
```json
{
  "data": {
    "id": "uuid",
    "name": "Camiseta Polo",
    "slug": "camiseta-polo",
    "description": "...",
    "base_price": 59.00,
    "category": { "id": "uuid", "name": "Camisetas", "slug": "camisetas" },
    "images": [{ "url": "...", "is_primary": true }],
    "variants": [
      {
        "id": "uuid",
        "sku": "POLO-AZU-M",
        "name": "Azul · M",
        "price": 59.00,
        "available": 8
      }
    ],
    "rating_avg": 4.3,
    "rating_count": 27
  }
}
```

### POST /products (admin)
### PUT /products/:id (admin)
### DELETE /products/:id (admin)
Soft delete — marca `is_active = false`.

---

## Product variants (admin)

### GET /products/:id/variants
### POST /products/:id/variants
```json
{ "sku": "POLO-AZU-M", "name": "Azul · M", "price": 59.00 }
```
Cria a variante e o registro de inventory correspondente (quantity=0).

### PUT /products/:id/variants/:variantId
### DELETE /products/:id/variants/:variantId
Retorna 409 se houver order_items vinculados.

---

## Inventory (admin)

### GET /inventory/:variantId
```json
{ "data": { "variant_id": "uuid", "quantity": 10, "reserved": 2, "available": 8 } }
```

### PATCH /inventory/:variantId
Ajuste manual de estoque (reposição, correção).
```json
{ "quantity": 50, "note": "Reposição lote maio/25" }
```

---

## Orders (requer auth)

### POST /orders
Cria um pedido. O service verifica disponibilidade e reserva estoque atomicamente.

**Request**
```json
{
  "address_id": "uuid",
  "items": [
    { "variant_id": "uuid", "quantity": 2 },
    { "variant_id": "uuid", "quantity": 1 }
  ]
}
```

**Response 201**
```json
{
  "data": {
    "id": "uuid",
    "status": "pending",
    "items": [...],
    "subtotal": 267.00,
    "shipping_fee": 18.00,
    "total": 285.00,
    "created_at": "..."
  }
}
```

**Erros possíveis**
- `404` variant não encontrada
- `422` estoque insuficiente para um ou mais itens
- `422` endereço não pertence ao usuário

### GET /orders
Lista pedidos do usuário autenticado. Query: `?status=pending&page=1&limit=10`

### GET /orders/:id
Retorna pedido com itens e histórico de status.

### PATCH /orders/:id/cancel
Cancela o pedido. Só permitido se status for `pending` ou `confirmed`.

---

## Orders — admin

### GET /admin/orders
Lista todos os pedidos. Query: `?status=&user_id=&page=&limit=`

### PATCH /admin/orders/:id/status
Avança o status do pedido.
```json
{ "status": "shipped", "note": "Rastreio: BR123456789" }
```
Valida transições permitidas. Registra em `order_status_history`.

---

## Reviews

### GET /products/:id/reviews
Lista avaliações aprovadas de um produto. Query: `?page=&limit=`

### POST /products/:id/reviews (requer auth)
Só permitido se o usuário tiver um pedido com aquele produto em status `delivered`.
```json
{ "rating": 5, "comment": "Produto excelente, entrega rápida." }
```
Cria com `is_approved = false`. Admin precisa aprovar.

---

## Códigos de erro padrão

| código | status HTTP | situação |
|---|---|---|
| UNAUTHORIZED | 401 | token ausente ou inválido |
| FORBIDDEN | 403 | sem permissão para o recurso |
| NOT_FOUND | 404 | recurso não existe |
| CONFLICT | 409 | violação de constraint (endereço com pedido, etc.) |
| UNPROCESSABLE | 422 | regra de negócio violada (estoque, transição inválida) |
| INTERNAL | 500 | erro inesperado |
