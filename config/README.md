# Sistema de Configuración con Viper

Este proyecto usa [Viper](https://github.com/spf13/viper) para gestionar la configuración de forma robusta y flexible.

## Características

- ✅ Múltiples ambientes (local, test, prod)
- ✅ Variables de entorno con máxima prioridad
- ✅ Sintaxis `${VAR_NAME:default}` para valores por defecto
- ✅ **Merge parcial**: solo define overrides en archivos de ambiente
- ✅ Merge inteligente profundo de archivos YAML
- ✅ Type-safe con structs de Go
- ✅ **Fail-fast**: Hace panic si la configuración no se puede cargar (la aplicación no puede arrancar sin config válida)

## Orden de Prioridad

La configuración se carga en el siguiente orden (de menor a mayor prioridad):

1. **config.yaml** - Configuración base común
2. **config_{env}.yaml** - Configuración específica del ambiente
3. **Variables de entorno** - Máxima prioridad (sobrescribe todo)

## Estructura de Archivos

```
config/
├── config.go             # Lógica de carga y structs
├── config.yaml           # Base común (TODOS los valores con defaults)
├── config_local.yaml     # Solo overrides para desarrollo local
├── config_test.yaml      # Solo overrides para testing
├── config_prod.yaml      # Solo overrides para producción
├── README.md            # Esta documentación
├── MERGE_STRATEGY.md    # Guía detallada de merge parcial
├── BEFORE_AFTER.md      # Comparación visual de merge parcial
└── PANIC_RATIONALE.md   # Por qué usamos panic en Load()
```

**Importante:** Los archivos `config_{env}.yaml` solo contienen valores que quieres sobrescribir. No necesitas repetir toda la configuración en cada archivo. Ver [MERGE_STRATEGY.md](MERGE_STRATEGY.md) para ejemplos detallados.

## Uso Básico

### En tu aplicación

```go
package main

import (
    "fmt"
    "log"
    "os"

    "clean-arq-layout/config"
)

func main() {
    // Cargar configuración (lee APP_ENV automáticamente, hace panic si hay error)
    cfg := config.Load()

    // Usar la configuración
    fmt.Printf("Aplicación: %s\n", cfg.AppName)
    fmt.Printf("Servidor: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
    fmt.Printf("Conectando a DB: %s:%d/%s\n",
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Name)
}
```

**Nota:** `config.Load()` lee automáticamente la variable de entorno `APP_ENV`. Si no está definida, usa "local" por defecto.

```go
// En tu shell
export APP_ENV=prod
go run cmd/main.go  // Usará config_prod.yaml

// Sin definir APP_ENV
go run cmd/main.go  // Usará config_local.yaml (default)
```

### Seleccionar ambiente

```bash
# Desarrollo local (por defecto, usa config_local.yaml)
go run cmd/main.go

# Testing (usa config_test.yaml)
APP_ENV=test go run cmd/main.go

# Producción (usa config_prod.yaml)
APP_ENV=prod go run cmd/main.go
```

**Importante:** `config.Load()` siempre lee la variable de entorno `APP_ENV`. No necesitas pasarla como parámetro.

## Variables de Entorno

### Sintaxis en archivos YAML

Los archivos YAML pueden referenciar variables de entorno usando dos sintaxis:

#### Sintaxis 1: Sin valor por defecto
```yaml
database:
  host: "${DB_HOST}"
```
Si `DB_HOST` no está definida, el valor será una cadena vacía.

#### Sintaxis 2: Con valor por defecto
```yaml
database:
  host: "${DB_HOST:localhost}"
  port: 5432
```
Si `DB_HOST` no está definida, usa `localhost`.

### Sobrescribir valores con ENV

Las variables de entorno siempre tienen la máxima prioridad. Viper convierte automáticamente `.` en `_`:

```bash
# Sobrescribir server.port
export SERVER_PORT=9000

# Sobrescribir database.host
export DATABASE_HOST=prod-db.example.com

# Sobrescribir database.password
export DATABASE_PASSWORD=secret123
```

## Ejemplos por Ambiente

### Desarrollo Local

```bash
# Sin variables de entorno (usa valores de config_local.yaml)
go run cmd/main.go
# Resultado: localhost:3000, DB local

# Con algunas variables sobrescritas
export DATABASE_HOST=192.168.1.100
go run cmd/main.go
# Resultado: localhost:3000, DB en 192.168.1.100
```

### Testing

```bash
APP_ENV=test go test ./...
# Usa config_test.yaml
# Log level: debug, swagger disabled
```

### Producción

```bash
# Establecer todas las variables necesarias
export APP_ENV=prod
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080
export DB_HOST=prod-db.internal
export DB_USERNAME=app_user
export DB_PASSWORD=supersecret
export DB_NAME=production_db
export MONGO_URL=mongodb://prod-mongo:27017
export MONGO_DB=prod_db

go run cmd/main.go
```

## Mejores Prácticas

### ✅ HACER

1. **Usar merge parcial - solo define lo que cambia**
   ```yaml
   # config.yaml - TODOS los valores
   server:
     host: "0.0.0.0"
     port: 3000

   # config_local.yaml - SOLO lo que cambia
   server:
     host: "localhost"
   # port: 3000 se hereda automáticamente de config.yaml
   ```

2. **Usar variables de entorno para datos sensibles**
   ```yaml
   database:
     password: "${DB_PASSWORD}"  # NUNCA hardcodear passwords
   ```

3. **Proporcionar valores por defecto razonables para desarrollo**
   ```yaml
   # En config.yaml
   server:
     port: "${SERVER_PORT:3000}"  # 3000 por defecto
   ```

4. **En producción, NO usar defaults para datos sensibles**
   ```yaml
   # config_prod.yaml
   database:
     password: "${DB_PASSWORD}"  # Sin :default - FUERZA ENV var
   ```

5. **Documentar todas las variables de entorno necesarias**
   ```bash
   # Crear un archivo .env.example
   DB_HOST=localhost
   DB_USERNAME=postgres
   DB_PASSWORD=changeme
   ```

### ❌ EVITAR

1. **NO repetir valores en archivos de ambiente**
   ```yaml
   # ❌ MAL - config_local.yaml
   app_name: "clean-arq-layout"  # Innecesario, ya está en config.yaml
   server:
     host: "localhost"
     port: 3000                  # Innecesario si es igual a config.yaml

   # ✅ BIEN - config_local.yaml
   server:
     host: "localhost"           # Solo lo que cambia
   ```

2. **NO hardcodear secrets en los YAMLs**
   ```yaml
   database:
     password: "secret123"  # ❌ MAL
   ```

3. **NO commitear archivos .env con datos reales**
   ```bash
   # Agregar a .gitignore
   .env
   .env.local
   ```

4. **NO usar config.yaml para valores específicos de ambiente**
   ```yaml
   # config.yaml debe tener defaults razonables para TODOS los ambientes
   # Valores específicos van en config_{env}.yaml
   ```

## Agregar Nuevos Campos

1. **Agregar el campo al struct en config.go**
   ```go
   type Config struct {
       // ... campos existentes
       Redis Redis `mapstructure:"redis"`
   }

   type Redis struct {
       Host string `mapstructure:"host"`
       Port int    `mapstructure:"port"`
   }
   ```

2. **Agregar valores en los archivos YAML**
   ```yaml
   # config.yaml
   redis:
     host: "${REDIS_HOST:localhost}"
     port: 6379
   ```

3. **Usar en tu código**
   ```go
   cfg := config.Load()  // Lee APP_ENV automáticamente
   redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
   ```

## Manejo de Errores

La función `config.Load()` **hace panic** en lugar de retornar error. Esto es intencional:

### ¿Por qué panic?

1. **Fail-fast**: Si la configuración no se puede cargar, la aplicación no puede funcionar
2. **Simplicidad**: No necesitas manejar errores en cada lugar que usa config
3. **Startup time**: Los errores de configuración ocurren al inicio, no en runtime
4. **Claridad**: Un panic con stack trace es más útil que una aplicación en estado inconsistente

```go
// ✅ Correcto - simple y directo
cfg := config.Load()  // Lee APP_ENV automáticamente

// ❌ Ya no es necesario
cfg, err := config.Load(env)
if err != nil {
    log.Fatal(err)
}
```

Si la configuración falla, verás un panic claro indicando el problema:
```
panic: Error leyendo config.yaml: open config/config.yaml: no such file or directory
```

**Para más detalles sobre por qué usamos panic, ver [PANIC_RATIONALE.md](PANIC_RATIONALE.md)**

## Troubleshooting

### Panic: "Error leyendo config.yaml"

El archivo `config.yaml` base es obligatorio. Verifica que existe en el directorio `config/`.

### Variables de entorno no se aplican

1. Verifica el nombre de la variable (usa `_` en lugar de `.`):
   - `server.port` → `SERVER_PORT`
   - `database.host` → `DATABASE_HOST`

2. Exporta la variable antes de ejecutar:
   ```bash
   export DATABASE_HOST=myhost
   go run cmd/main.go
   ```

### Valores no se expanden

Asegúrate de usar la sintaxis correcta:
- ✅ `"${VAR_NAME}"`
- ✅ `"${VAR_NAME:default}"`
- ❌ `"$VAR_NAME"`
- ❌ `"{{VAR_NAME}}"`

## Referencias

- [Viper Documentation](https://github.com/spf13/viper)
- [12 Factor App - Config](https://12factor.net/config)
