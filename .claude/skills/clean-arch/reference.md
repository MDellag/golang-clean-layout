# Clean Architecture - Cheat Sheet Rápida

## 🗂️ ¿Dónde va mi código?

| Necesito crear... | Va en... | Ejemplo |
|------------------|----------|---------|
| Entidad de negocio | `internal/domain/entity/` | `user.go`, `order.go` |
| DTO Request | `internal/domain/dto/request/` | `create_user.go` |
| DTO Response | `internal/domain/dto/response/` | `user_response.go` |
| Interface de Repository | `internal/domain/interfaces/repositories.go` | `UserRepository` |
| Interface de Service | `internal/domain/interfaces/services.go` | `UserService` |
| Value Object | `internal/domain/valueobjects/` | `money.go`, `email.go` |
| Error de dominio | `internal/domain/errors/` | `errors.go` |
| Constantes de negocio | `internal/domain/constants/` | `roles.go`, `status.go` |
| Lógica de negocio | `internal/services/` | `users.go` |
| Repository implementation | `internal/repositories/postgres/` | `users.go` |
| Database model | `internal/repositories/postgres/models/` | `user.go` |
| Mapper (DB ↔ Entity) | `internal/repositories/postgres/mappers/` | `user.go` |
| HTTP Handler | `internal/delivery/handlers/` | `users.go` |
| gRPC Handler | `internal/delivery/grpc/` | `users.go` |
| Event Listener | `internal/delivery/listeners/` | `kafka.go` |
| Routes | `internal/delivery/router/` | `router.go` |
| Background Job | `internal/workers/jobs/` | `email_job.go` |
| AWS Integration | `internal/infrastructure/aws/` | `s3.go`, `sqs.go` |
| Cache | `internal/infrastructure/cache/` | `redis.go` |
| External API Client | `internal/infrastructure/http/clients/` | `payment_client.go` |
| Middleware | `internal/infrastructure/middlewares/` | `auth.go` |
| DI Container | `internal/dependencies/` | `container.go` |
| Config | `config/` | `config.go` |
| Integration Test | `tests/integration/` | `user_test.go` |

## 📋 Templates Rápidos

### Nueva Entity

```go
// internal/domain/entity/product.go
package entity

import "time"

type Product struct {
    ID          string
    Name        string
    Price       decimal.Decimal
    Stock       int
    CreatedAt   time.Time
    UpdateatedAt time.Time
}

// Business methods
func (p *Product) IsAvailable() bool {
    return p.Stock > 0
}

func (p *Product) Validate() error {
    if p.Name == "" {
        return errors.New("product name is required")
    }
    if p.Price.IsNegative() {
        return errors.New("price cannot be negative")
    }
    return nil
}
```

### Repository Interface

```go
// internal/domain/interfaces/repositories.go
package interfaces

import (
    "context"
    "github.com/yourproject/internal/domain/entity"
)

type ProductRepository interface {
    Create(ctx context.Context, product *entity.Product) error
    GetByID(ctx context.Context, id string) (*entity.Product, error)
    Update(ctx context.Context, product *entity.Product) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filters map[string]interface{}) ([]*entity.Product, error)
}
```

### Service Interface

```go
// internal/domain/interfaces/services.go
package interfaces

import (
    "context"
    "github.com/yourproject/internal/domain/dto"
    "github.com/yourproject/internal/domain/entity"
)

type ProductService interface {
    CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*entity.Product, error)
    GetProduct(ctx context.Context, id string) (*entity.Product, error)
    UpdateProduct(ctx context.Context, id string, req dto.UpdateProductRequest) (*entity.Product, error)
    DeleteProduct(ctx context.Context, id string) error
    ListProducts(ctx context.Context, filters dto.ProductFilters) ([]*entity.Product, error)
}
```

### Service Implementation

```go
// internal/services/products.go
package services

import (
    "context"
    "github.com/yourproject/internal/domain/dto"
    "github.com/yourproject/internal/domain/entity"
    "github.com/yourproject/internal/domain/interfaces"
)

type productService struct {
    productRepo interfaces.ProductRepository
    cache       cache.Cache
}

func NewProductService(repo interfaces.ProductRepository, cache cache.Cache) interfaces.ProductService {
    return &productService{
        productRepo: repo,
        cache:       cache,
    }
}

func (s *productService) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*entity.Product, error) {
    product := &entity.Product{
        ID:    uuid.New().String(),
        Name:  req.Name,
        Price: req.Price,
        Stock: req.Stock,
    }

    if err := product.Validate(); err != nil {
        return nil, fmt.Errorf("invalid product: %w", err)
    }

    if err := s.productRepo.Create(ctx, product); err != nil {
        return nil, fmt.Errorf("failed to create product: %w", err)
    }

    return product, nil
}
```

### Repository Implementation

```go
// internal/repositories/postgres/products.go
package postgres

import (
    "context"
    "github.com/yourproject/internal/domain/entity"
    "github.com/yourproject/internal/domain/interfaces"
    "github.com/yourproject/internal/repositories/postgres/mappers"
    "github.com/yourproject/internal/repositories/postgres/models"
    "gorm.io/gorm"
)

type productRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) interfaces.ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
    model := mappers.ProductToModel(product)
    return r.db.WithContext(ctx).Create(model).Error
}

func (r *productRepository) GetByID(ctx context.Context, id string) (*entity.Product, error) {
    var model models.Product
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        return nil, err
    }
    return mappers.ProductToEntity(&model), nil
}
```

### Database Model

```go
// internal/repositories/postgres/models/product.go
package models

import "time"

type Product struct {
    ID        string    `gorm:"primaryKey;type:uuid"`
    Name      string    `gorm:"not null"`
    Price     float64   `gorm:"not null"`
    Stock     int       `gorm:"default:0"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Product) TableName() string {
    return "products"
}
```

### Mappers

```go
// internal/repositories/postgres/mappers/product.go
package mappers

import (
    "github.com/shopspring/decimal"
    "github.com/yourproject/internal/domain/entity"
    "github.com/yourproject/internal/repositories/postgres/models"
)

func ProductToEntity(model *models.Product) *entity.Product {
    return &entity.Product{
        ID:        model.ID,
        Name:      model.Name,
        Price:     decimal.NewFromFloat(model.Price),
        Stock:     model.Stock,
        CreatedAt: model.CreatedAt,
        UpdatedAt: model.UpdatedAt,
    }
}

func ProductToModel(entity *entity.Product) *models.Product {
    price, _ := entity.Price.Float64()
    return &models.Product{
        ID:        entity.ID,
        Name:      entity.Name,
        Price:     price,
        Stock:     entity.Stock,
        CreatedAt: entity.CreatedAt,
        UpdatedAt: entity.UpdatedAt,
    }
}
```

### DTOs

```go
// internal/domain/dto/request/product.go
package request

import "github.com/shopspring/decimal"

type CreateProductRequest struct {
    Name  string          `json:"name" validate:"required,min=3,max=100"`
    Price decimal.Decimal `json:"price" validate:"required,gt=0"`
    Stock int             `json:"stock" validate:"gte=0"`
}

// internal/domain/dto/response/product.go
package response

import (
    "github.com/shopspring/decimal"
    "time"
)

type ProductResponse struct {
    ID        string          `json:"id"`
    Name      string          `json:"name"`
    Price     decimal.Decimal `json:"price"`
    Stock     int             `json:"stock"`
    CreatedAt time.Time       `json:"created_at"`
    UpdatedAt time.Time       `json:"updated_at"`
}
```

### HTTP Handler

```go
// internal/delivery/handlers/products.go
package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/yourproject/internal/domain/dto/request"
    "github.com/yourproject/internal/domain/dto/response"
    "github.com/yourproject/internal/domain/interfaces"
)

type ProductHandler struct {
    productService interfaces.ProductService
}

func NewProductHandler(service interfaces.ProductService) *ProductHandler {
    return &ProductHandler{productService: service}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
    var req request.CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    product, err := h.productService.CreateProduct(r.Context(), req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    resp := response.ProductResponse{
        ID:        product.ID,
        Name:      product.Name,
        Price:     product.Price,
        Stock:     product.Stock,
        CreatedAt: product.CreatedAt,
        UpdatedAt: product.UpdatedAt,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(resp)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    product, err := h.productService.GetProduct(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    resp := response.ProductResponse{
        ID:        product.ID,
        Name:      product.Name,
        Price:     product.Price,
        Stock:     product.Stock,
        CreatedAt: product.CreatedAt,
        UpdatedAt: product.UpdatedAt,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

## 🚦 Reglas Visuales

### Flujo de Dependencias

```
┌─────────────────┐
│  HTTP Request   │
└────────┬────────┘
         │
         v
┌─────────────────┐
│    Handler      │ ← Parseo, Validación, Respuesta
└────────┬────────┘
         │ usa
         v
┌─────────────────┐
│    Service      │ ← Lógica de Negocio
└────────┬────────┘
         │ usa
         v
┌─────────────────┐
│   Repository    │ ← Persistencia
└────────┬────────┘
         │ usa
         v
┌─────────────────┐
│   Database      │
└─────────────────┘
```

### ¿Qué puede depender de qué?

```
✅ Handler → Service
✅ Service → Repository
✅ Service → Entity
✅ Repository → Entity
✅ Todo → Domain Interfaces

❌ Entity → Service
❌ Entity → Repository
❌ Repository → Service
❌ Domain → Infrastructure
```

## 🔧 Verificación Rápida

Antes de commit, pregúntate:

- [ ] ¿Mi entity tiene tags de DB? → ❌ Muévelos a models
- [ ] ¿Mi handler tiene lógica de negocio? → ❌ Muévela a service
- [ ] ¿Mi service accede a DB directamente? → ❌ Usa repository
- [ ] ¿Estoy retornando DB models del repo? → ❌ Usa mapper
- [ ] ¿Tengo dependencias circulares? → ❌ Revisa arquitectura
- [ ] ¿Mis tests están junto al código? → ✅ Correcto

## 📚 Recursos

- [Clean Architecture - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://golang.org/doc/effective_go)
