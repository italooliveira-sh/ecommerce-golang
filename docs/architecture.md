# Architecture

## Visão geral

Arquitetura em camadas inspirada em Clean Architecture, adaptada para o idioma Go. O fluxo de dependência é sempre de fora para dentro:

```
HTTP → Handler → Service → Repository (interface) → Infra (implementação)
```

A camada de domínio (`internal/domain`) não depende de ninguém. Ela define as entidades e as interfaces que as camadas externas implementam.

---

## Estrutura de pastas

```
ecommerce/
│
├── cmd/
│   └── api/
│       └── main.go               # entry point: inicializa config, DB, Redis e injeta dependências
│
├── internal/
│   │
│   ├── domain/                   # entidades e interfaces — núcleo da aplicação
│   │   ├── user.go               # struct User, Address
│   │   ├── product.go            # struct Product, Variant, Inventory, Category
│   │   ├── order.go              # struct Order, OrderItem, OrderStatus, Review
│   │   └── errors.go             # erros de domínio tipados (ErrNotFound, ErrOutOfStock, etc.)
│   │
│   ├── repository/               # interfaces de acesso a dados (contratos)
│   │   ├── user_repository.go    # interface UserRepository
│   │   ├── product_repository.go # interface ProductRepository, InventoryRepository
│   │   └── order_repository.go   # interface OrderRepository
│   │
│   ├── service/                  # regras de negócio e orquestração
│   │   ├── auth_service.go       # login, registro, refresh token
│   │   ├── user_service.go       # perfil, endereços
│   │   ├── product_service.go    # catálogo, variantes, imagens
│   │   └── order_service.go      # criação de pedido, mudança de status, reviews
│   │
│   ├── handler/                  # HTTP handlers — decode, validate, encode
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── product_handler.go
│   │   ├── order_handler.go
│   │   ├── router.go             # monta todas as rotas com chi
│   │   └── middleware/
│   │       ├── auth.go           # valida JWT, injeta user no context
│   │       ├── admin.go          # verifica role = admin
│   │       ├── logger.go         # loga request/response
│   │       └── ratelimit.go      # rate limiting por IP
│   │
│   └── infra/
│       ├── postgres/             # implementações dos repositories
│       │   ├── user_repo.go
│       │   ├── product_repo.go
│       │   └── order_repo.go
│       └── redis/
│           └── cache.go          # cache de sessão e refresh tokens
│
├── pkg/                          # utilitários reutilizáveis sem lógica de negócio
│   ├── jwt/
│   │   └── jwt.go               # geração e validação de tokens
│   ├── password/
│   │   └── password.go          # hash e comparação bcrypt
│   ├── validator/
│   │   └── validator.go         # wrapper do go-playground/validator
│   ├── pagination/
│   │   └── pagination.go        # helper de offset/limit e metadados de página
│   └── response/
│       └── response.go          # envelope JSON padronizado { data, error }
│
├── migrations/                   # arquivos SQL de migração (goose)
│   ├── 001_create_users.sql
│   ├── 002_create_addresses.sql
│   ├── 003_create_categories.sql
│   ├── 004_create_products.sql
│   ├── 005_create_product_variants.sql
│   ├── 006_create_product_images.sql
│   ├── 007_create_inventory.sql
│   ├── 008_create_orders.sql
│   ├── 009_create_order_items.sql
│   ├── 010_create_order_status_history.sql
│   └── 011_create_reviews.sql
│
├── sqlc/
│   ├── sqlc.yaml                 # configuração do sqlc
│   ├── queries/                  # queries SQL — sqlc lê daqui e gera código Go
│   │   ├── users.sql
│   │   ├── addresses.sql
│   │   ├── categories.sql
│   │   ├── products.sql
│   │   ├── variants.sql
│   │   ├── inventory.sql
│   │   ├── orders.sql
│   │   └── reviews.sql
│   └── generated/                # gerado pelo sqlc — não editar manualmente
│       ├── db.go
│       ├── models.go
│       └── *.sql.go
│
├── config/
│   └── config.go                 # leitura de variáveis de ambiente
│
├── docs/                         # documentação do projeto (este diretório)
│   ├── domain.md
│   ├── api.md
│   ├── architecture.md
│   └── decisions.md
│
├── docker-compose.yml            # Postgres + Redis para desenvolvimento local
├── Makefile                      # comandos: make run, make migrate, make sqlc, make test
├── .env.example                  # variáveis de ambiente necessárias
└── go.mod
```

---

## Camadas em detalhe

### domain/
Só structs e interfaces. Zero dependências externas. Não importa nada de `infra`, `handler` ou libs de terceiros.

```go
// domain/order.go
type OrderStatus string

const (
    StatusPending   OrderStatus = "pending"
    StatusConfirmed OrderStatus = "confirmed"
    StatusShipped   OrderStatus = "shipped"
    StatusDelivered OrderStatus = "delivered"
    StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    AddressID   uuid.UUID
    Status      OrderStatus
    Subtotal    decimal.Decimal
    ShippingFee decimal.Decimal
    Total       decimal.Decimal
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### repository/
Interfaces puras. A implementação fica em `infra/postgres`. Isso permite trocar o banco sem tocar nos services.

```go
// repository/order_repository.go
type OrderRepository interface {
    Create(ctx context.Context, order *domain.Order, items []domain.OrderItem) error
    FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
    FindByUserID(ctx context.Context, userID uuid.UUID, filters OrderFilters) ([]domain.Order, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus, note string) error
}
```

### service/
Recebe as interfaces de repository via injeção de dependência no construtor. Não conhece HTTP, não conhece SQL.

```go
// service/order_service.go
type OrderService struct {
    orderRepo     repository.OrderRepository
    inventoryRepo repository.InventoryRepository
    addressRepo   repository.AddressRepository
}

func NewOrderService(
    orderRepo repository.OrderRepository,
    inventoryRepo repository.InventoryRepository,
    addressRepo repository.AddressRepository,
) *OrderService {
    return &OrderService{orderRepo, inventoryRepo, addressRepo}
}
```

### handler/
Só decodifica request, valida campos e chama o service. Não tem lógica de negócio.

```go
// handler/order_handler.go
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "invalid body")
        return
    }
    if err := h.validator.Struct(req); err != nil {
        response.ValidationError(w, err)
        return
    }
    userID := middleware.UserIDFromContext(r.Context())
    order, err := h.orderService.CreateOrder(r.Context(), userID, req.AddressID, req.Items)
    if err != nil {
        response.HandleServiceError(w, err)
        return
    }
    response.Created(w, order)
}
```

---

## Stack de dependências

| lib | versão | função |
|---|---|---|
| go-chi/chi | v5 | roteador HTTP |
| jackc/pgx | v5 | driver PostgreSQL |
| sqlc-dev/sqlc | v1 | geração de código a partir de SQL |
| pressly/goose | v3 | migrations |
| redis/go-redis | v9 | cliente Redis |
| golang-jwt/jwt | v5 | JWT |
| go-playground/validator | v10 | validação de structs |
| google/uuid | v1 | geração de UUIDs |
| shopspring/decimal | v1 | aritmética decimal segura para dinheiro |
| spf13/viper | v1 | configuração via env/arquivo |

---

## Fluxo de desenvolvimento por domínio (vertical slices)

O projeto é construído em fatias verticais: cada domínio vai do banco até o HTTP antes de começar o próximo. Isso simula o ambiente real de equipes de produto e evita a armadilha de ter 11 migrations prontas mas nenhuma feature funcionando.

### Ciclo por domínio

```
1. migration(s)
      ↓
2. queries SQL  →  sqlc/queries/<domínio>.sql
      ↓
3. make sqlc    →  gera sqlc/generated/
      ↓
4. repository   →  internal/repository/<domínio>_repository.go  (interface)
      ↓
5. service TDD  →  internal/service/<domínio>_service.go
   Red → Green → Refactor
      ↓
6. handler      →  internal/handler/<domínio>_handler.go
      ↓
7. rotas        →  internal/handler/router.go
      ↓
8. teste E2E do fluxo completo
```

### Domínios e estado

| domínio | migrations | queries | service | handler |
|---|---|---|---|---|
| identidade (users + addresses) | ✅ | — | — | — |
| catálogo (categories → inventory) | — | — | — | — |
| vendas (orders → reviews) | — | — | — | — |

---

## Makefile — comandos principais

```makefile
make run          # sobe a aplicação
make dev          # sobe com hot reload (air)
make migrate-up   # roda as migrations pendentes
make migrate-down # reverte a última migration
make sqlc         # regenera o código do sqlc
make test         # roda todos os testes
make lint         # roda o golangci-lint
make docker-up    # sobe Postgres + Redis via docker-compose
make docker-down  # para os containers
```

---

## Variáveis de ambiente (.env.example)

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=troque-por-um-segredo-forte
JWT_EXPIRY_HOURS=1
JWT_REFRESH_EXPIRY_DAYS=7

# Bcrypt
BCRYPT_COST=12
```
