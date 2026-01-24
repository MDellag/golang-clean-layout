# Sistema de Configuración con Viper

Este proyecto usa [Viper](https://github.com/spf13/viper) para gestionar la configuración de forma robusta y flexible.

## Características

- ✅ Múltiples ambientes (local, test, prod)
- ✅ Variables de entorno con máxima prioridad
- ✅ Sintaxis `${VAR_NAME:default}` para valores por defecto
- ✅ **Merge parcial**: solo define overrides en archivos de ambiente
- ✅ Merge inteligente profundo de archivos YAML
- ✅ Type-safe con structs de Go

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
└── MERGE_STRATEGY.md    # Guía detallada de merge parcial
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
    // Leer el ambiente desde variable de entorno
    // Si APP_ENV no está definida, usa "local" por defecto
    env := os.Getenv("APP_ENV")

    // Cargar configuración
    cfg, err := config.Load(env)
    if err != nil {
        log.Fatalf("Error cargando configuración: %v", err)
    }

    // Usar la configuración
    fmt.Printf("Aplicación: %s\n", cfg.AppName)
    fmt.Printf("Servidor: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
    fmt.Printf("Conectando a DB: %s:%d/%s\n",
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Name)
}
```

### Seleccionar ambiente

#### Opción 1: Variable de entorno APP_ENV

```bash
# Desarrollo local (por defecto)
go run cmd/main.go

# Testing
APP_ENV=test go run cmd/main.go

# Producción
APP_ENV=prod go run cmd/main.go
```

#### Opción 2: Pasar directamente al código

```go
cfg, err := config.Load("prod")
```

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
   redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
   ```

## Troubleshooting

### "Error leyendo config.yaml"

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
