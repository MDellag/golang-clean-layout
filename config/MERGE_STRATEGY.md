# Estrategia de Merge de Configuración

Este documento explica cómo funciona el merge de archivos de configuración.

## Concepto Principal

El sistema usa **merge parcial**:
- `config.yaml` contiene TODOS los valores base/defaults
- `config_{env}.yaml` contiene SOLO los valores que quieres sobrescribir
- No necesitas repetir toda la configuración en cada ambiente

## Cómo Funciona el Merge

### Paso 1: Cargar Base
```yaml
# config.yaml (BASE)
app_name: "clean-arq-layout"
log_level: "info"
server:
  host: "0.0.0.0"
  port: 3000
database:
  host: "${DB_HOST:localhost}"
  port: 5432
  username: "${DB_USERNAME:postgres}"
  password: "${DB_PASSWORD:postgres}"
  name: "${DB_NAME:myapp}"
```

### Paso 2: Merge con Ambiente
```yaml
# config_local.yaml (SOLO OVERRIDES)
log_level: "debug"         # ✅ Sobrescribe
server:
  host: "localhost"        # ✅ Sobrescribe solo server.host
                          # ❌ NO necesitas definir server.port
                          # (se mantiene 3000 de config.yaml)
```

### Paso 3: Resultado Final
```yaml
# Configuración resultante para ambiente "local"
app_name: "clean-arq-layout"  # ← de config.yaml
log_level: "debug"            # ← de config_local.yaml ✅
server:
  host: "localhost"           # ← de config_local.yaml ✅
  port: 3000                  # ← de config.yaml (heredado)
database:                     # ← TODO de config.yaml (heredado)
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "myapp"
```

## Ejemplos por Ambiente

### 🏠 Local (Desarrollo)

**config_local.yaml:**
```yaml
log_level: "debug"
server:
  host: "localhost"
```

**Resultado:**
- ✅ Log level: `debug` (sobrescrito)
- ✅ Server host: `localhost` (sobrescrito)
- ✅ Server port: `3000` (heredado de config.yaml)
- ✅ Database: todos los valores heredados de config.yaml
- ✅ Swagger: todos los valores heredados de config.yaml
- ✅ Mongo: todos los valores heredados de config.yaml

### 🧪 Test

**config_test.yaml:**
```yaml
log_level: "debug"
server:
  port: 8080
database:
  username: "${DB_USERNAME:test_user}"
  password: "${DB_PASSWORD:test_pass}"
  name: "${DB_NAME:myapp_test}"
swagger:
  enabled: false
mongo:
  url: "${MONGO_URL:mongodb://test:test@localhost:27017}"
  db: "${MONGO_DB:test_db}"
```

**Resultado:**
- ✅ Log level: `debug` (sobrescrito)
- ✅ Server port: `8080` (sobrescrito)
- ✅ Server host: `0.0.0.0` (heredado de config.yaml)
- ✅ Database: username, password, name sobrescritos
- ✅ Database: host, port heredados de config.yaml
- ✅ Swagger: enabled sobrescrito, hostname heredado
- ✅ Mongo: url, db sobrescritos

### 🚀 Producción

**config_prod.yaml:**
```yaml
log_level: "warn"
server:
  port: 8080
database:
  host: "${DB_HOST}"        # Sin default - REQUIERE ENV var
  username: "${DB_USERNAME}"
  password: "${DB_PASSWORD}"
  name: "${DB_NAME}"
swagger:
  enabled: false
mongo:
  url: "${MONGO_URL}"       # Sin default - REQUIERE ENV var
  db: "${MONGO_DB}"
```

**Resultado:**
- ✅ Log level: `warn` (sobrescrito)
- ✅ Server port: `8080` (sobrescrito)
- ✅ Server host: `0.0.0.0` (heredado de config.yaml)
- ✅ Database: valores vienen de ENV vars (sin defaults)
- ✅ Swagger: enabled sobrescrito, hostname heredado
- ✅ Mongo: valores vienen de ENV vars (sin defaults)

## Merge de Estructuras Anidadas

El merge es **profundo** (deep merge), no superficial:

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 3000
  timeout: 30

# config_prod.yaml
server:
  port: 8080

# RESULTADO FINAL
server:
  host: "0.0.0.0"     # ← heredado
  port: 8080          # ← sobrescrito
  timeout: 30         # ← heredado
```

## Ventajas de esta Estrategia

### ✅ DRY (Don't Repeat Yourself)
No repites configuración innecesariamente. Solo defines lo que cambia.

**❌ Antes (repetición):**
```yaml
# config_local.yaml - 50 líneas
# config_test.yaml - 50 líneas
# config_prod.yaml - 50 líneas
# Total: 150 líneas (mucha duplicación)
```

**✅ Ahora (solo overrides):**
```yaml
# config.yaml - 30 líneas (base)
# config_local.yaml - 3 líneas (solo overrides)
# config_test.yaml - 15 líneas (solo overrides)
# config_prod.yaml - 15 líneas (solo overrides)
# Total: 63 líneas (menos duplicación, más claro)
```

### ✅ Mantenibilidad
Si cambias un valor default, solo lo cambias en `config.yaml`:

```yaml
# Cambiar puerto default de 3000 a 4000
# config.yaml
server:
  port: 4000  # ← Un solo lugar

# Automáticamente afecta a todos los ambientes que no lo sobrescriben
```

### ✅ Claridad
Al leer `config_prod.yaml`, sabes exactamente qué es diferente en producción:

```yaml
# config_prod.yaml
log_level: "warn"          # Producción usa warn, no info
server:
  port: 8080              # Producción usa 8080, no 3000
swagger:
  enabled: false          # Swagger disabled en prod
# ... solo lo que cambia
```

### ✅ Menos Errores
Reduces riesgo de tener valores inconsistentes entre ambientes:

**❌ Con repetición:**
```yaml
# config_local.yaml
database:
  port: 5432

# config_prod.yaml
database:
  port: 5433  # ← Typo! Debería ser 5432
```

**✅ Con merge:**
```yaml
# config.yaml
database:
  port: 5432  # ← Definido una vez

# config_prod.yaml
# (no necesita definir port, se hereda correctamente)
```

## Cuándo Sobrescribir

### ✅ SOBRESCRIBIR cuando:
- El valor debe ser diferente en ese ambiente
- Quieres hacer explícito un comportamiento específico
- Necesitas remover defaults para forzar ENV vars (producción)

### ❌ NO SOBRESCRIBIR cuando:
- El valor es el mismo que en config.yaml
- Quieres usar el comportamiento default
- No hay razón específica del ambiente para cambiarlo

## Ejemplo Completo

### config.yaml (Base Común)
```yaml
app_name: "clean-arq-layout"
log_level: "info"
server:
  host: "0.0.0.0"
  port: 3000
  read_timeout: 30
  write_timeout: 30
database:
  host: "${DB_HOST:localhost}"
  port: 5432
  username: "${DB_USERNAME:postgres}"
  password: "${DB_PASSWORD:postgres}"
  name: "${DB_NAME:myapp}"
  max_connections: 10
  connection_timeout: 5
swagger:
  hostname: "${SWAGGER_HOSTNAME:localhost}"
  enabled: true
  base_path: "/api"
mongo:
  url: "${MONGO_URL:mongodb://user:123@localhost:27017}"
  db: "${MONGO_DB:db}"
  timeout: 10
```

### config_local.yaml (Solo 2 overrides)
```yaml
log_level: "debug"
server:
  host: "localhost"
```

### config_test.yaml (Solo lo necesario para tests)
```yaml
log_level: "debug"
server:
  port: 8080
database:
  name: "${DB_NAME:myapp_test}"
swagger:
  enabled: false
mongo:
  db: "${MONGO_DB:test_db}"
```

### config_prod.yaml (Seguridad + optimizaciones)
```yaml
log_level: "warn"
server:
  port: 8080
  read_timeout: 60
  write_timeout: 60
database:
  host: "${DB_HOST}"              # Sin default
  username: "${DB_USERNAME}"      # Sin default
  password: "${DB_PASSWORD}"      # Sin default
  name: "${DB_NAME}"              # Sin default
  max_connections: 50             # Más conexiones
swagger:
  enabled: false
mongo:
  url: "${MONGO_URL}"             # Sin default
  db: "${MONGO_DB}"               # Sin default
```

## Verificar el Merge

Para ver qué valores finales tienes en cada ambiente:

```bash
# Ver configuración de local
go run examples/config_usage.go

# Ver configuración de test
APP_ENV=test go run examples/config_usage.go

# Ver configuración de prod (con ENVs)
APP_ENV=prod \
  DB_HOST=prod-db.internal \
  DB_USERNAME=app_user \
  DB_PASSWORD=secret \
  DB_NAME=production \
  MONGO_URL=mongodb://prod:secret@mongo:27017 \
  MONGO_DB=prod_db \
  go run examples/config_usage.go
```

## Conclusión

El merge parcial te permite:
- 📝 Definir defaults sensatos una vez en `config.yaml`
- 🎯 Sobrescribir solo lo necesario en `config_{env}.yaml`
- 🔒 Forzar ENV vars en producción (sin defaults)
- 🚀 Mantener configuración DRY y fácil de mantener
- ✨ Ver claramente qué es diferente en cada ambiente
