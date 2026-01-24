# Ejemplos Prácticos del Skill /clean-arch

## Escenarios de Uso Común

### Ejemplo 1: Agregar Nueva Feature - Sistema de Notificaciones

**Pregunta:**
```
/clean-arch necesito agregar un sistema de notificaciones por email y SMS. ¿Cómo lo estructuro?
```

**Respuesta Esperada:**

```
📁 Estructura propuesta para Sistema de Notificaciones:

1. Domain Layer:
   ├── internal/domain/entity/notification.go
   │   └── type Notification struct { ID, Type, Recipient, Content, Status }
   │
   ├── internal/domain/valueobjects/notification_type.go
   │   └── const (Email, SMS, Push)
   │
   ├── internal/domain/dto/request/send_notification.go
   │   └── type SendNotificationRequest struct
   │
   ├── internal/domain/dto/response/notification.go
   │   └── type NotificationResponse struct
   │
   └── internal/domain/interfaces/
       ├── repositories.go → Add NotificationRepository
       └── services.go → Add NotificationService

2. Service Layer:
   └── internal/services/notifications.go
       └── Implementa NotificationService

3. Infrastructure Layer:
   ├── internal/infrastructure/email/
   │   ├── smtp.go
   │   └── sendgrid.go (si usas SendGrid)
   │
   └── internal/infrastructure/sms/
       ├── twilio.go
       └── aws_sns.go

4. Repository Layer:
   ├── internal/repositories/postgres/notifications.go
   ├── internal/repositories/postgres/models/notification.go
   └── internal/repositories/postgres/mappers/notification.go

5. Delivery Layer:
   ├── internal/delivery/handlers/notifications.go
   └── internal/workers/jobs/notification_job.go (para async)

6. Tests:
   ├── tests/integration/notification_test.go
   └── internal/services/notifications_test.go
```

---

### Ejemplo 2: Integración con API Externa - Stripe

**Pregunta:**
```
/clean-arch voy a integrar Stripe para pagos. ¿Dónde va cada cosa?
```

**Respuesta Esperada:**

```
📁 Estructura para Integración de Stripe:

1. Infrastructure Layer (implementación):
   └── internal/infrastructure/payments/stripe/
       ├── client.go           # Cliente de Stripe
       ├── payment.go          # Operaciones de pago
       ├── webhook.go          # Manejo de webhooks
       └── mapper.go           # Stripe models → Domain entities

2. Domain Layer (contratos):
   ├── internal/domain/entity/payment.go
   │   └── type Payment struct { Amount, Currency, Status }
   │
   ├── internal/domain/interfaces/infrastructure.go
   │   └── type PaymentGateway interface {
   │         ProcessPayment(ctx, amount, currency) (*Payment, error)
   │         RefundPayment(ctx, paymentID) error
   │       }
   │
   └── internal/domain/dto/request/payment.go

3. Service Layer:
   └── internal/services/payments.go
       └── Usa PaymentGateway interface

4. Delivery Layer:
   ├── internal/delivery/handlers/payments.go
   └── internal/delivery/handlers/stripe_webhook.go

5. Configuration:
   └── config/config.go
       └── type Config struct {
             StripeAPIKey    string `envconfig:"STRIPE_API_KEY"`
             StripeWebhookSecret string `envconfig:"STRIPE_WEBHOOK_SECRET"`
           }
```

**Reglas importantes:**
- El service NO debe conocer Stripe directamente
- Usa la interface `PaymentGateway`
- Si cambias a otro gateway, solo cambias la implementación
- Los webhooks son handlers especiales en delivery layer

---

### Ejemplo 3: Refactorizar Código Existente

**Situación Actual (INCORRECTA):**

```go
// internal/delivery/handlers/user.go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    json.NewDecoder(r.Body).Decode(&user)

    // ❌ Lógica de negocio en handler
    if len(user.Password) < 8 {
        http.Error(w, "password too short", 400)
        return
    }

    // ❌ Acceso directo a DB desde handler
    db.Create(&user)

    json.NewEncoder(w).Encode(user)
}
```

**Pregunta:**
```
/clean-arch necesito refactorizar este handler que tiene lógica de negocio y acceso a DB
```

**Refactorización CORRECTA:**

```go
// 1. Domain Entity
// internal/domain/entity/user.go
type User struct {
    ID       string
    Email    string
    Password string
}

func (u *User) ValidatePassword() error {
    if len(u.Password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    return nil
}

// 2. DTO Request
// internal/domain/dto/request/user.go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

// 3. Service
// internal/services/users.go
func (s *userService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*entity.User, error) {
    user := &entity.User{
        ID:       uuid.New().String(),
        Email:    req.Email,
        Password: req.Password,
    }

    if err := user.ValidatePassword(); err != nil {
        return nil, err
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}

// 4. Handler (delgado)
// internal/delivery/handlers/users.go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    user, err := h.userService.CreateUser(r.Context(), req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user)
}
```

---

### Ejemplo 4: Worker/Background Job

**Pregunta:**
```
/clean-arch necesito procesar reportes en background. ¿Cómo implemento un job?
```

**Estructura:**

```go
// 1. Job Definition
// internal/workers/jobs/report_job.go
package jobs

type ReportJob struct {
    reportService domain.ReportService
}

func NewReportJob(service domain.ReportService) *ReportJob {
    return &ReportJob{reportService: service}
}

func (j *ReportJob) Execute(ctx context.Context, payload interface{}) error {
    reportID, ok := payload.(string)
    if !ok {
        return errors.New("invalid payload")
    }

    return j.reportService.GenerateReport(ctx, reportID)
}

// 2. Register en Dispatcher
// internal/workers/dispatcher.go
dispatcher.RegisterJob("generate_report", reportJob)

// 3. Encolar desde Handler
// internal/delivery/handlers/reports.go
func (h *ReportHandler) RequestReport(w http.ResponseWriter, r *http.Request) {
    reportID := generateID()

    // Encolar job
    h.dispatcher.Enqueue("generate_report", reportID)

    // Responder inmediatamente
    json.NewEncoder(w).Encode(map[string]string{
        "report_id": reportID,
        "status": "processing",
    })
}
```

---

### Ejemplo 5: Tests

**Pregunta:**
```
/clean-arch ¿dónde van los tests para cada capa?
```

**Estructura de Tests:**

```
1. Unit Tests (junto al código):
   internal/services/users_test.go
   internal/repositories/postgres/users_test.go
   internal/domain/entity/user_test.go

2. Integration Tests:
   tests/integration/user_api_test.go
   tests/integration/user_repository_test.go

3. E2E Tests:
   tests/e2e/user_flow_test.go

4. Mocks:
   tests/testdata/mocks/user_repository_mock.go
   tests/testdata/mocks/user_service_mock.go
```

**Ejemplo de Test de Service:**

```go
// internal/services/users_test.go
package services_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := new(mocks.UserRepository)
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    service := services.NewUserService(mockRepo, nil)

    req := dto.CreateUserRequest{
        Email:    "test@example.com",
        Password: "password123",
    }

    // Act
    user, err := service.CreateUser(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, req.Email, user.Email)
    mockRepo.AssertExpectations(t)
}
```

---

## Comandos Útiles para Verificación

```bash
# Ver dependencias entre paquetes
go mod graph | grep internal

# Detectar dependencias circulares
go list -f '{{.ImportPath}} {{.Imports}}' ./... | grep internal

# Verificar que domain no depende de infrastructure
go list -f '{{.ImportPath}} {{.Imports}}' ./internal/domain/... | grep infrastructure
# ↑ Esto NO debe retornar nada

# Ver estructura del proyecto
tree -L 4 -I 'vendor|.git' internal/

# Ejecutar tests por capa
go test ./internal/domain/...
go test ./internal/services/...
go test ./internal/repositories/...
```

---

## Checklist de Validación

Usa este checklist antes de crear un PR:

```markdown
- [ ] ¿Las entities están en internal/domain/entity/?
- [ ] ¿Las entities NO tienen tags de DB ni JSON?
- [ ] ¿Los DTOs están en internal/domain/dto/request o response/?
- [ ] ¿Las interfaces están en internal/domain/interfaces/?
- [ ] ¿La lógica de negocio está en internal/services/?
- [ ] ¿Los handlers SOLO parsean requests y llaman services?
- [ ] ¿Los repositories implementan interfaces del domain?
- [ ] ¿Uso mappers para convertir DB models ↔ entities?
- [ ] ¿Las integraciones externas están en infrastructure/?
- [ ] ¿No hay dependencias circulares?
- [ ] ¿Domain layer es independiente?
- [ ] ¿Escribí tests para mi código?
```

---

## Anti-Patrones Comunes

### ❌ NO hagas esto:

```go
// Entity con tags de DB
type User struct {
    ID string `gorm:"primaryKey"` // ❌ NUNCA
}

// Handler con lógica de negocio
func (h *Handler) Create(w, r) {
    if user.Age < 18 { // ❌ Esto es lógica de negocio
        return error
    }
}

// Service accediendo a DB directamente
func (s *Service) GetUser(id) {
    db.Where("id = ?", id).First(&user) // ❌ Usa repository
}

// Repository retornando DB models
func (r *Repo) GetUser(id) (*models.User, error) { // ❌ Retorna entity
}
```

### ✅ SÍ haz esto:

```go
// Entity limpia
type User struct {
    ID  string
    Age int
}

func (u *User) IsAdult() bool { // ✅ Lógica en entity
    return u.Age >= 18
}

// Handler delegando
func (h *Handler) Create(w, r) {
    user, err := h.service.CreateUser(req) // ✅ Delegar a service
}

// Service usando repository
func (s *Service) GetUser(id) {
    return s.repo.GetByID(id) // ✅ Usa repository
}

// Repository con mapper
func (r *Repo) GetUser(id) (*entity.User, error) {
    var model models.User
    db.First(&model, id)
    return mappers.ToEntity(&model), nil // ✅ Retorna entity
}
```

---

## Recursos Adicionales

- **CLAUDE.md** - Documentación general del proyecto
- **.claude-code/skills/clean-arch.md** - Skill completo con reglas detalladas
- **.claude-code/skills/CHEATSHEET.md** - Referencia rápida y templates
- **Clean Architecture Blog** - https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
