# Antes vs Después: Merge Parcial

Comparación visual del antes y después de implementar merge parcial.

## ❌ ANTES: Repetición Completa

Antes, cada archivo tenía TODA la configuración repetida:

### config.yaml (50 líneas)
```yaml
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
swagger:
  hostname: "${SWAGGER_HOSTNAME:localhost}"
  enabled: true
mongo:
  url: "${MONGO_URL:mongodb://user:123@localhost:27017}"
  db: "${MONGO_DB:db}"
```

### config_local.yaml (50 líneas) ❌ REPETICIÓN
```yaml
app_name: "clean-arq-layout"        # ← Repetido
log_level: "debug"                   # ← Realmente diferente
server:
  host: "localhost"                  # ← Realmente diferente
  port: 3000                         # ← Repetido
database:
  host: "${DB_HOST:localhost}"       # ← Repetido
  port: 5432                         # ← Repetido
  username: "${DB_USERNAME:postgres}" # ← Repetido
  password: "${DB_PASSWORD:postgres}" # ← Repetido
  name: "${DB_NAME:myapp_dev}"       # ← Realmente diferente
swagger:
  hostname: "${SWAGGER_HOSTNAME:localhost}" # ← Repetido
  enabled: true                      # ← Repetido
mongo:
  url: "${MONGO_URL:mongodb://user:123@localhost:27017}" # ← Repetido
  db: "${MONGO_DB:dev_db}"           # ← Realmente diferente
```

### config_test.yaml (50 líneas) ❌ REPETICIÓN
```yaml
app_name: "clean-arq-layout"        # ← Repetido
log_level: "debug"                   # ← Repetido de local
server:
  host: "localhost"                  # ← Repetido de local
  port: 8080                         # ← Realmente diferente
database:
  host: "${DB_HOST:localhost}"       # ← Repetido
  port: 5432                         # ← Repetido
  username: "${DB_USERNAME:test_user}" # ← Realmente diferente
  password: "${DB_PASSWORD:test_pass}" # ← Realmente diferente
  name: "${DB_NAME:myapp_test}"      # ← Realmente diferente
swagger:
  hostname: "${SWAGGER_HOSTNAME:localhost}" # ← Repetido
  enabled: false                     # ← Realmente diferente
mongo:
  url: "${MONGO_URL:mongodb://test:test@localhost:27017}" # ← Realmente diferente
  db: "${MONGO_DB:test_db}"          # ← Realmente diferente
```

### config_prod.yaml (50 líneas) ❌ REPETICIÓN
```yaml
app_name: "clean-arq-layout"        # ← Repetido
log_level: "warn"                    # ← Realmente diferente
server:
  host: "0.0.0.0"                    # ← Repetido de base
  port: 8080                         # ← Repetido de test
database:
  host: "${DB_HOST}"                 # ← Realmente diferente
  port: 5432                         # ← Repetido
  username: "${DB_USERNAME}"         # ← Realmente diferente
  password: "${DB_PASSWORD}"         # ← Realmente diferente
  name: "${DB_NAME}"                 # ← Realmente diferente
swagger:
  hostname: "${SWAGGER_HOSTNAME}"    # ← Realmente diferente
  enabled: false                     # ← Repetido de test
mongo:
  url: "${MONGO_URL}"                # ← Realmente diferente
  db: "${MONGO_DB}"                  # ← Realmente diferente
```

**Total: ~200 líneas con mucha duplicación** 😞

### Problemas:
- ❌ 80% de duplicación innecesaria
- ❌ Difícil ver qué es diferente en cada ambiente
- ❌ Si cambias un default, debes cambiar en 4 archivos
- ❌ Propenso a errores (olvidar actualizar un archivo)
- ❌ Valores inconsistentes entre ambientes

---

## ✅ DESPUÉS: Solo Overrides

Ahora, cada archivo solo tiene lo que realmente cambia:

### config.yaml (30 líneas) - Base común
```yaml
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
swagger:
  hostname: "${SWAGGER_HOSTNAME:localhost}"
  enabled: true
mongo:
  url: "${MONGO_URL:mongodb://user:123@localhost:27017}"
  db: "${MONGO_DB:db}"
```

### config_local.yaml (3 líneas) ✅ SOLO OVERRIDES
```yaml
# Solo lo que es diferente en local
log_level: "debug"
server:
  host: "localhost"
```

**Heredado de config.yaml:**
- ✅ app_name: "clean-arq-layout"
- ✅ server.port: 3000
- ✅ database.* (todos los valores)
- ✅ swagger.* (todos los valores)
- ✅ mongo.* (todos los valores)

### config_test.yaml (15 líneas) ✅ SOLO OVERRIDES
```yaml
# Solo lo que es diferente en test
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

**Heredado de config.yaml:**
- ✅ app_name: "clean-arq-layout"
- ✅ server.host: "0.0.0.0"
- ✅ database.host: localhost
- ✅ database.port: 5432
- ✅ swagger.hostname: localhost

### config_prod.yaml (15 líneas) ✅ SOLO OVERRIDES
```yaml
# Solo lo que es diferente en producción
log_level: "warn"
server:
  port: 8080
database:
  host: "${DB_HOST}"
  username: "${DB_USERNAME}"
  password: "${DB_PASSWORD}"
  name: "${DB_NAME}"
swagger:
  enabled: false
mongo:
  url: "${MONGO_URL}"
  db: "${MONGO_DB}"
```

**Heredado de config.yaml:**
- ✅ app_name: "clean-arq-layout"
- ✅ server.host: "0.0.0.0"
- ✅ database.port: 5432
- ✅ swagger.hostname: localhost (debería sobrescribirse con ENV)

**Total: ~63 líneas sin duplicación** 😊

### Ventajas:
- ✅ 68% menos código
- ✅ Cero duplicación
- ✅ Inmediatamente obvio qué es diferente
- ✅ Cambias un default en UN solo lugar
- ✅ Menos propenso a errores
- ✅ Mantenimiento más fácil

---

## Comparación: Cambiar un Default

### ❌ ANTES (con repetición)

Necesitas cambiar el puerto default de 3000 a 4000:

```yaml
# config.yaml
server:
  port: 4000  # ← Cambiar aquí

# config_local.yaml
server:
  port: 4000  # ← Y aquí

# config_test.yaml NO (usa 8080)

# config_prod.yaml NO (usa 8080)
```

**Pasos:** Buscar en 4 archivos, cambiar 2 lugares

### ✅ DESPUÉS (con merge parcial)

```yaml
# config.yaml
server:
  port: 4000  # ← Cambiar SOLO aquí

# config_local.yaml
# (no necesita cambios, hereda automáticamente)

# config_test.yaml
# (sigue usando 8080 como override)

# config_prod.yaml
# (sigue usando 8080 como override)
```

**Pasos:** Cambiar 1 lugar, listo

---

## Comparación: Ver Diferencias de un Ambiente

### ❌ ANTES (con repetición)

Para saber qué es diferente en test vs base:

1. Abrir config.yaml
2. Abrir config_test.yaml
3. Comparar línea por línea (50 líneas vs 50 líneas)
4. Identificar manualmente las diferencias

**Tiempo:** ~2-3 minutos

### ✅ DESPUÉS (con merge parcial)

Para saber qué es diferente en test:

1. Abrir config_test.yaml
2. Ver solo las 15 líneas que contiene
3. **Todo lo que ves es diferente, lo demás es igual**

**Tiempo:** ~10 segundos

---

## Ejemplo Real: Agregar un Nuevo Campo

Necesitas agregar soporte para Redis:

### ❌ ANTES (con repetición)

```yaml
# config.yaml
redis:
  host: "${REDIS_HOST:localhost}"
  port: 6379

# config_local.yaml
redis:
  host: "${REDIS_HOST:localhost}"
  port: 6379

# config_test.yaml
redis:
  host: "${REDIS_HOST:localhost}"
  port: 6379

# config_prod.yaml
redis:
  host: "${REDIS_HOST}"
  port: 6379
```

**4 archivos modificados, mucha repetición**

### ✅ DESPUÉS (con merge parcial)

```yaml
# config.yaml
redis:
  host: "${REDIS_HOST:localhost}"
  port: 6379

# config_local.yaml
# (nada, usa el default)

# config_test.yaml
# (nada, usa el default)

# config_prod.yaml
redis:
  host: "${REDIS_HOST}"  # Solo override: sin default en prod
```

**1-2 archivos modificados, sin repetición**

---

## Conclusión

### Antes: Repetición Total
- 📝 200 líneas de código
- ❌ 80% duplicación
- ⏱️ Más tiempo para mantener
- 🐛 Más propenso a errores
- 🔍 Difícil ver diferencias

### Después: Merge Parcial
- 📝 63 líneas de código (68% menos)
- ✅ 0% duplicación
- ⏱️ Menos tiempo para mantener
- 🐛 Menos propenso a errores
- 🔍 Diferencias obviamente visibles

### Recomendación

**SIEMPRE usa merge parcial:**

1. Define TODOS los valores en `config.yaml`
2. En `config_{env}.yaml`, define SOLO lo que cambia
3. Deja que Viper haga el merge automáticamente
4. Disfruta de código más limpio y mantenible

Ver [MERGE_STRATEGY.md](MERGE_STRATEGY.md) para más ejemplos.
