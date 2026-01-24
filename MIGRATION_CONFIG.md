# Migración al Sistema de Configuración con Viper

Este documento describe la migración del sistema de configuración de `envconfig` a `Viper`.

## Cambios Realizados

### 1. Dependencias Agregadas

- `github.com/spf13/viper v1.21.0` - Sistema de configuración flexible

### 2. Archivos Nuevos

#### Configuración Principal
- `config/config.go` - Lógica de carga con Viper (reemplaza implementación anterior con envconfig)
- `config/README.md` - Documentación completa del sistema de configuración

#### Archivos YAML
- `config/config.yaml` - Configuración base común para todos los ambientes
- `config/config_local.yaml` - Configuración para desarrollo local
- `config/config_test.yaml` - Configuración para testing
- `config/config_prod.yaml` - Configuración para producción

#### Ejemplos y Documentación
- `examples/config_usage.go` - Ejemplo de uso del sistema de configuración
- `.env.example` - Plantilla de variables de entorno
- `MIGRATION_CONFIG.md` - Este archivo

### 3. Archivos Modificados

#### `cmd/main.go`
**Antes:**
```go
func main() {
    app.Start()
}
```

**Después:**
```go
func main() {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "local"
        log.Printf("APP_ENV no definida, usando ambiente: %s", env)
    } else {
        log.Printf("Iniciando aplicación en ambiente: %s", env)
    }
    app.Start(env)
}
```

#### `internal/app/start.go`
**Antes:**
```go
func Start() {
    container := dig.New()
    err := container.Provide(func() *config.Config {
        cfg := config.Get()
        return &cfg
    })
    // ... resto del código
    err = container.Provide(func(cfg *config.Config) (*mongo.Client, error) {
        return mongo.NewClient(cfg.Mongo.Url, cfg.Mongo.DB)
    })
}
```

**Después:**
```go
func Start(env string) {
    container := dig.New()
    err := container.Provide(func() (*config.Config, error) {
        return config.Load(env)
    })
    // ... resto del código
    err = container.Provide(func(cfg *config.Config) (*mongo.Client, error) {
        return mongo.NewClient(cfg.Mongo.URL, cfg.Mongo.DB)
    })
}
```

**Cambios clave:**
- `Start()` ahora acepta parámetro `env string`
- `config.Get()` reemplazado por `config.Load(env)`
- `cfg.Mongo.Url` cambió a `cfg.Mongo.URL` (mayúsculas)

#### `.gitignore`
Agregadas entradas para archivos sensibles:
```
.env
.env.local
.env.*.local
app
*.exe
*.dll
*.so
*.dylib
*.out
coverage.txt
```

## Estructura del Sistema de Configuración

### Orden de Prioridad (de menor a mayor)

1. **config.yaml** - Base común
2. **config_{env}.yaml** - Específico del ambiente (local, test, prod)
3. **Variables de entorno** - Máxima prioridad

### Sintaxis de Variables de Entorno en YAML

```yaml
# Con valor por defecto
database:
  host: "${DB_HOST:localhost}"

# Sin valor por defecto (cadena vacía si no existe)
database:
  password: "${DB_PASSWORD}"
```

### Mapeo de Variables de Entorno

Viper convierte automáticamente `.` en `_`:

- `server.port` → `SERVER_PORT`
- `database.host` → `DATABASE_HOST`
- `database.password` → `DATABASE_PASSWORD`

## Migración de Código Existente

### Cambio en la Estructura Config

**Antes (envconfig):**
```go
type Config struct {
    Port    int    `required:"true" default:"3000"`
    Swagger Swagger
    Mongo   Mongo
}
```

**Después (Viper):**
```go
type Config struct {
    AppName  string   `mapstructure:"app_name"`
    LogLevel string   `mapstructure:"log_level"`
    Server   Server   `mapstructure:"server"`
    Database Database `mapstructure:"database"`
    Swagger  Swagger  `mapstructure:"swagger"`
    Mongo    Mongo    `mapstructure:"mongo"`
}
```

**Cambios importantes:**
1. Tags cambian de `envconfig` a `mapstructure`
2. Campos anidados requieren mapstructure tags
3. Nombres de variables usan snake_case en YAML

### Actualización de Código que Usa Config

**Antes:**
```go
cfg := config.Get()
port := cfg.Port
mongoURL := cfg.Mongo.Url
```

**Después:**
```go
// config.Load ya no retorna error - hace panic si falla
// Lee APP_ENV automáticamente desde las variables de entorno
cfg := config.Load()
port := cfg.Server.Port
mongoURL := cfg.Mongo.URL  // Nota: URL en mayúsculas
```

## Pasos para Migrar tu Código

### 1. Actualizar referencias a Config

Buscar y reemplazar en tu código:
- `cfg.Port` → `cfg.Server.Port`
- `cfg.Mongo.Url` → `cfg.Mongo.URL`

### 2. Actualizar funciones que usan Config

**Antes:**
```go
func NewService() *Service {
    cfg := config.Get()
    // ...
}
```

**Después:**
```go
func NewService(cfg *config.Config) *Service {
    // Recibir config como parámetro vía dependency injection
    // Nota: config.Load() lee APP_ENV automáticamente y hace panic si falla
    // ...
}
```

### 3. Usar Dependency Injection

El sistema ahora usa DI con `dig`. La configuración se inyecta automáticamente:

```go
container.Provide(func(cfg *config.Config) *MyService {
    return NewMyService(cfg)
})
```

## Uso en Diferentes Ambientes

### Desarrollo Local
```bash
# Usa config_local.yaml
go run cmd/main.go

# O explícitamente
APP_ENV=local go run cmd/main.go
```

### Testing
```bash
# Usa config_test.yaml
APP_ENV=test go test ./...
```

### Producción
```bash
# Usa config_prod.yaml + variables de entorno
export APP_ENV=prod
export DB_HOST=prod-db.internal
export DB_PASSWORD=secret
export SERVER_PORT=8080
go run cmd/main.go
```

## Ventajas del Nuevo Sistema

### ✅ Merge Parcial (Principal Ventaja)
- **No necesitas repetir toda la configuración en cada ambiente**
- `config.yaml` tiene TODOS los valores base
- `config_{env}.yaml` solo tiene los valores que cambian
- Reduce duplicación de ~150 líneas a ~60 líneas en el proyecto
- Hace explícito qué es diferente en cada ambiente

```yaml
# config_local.yaml - Solo 3 líneas en lugar de 50
log_level: "debug"
server:
  host: "localhost"
```

### ✅ Flexibilidad
- Múltiples formatos de config (YAML, JSON, TOML, ENV)
- Merge inteligente profundo de estructuras anidadas
- Hot reload de configuración (con fsnotify)

### ✅ Seguridad
- Secrets en variables de entorno, no en archivos
- Valores por defecto solo para desarrollo
- Producción sin defaults - fuerza uso de ENV vars
- Fácil integración con servicios de secrets (AWS Secrets Manager, Vault, etc.)

### ✅ Mantenibilidad
- DRY: cambias un default en un solo lugar
- Configuración declarativa en YAML
- Separación clara entre ambientes
- Merge profundo de estructuras anidadas
- Documentación integrada

### ✅ Developer Experience
- Valores por defecto razonables para desarrollo
- Fácil sobrescritura con ENV vars
- Validación de tipos en tiempo de compilación
- Claro qué valores son específicos de cada ambiente

## Troubleshooting

### Error: "error leyendo config.yaml"

**Problema:** El archivo config.yaml no existe o no se encuentra.

**Solución:**
```bash
# Verificar que existe
ls config/config.yaml

# Verificar el directorio de ejecución
pwd
```

### Variables de entorno no se aplican

**Problema:** Los valores de ENV no sobrescriben la config.

**Solución:**
1. Usar nombres correctos (guiones bajos en lugar de puntos)
2. Exportar las variables:
```bash
export DATABASE_HOST=myhost
go run cmd/main.go
```

### Valores vacíos en producción

**Problema:** En producción algunos valores están vacíos.

**Solución:**
1. Verificar que todas las ENV vars necesarias están definidas
2. Revisar config_prod.yaml - NO debe tener defaults para secrets
3. Usar .env.example como checklist

## Ver Merge Parcial en Acción

```bash
# Demostración interactiva del merge
go run examples/merge_demo.go

# Muestra qué valores vienen de config.yaml (heredados)
# vs qué valores vienen de config_{env}.yaml (sobrescritos)
```

## Referencias

- [Documentación completa](config/README.md)
- [Estrategia de merge parcial](config/MERGE_STRATEGY.md) ⭐ IMPORTANTE
- [Ejemplo de uso básico](examples/config_usage.go)
- [Demostración de merge](examples/merge_demo.go)
- [Viper Documentation](https://github.com/spf13/viper)
