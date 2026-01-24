# Guía de Contribución

## Bienvenido

Gracias por contribuir a este proyecto! Este documento te ayudará a mantener la calidad y coherencia del código.

## Principios Fundamentales

Este proyecto sigue **Clean Architecture** estrictamente. Antes de contribuir:

1. Lee `CLAUDE.md` - Visión general del proyecto
2. Revisa `.claude-code/skills/clean-arch.md` - Reglas arquitectónicas
3. Consulta `.claude-code/skills/CHEATSHEET.md` - Referencia rápida

## Flujo de Contribución

### 1. Antes de Empezar

```bash
# Clonar el repositorio
git clone <repo-url>
cd clean-arq-layout

# Instalar dependencias
go mod download

# Verificar que todo funciona
go test ./...
```

### 2. Durante el Desarrollo

#### Si usas Claude Code:

```bash
# Usa el skill /clean-arch para guiarte
/clean-arch voy a implementar [tu feature]
```

El skill te ayudará a:
- Decidir dónde colocar archivos
- Seguir las convenciones del proyecto
- Evitar violaciones arquitectónicas

#### Checklist de Desarrollo:

- [ ] Leí la documentación de Clean Architecture
- [ ] Entiendo dónde va cada tipo de código
- [ ] Uso `/clean-arch` para validar mi enfoque
- [ ] Sigo las convenciones de nombres de Go
- [ ] Escribí tests para mi código
- [ ] Ejecuté `./scripts/validate_architecture.sh`
- [ ] Todos los tests pasan

### 3. Estructura de Commits

Usa [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add user authentication
fix: resolve race condition in worker pool
docs: update Clean Architecture guidelines
refactor: extract payment logic to service layer
test: add integration tests for orders
```

**Tipos:**
- `feat`: Nueva feature
- `fix`: Bug fix
- `docs`: Cambios en documentación
- `refactor`: Refactorización sin cambio de funcionalidad
- `test`: Agregar o modificar tests
- `chore`: Tareas de mantenimiento

### 4. Pull Requests

#### Antes de crear el PR:

```bash
# Ejecutar validaciones
go fmt ./...
go vet ./...
go test ./...
./scripts/validate_architecture.sh
```

#### Template de PR:

```markdown
## Descripción
[Describe qué cambia y por qué]

## Tipo de cambio
- [ ] Bug fix
- [ ] Nueva feature
- [ ] Breaking change
- [ ] Refactorización
- [ ] Documentación

## Checklist
- [ ] Código sigue Clean Architecture
- [ ] Tests agregados/actualizados
- [ ] Documentación actualizada
- [ ] `validate_architecture.sh` pasa
- [ ] Todos los tests pasan
- [ ] Sin warnings de `go vet`

## Estructura de archivos creados/modificados
[Lista de archivos y su propósito]

## Testing
[Cómo testear los cambios]
```

## Reglas de Oro

### ✅ DO (Hacer)

1. **Usa el skill `/clean-arch`** cuando tengas dudas
2. **Escribe tests** para tu código
3. **Sigue la estructura** de carpetas existente
4. **Usa interfaces** para dependencies
5. **Valida** con `validate_architecture.sh` antes de commit
6. **Documenta** funciones públicas
7. **Maneja errores** apropiadamente con context

### ❌ DON'T (No Hacer)

1. **NO pongas** lógica de negocio en handlers
2. **NO accedas** a DB directamente desde services
3. **NO uses** tags de DB en domain entities
4. **NO retornes** database models desde repositories
5. **NO crees** dependencias circulares
6. **NO ignores** las interfaces definidas
7. **NO hagas** commits sin ejecutar tests

## Guía por Tipo de Contribución

### Agregar Nueva Entity

1. **Domain Entity:** `internal/domain/entity/product.go`
   ```go
   type Product struct {
       ID   string
       Name string
   }

   func (p *Product) Validate() error { /* ... */ }
   ```

2. **DTOs:** `internal/domain/dto/request/product.go` y `response/product.go`

3. **Interfaces:**
   - Agrega a `internal/domain/interfaces/repositories.go`
   - Agrega a `internal/domain/interfaces/services.go`

4. **Service:** `internal/services/products.go`

5. **Repository:** `internal/repositories/postgres/products.go`
   - Modelo: `internal/repositories/postgres/models/product.go`
   - Mapper: `internal/repositories/postgres/mappers/product.go`

6. **Handler:** `internal/delivery/handlers/products.go`

7. **Routes:** Actualiza `internal/delivery/router/router.go`

8. **DI:** Registra en `internal/dependencies/`

9. **Tests:**
   - Unit: `internal/services/products_test.go`
   - Integration: `tests/integration/products_test.go`

### Agregar Integración Externa

1. **Infrastructure:** `internal/infrastructure/twilio/client.go`

2. **Interface (opcional):** `internal/domain/interfaces/infrastructure.go`

3. **Config:** Agrega variables en `config/config.go`

4. **DI:** Registra en `internal/dependencies/`

5. **Tests:** Mockea la interface en tests

### Refactorizar Código

1. **Identifica** violaciones arquitectónicas con `validate_architecture.sh`

2. **Planea** la refactorización usando `/clean-arch`

3. **Refactoriza** en pasos pequeños:
   - Extrae interfaces primero
   - Mueve lógica a la capa correcta
   - Actualiza tests
   - Verifica que todo funciona

4. **Valida** que no rompiste nada

## Convenciones de Código

### Nombres

```go
// Packages: lowercase, single word
package user

// Files: snake_case
user_service.go
user_repository.go

// Types: PascalCase
type UserService struct {}

// Interfaces: PascalCase, often with -er suffix
type UserRepository interface {}

// Functions/Methods: camelCase (unexported), PascalCase (exported)
func createUser() {}  // unexported
func CreateUser() {}  // exported

// Constants: PascalCase or SCREAMING_SNAKE_CASE
const MaxRetries = 3
const DEFAULT_TIMEOUT = 30
```

### Organización de Archivos

```go
// 1. Package declaration
package services

// 2. Imports (grouped: stdlib, external, internal)
import (
    "context"
    "fmt"

    "github.com/google/uuid"

    "yourproject/internal/domain/entity"
    "yourproject/internal/domain/interfaces"
)

// 3. Constants
const maxRetries = 3

// 4. Types
type userService struct {
    repo interfaces.UserRepository
}

// 5. Constructor
func NewUserService(repo interfaces.UserRepository) interfaces.UserService {
    return &userService{repo: repo}
}

// 6. Methods (grouped by receiver)
func (s *userService) Create(ctx context.Context, user *entity.User) error {
    // ...
}
```

### Error Handling

```go
// ✅ Good: Wrap errors with context
if err := s.repo.Create(ctx, user); err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ❌ Bad: Don't ignore or swallow errors
s.repo.Create(ctx, user) // ❌

// ✅ Good: Custom domain errors
var ErrUserNotFound = errors.New("user not found")

// ✅ Good: Check for specific errors
if errors.Is(err, ErrUserNotFound) {
    // handle
}
```

### Context Usage

```go
// ✅ Good: Context as first parameter
func (s *Service) GetUser(ctx context.Context, id string) (*User, error)

// ✅ Good: Pass context through
user, err := s.repo.FindByID(ctx, id)

// ❌ Bad: Don't use context.Background() in application code
ctx := context.Background() // Only in main() or tests
```

## Testing

### Unit Tests

```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := new(mocks.UserRepository)
    service := services.NewUserService(mockRepo)

    // Act
    result, err := service.CreateUser(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Tests

```go
func TestUserAPI_CreateUser(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    // Test actual integration
    resp := makeRequest("/users", userData)
    assert.Equal(t, 201, resp.StatusCode)
}
```

### Coverage

```bash
# Generar reporte de coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Meta: >80% coverage para código de negocio

## Debugging

### Common Issues

**Error: "import cycle not allowed"**
- Revisa flujo de dependencias
- Domain no debe importar otras capas
- Usa interfaces para romper ciclos

**Error: "validate_architecture.sh fails"**
- Lee el mensaje de error
- Consulta `.claude-code/skills/clean-arch.md`
- Usa `/clean-arch` para guiarte

**Tests fallan después de refactor**
- Actualiza mocks
- Verifica que interfaces no cambiaron
- Revisa dependency injection

## Recursos

### Documentación del Proyecto
- `CLAUDE.md` - Overview y comandos
- `.claude-code/skills/clean-arch.md` - Reglas completas
- `.claude-code/skills/CHEATSHEET.md` - Referencia rápida
- `.claude-code/skills/EXAMPLES.md` - Ejemplos prácticos

### Clean Architecture
- [The Clean Architecture - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Clean Architecture in Go](https://medium.com/@hatajoe/clean-architecture-in-go-4030f11ec1b1)

### Go Best Practices
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

## Preguntas?

Si tienes preguntas sobre:
- **Arquitectura:** Usa `/clean-arch` en Claude Code
- **Bugs:** Abre un issue
- **Features:** Abre un issue para discusión primero

## Licencia

Al contribuir, aceptas que tus contribuciones se licencien bajo la misma licencia del proyecto.

---

**Gracias por contribuir y mantener la calidad del código!** 🚀
