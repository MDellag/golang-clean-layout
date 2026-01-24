# Por Qué Usamos Panic en config.Load()

Este documento explica la decisión de diseño de usar `panic` en lugar de retornar error en `config.Load()`.

## TL;DR

**La configuración es crítica para el arranque de la aplicación. Si falla, la aplicación no puede funcionar. Usar panic es más simple y seguro que retornar error.**

## Comparación

### ❌ Retornando Error (Antes)

```go
cfg, err := config.Load(env)
if err != nil {
    log.Fatal(err)  // Termina la aplicación de todas formas
}
```

### ✅ Con Panic (Ahora)

```go
cfg := config.Load()  // Lee APP_ENV automáticamente
// Si falla, panic automáticamente
```

## Razones para Usar Panic

### 1. Fail-Fast es Mejor

La configuración se carga **una vez al inicio**. Si falla:
- La aplicación no puede arrancar correctamente
- No tiene sentido continuar en estado inconsistente
- Mejor terminar inmediatamente con mensaje claro

```go
// ❌ Malo: aplicación arranca parcialmente sin config válida
cfg, err := config.Load(env)
if err != nil {
    // ¿Qué hacemos aquí? ¿Usar valores default? ¿Continuar sin DB?
    // Estado inconsistente y peligroso
}

// ✅ Bueno: si no hay config válida, panic inmediato
cfg := config.Load()  // Lee APP_ENV automáticamente
// Si llegamos aquí, sabemos que cfg es válido
```

### 2. Simplicidad del Código

No necesitas manejar errores en cada lugar que carga config:

```go
// ❌ Verboso - 4 líneas para cargar config
cfg, err := config.Load(env)
if err != nil {
    log.Fatal(err)
}

// ✅ Conciso - 1 línea
cfg := config.Load()  // Lee APP_ENV automáticamente
```

### 3. Errores de Configuración son Bugs, no Casos de Uso

Los errores de configuración son:
- ❌ Archivo faltante
- ❌ YAML mal formado
- ❌ Variables de entorno requeridas no definidas

Estos son **errores de deployment/setup**, no casos de negocio que debas manejar gracefully.

### 4. Stack Trace Útil

Panic proporciona un stack trace completo mostrando dónde falló:

```
panic: Error leyendo config.yaml: open config/config.yaml: no such file or directory

goroutine 1 [running]:
clean-arq-layout/config.Load(...)
    /app/config/config.go:65
main.main()
    /app/cmd/main.go:15
```

Más útil que:
```
Error loading config: open config/config.yaml: no such file or directory
exit status 1
```

### 5. No Hay Recuperación Posible

¿Qué harías si config.Load() retorna error?

```go
cfg, err := config.Load()
if err != nil {
    // Opción 1: log.Fatal (igual que panic)
    log.Fatal(err)

    // Opción 2: usar defaults (peligroso - ¿qué DB usar?)
    cfg = getDefaultConfig()

    // Opción 3: reintentar (no tiene sentido, archivo sigue sin existir)
    time.Sleep(1 * time.Second)
    cfg, err = config.Load()
}
```

**Ninguna opción es mejor que panic.** La única opción realista es terminar la aplicación.

### 6. Patrón Común en Go

Otros paquetes de Go usan el mismo patrón:

```go
// regexp.MustCompile - panic si regex inválido
re := regexp.MustCompile(`[a-z]+`)

// template.Must - panic si template inválido
tmpl := template.Must(template.New("").Parse(src))

// config.Load - panic si config inválida
cfg := config.Load()  // Lee APP_ENV automáticamente
```

El sufijo "Must" indica "debe tener éxito o panic".

## Cuándo NO Usar Panic

Panic es apropiado para errores de configuración **al inicio**, pero NO para:

### ❌ Errores de Runtime

```go
// MAL - nunca hagas panic en runtime
func GetUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        panic(err)  // ❌ MAL - esto es error de runtime
    }
    return user, nil
}

// BIEN - retorna error para manejar en runtime
func GetUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return nil, err  // ✅ BIEN - error recuperable
    }
    return user, nil
}
```

### ❌ Errores de Negocio

```go
// MAL - error de validación no es panic
func CreateOrder(amount int) (*Order, error) {
    if amount <= 0 {
        panic("amount must be positive")  // ❌ MAL
    }
}

// BIEN - retorna error de validación
func CreateOrder(amount int) (*Order, error) {
    if amount <= 0 {
        return nil, errors.New("amount must be positive")  // ✅ BIEN
    }
}
```

### ❌ Errores Recuperables

```go
// MAL - errores de red son recuperables
func FetchData() ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        panic(err)  // ❌ MAL - podríamos reintentar
    }
}

// BIEN - permite reintentos
func FetchData() ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err  // ✅ BIEN - caller puede reintentar
    }
}
```

## Regla General

### ✅ USA PANIC cuando:
- Error ocurre **al inicio** de la aplicación
- No hay forma razonable de continuar
- Es un error de **configuración/setup** (bug del desarrollador/ops)
- Ejemplo: `config.Load()`, `regexp.MustCompile()`

### ❌ RETORNA ERROR cuando:
- Error ocurre **durante runtime**
- El caller puede manejar/recuperar del error
- Es un error de **negocio** o **usuario**
- Ejemplo: `db.FindUser()`, `http.Get()`, `ValidateInput()`

## Ejemplo Completo

```go
package main

import (
    "clean-arq-layout/config"
    "log"
)

func main() {
    // ✅ Panic OK - startup time, no hay recuperación
    cfg := config.Load()  // Lee APP_ENV automáticamente

    // Inicializar servicios con la config válida
    db := initDB(cfg.Database)
    server := initServer(cfg.Server)

    // ❌ Errores retornados - runtime, recuperables
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}

// Runtime - retorna error
func (s *Server) HandleRequest(r *Request) (*Response, error) {
    user, err := s.db.FindUser(r.UserID)
    if err != nil {
        return nil, err  // ✅ Retorna error - caller puede manejar
    }
    return &Response{User: user}, nil
}
```

## Preguntas Frecuentes

### ¿No es panic considerado "bad practice" en Go?

**No.** Panic es apropiado para errores irrecuperables. La guía de Go dice:

> "Panic is appropriate when something has gone catastrophically wrong and the program cannot continue."

Config inválida es exactamente esto - la aplicación no puede continuar.

### ¿Qué pasa si quiero hacer testing?

En tests, puedes usar `defer` y `recover()` si necesitas capturar panics:

```go
func TestConfigPanic(t *testing.T) {
    defer func() {
        if r := recover(); r != nil {
            // Panic esperado
            if !strings.Contains(fmt.Sprint(r), "config.yaml") {
                t.Errorf("unexpected panic: %v", r)
            }
        }
    }()

    // Configurar un ambiente inválido
    os.Setenv("APP_ENV", "nonexistent")

    // Esto debería hacer panic
    config.Load()
    t.Error("expected panic")
}
```

### ¿Y si necesito config dinámica en runtime?

Para config dinámica (feature flags, etc.), usa un sistema diferente que retorne error:

```go
// Startup config - panic OK
cfg := config.Load()  // Lee APP_ENV automáticamente

// Runtime config - retorna error
func GetFeatureFlag(name string) (bool, error) {
    flag, err := featureFlagService.Get(name)
    if err != nil {
        return false, err  // ✅ Error recuperable
    }
    return flag.Enabled, nil
}
```

## Conclusión

Usar `panic` en `config.Load()` es:
- ✅ **Simple**: menos código boilerplate
- ✅ **Seguro**: fail-fast previene estados inconsistentes
- ✅ **Apropiado**: config es startup-time, no runtime
- ✅ **Idiomático**: patrón común en Go (Must functions)
- ✅ **Claro**: stack trace muestra exactamente dónde falló

La configuración es la base de tu aplicación. Si falla, es mejor saberlo inmediatamente con un panic claro que continuar en estado indefinido.
