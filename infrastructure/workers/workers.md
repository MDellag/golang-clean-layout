# Workers Implementation

Sistema de workers genérico para ejecutar código asíncronamente con gestión de go-routines y tracking de métricas.

## Características

- **Pool de Workers Genérico**: Define el tipo de datos que procesarán tus workers mediante generics
- **Dispatchers Intercambiables**: Dos implementaciones según si necesitas respuestas o no
- **Métricas**: Tracking automático de jobs ejecutados, fallidos y tiempos de ejecución
- **Cancelación**: Capacidad de cancelar la ejecución de jobs mediante context
- **Thread-Safe**: Todas las operaciones son seguras para concurrencia
- **Desacoplamiento**: Workers recibe el dispatcher como dependencia
- **Context Management**: El dispatcher maneja su propio context cancelable

## Estructura

```
internal/workers/
├── workers.go           # Workers manager y Dispatcher interface
├── metrics.go           # Sistema de métricas
├── dispatcher/
│   ├── dispatcher.go                # Funciones compartidas y MetricsTracker
│   ├── response_dispatcher.go       # Dispatcher con respuestas
│   └── no_response_dispatcher.go    # Dispatcher sin respuestas
└── jobs/
    ├── job.go          # Interface Job
    └── example_job.go  # Ejemplos de implementaciones
```

## Componentes Principales

### 1. Job Interface

Todos los jobs deben implementar esta interface:

```go
type Job[T any] interface {
    Execute(ctx context.Context) (T, error)
}
```

### 2. Dispatcher Interface

La interface que define cómo se despachan los jobs:

```go
type Dispatcher[T any] interface {
    Dispatch(numWorkers int, jobs []jobs.Job[T]) <-chan dispatcher.JobResult[T]
    DispatchSingle(numWorkers int, job jobs.Job[T]) <-chan dispatcher.JobResult[T]
    Cancel()
}
```

### 3. Workers

El manager de workers que recibe número de workers y dispatcher.

**Nota**: Workers NO es genérico. El type-safety está garantizado por el dispatcher que pasas al constructor. El tipo `T` se infiere automáticamente del dispatcher.

```go
import (
    "context"
    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/dispatcher"
)

// Crear context
ctx := context.Background()

// Crear workers - las métricas se inicializan automáticamente
workerPool := workers.NewWorkers(5, ctx, dispatcher.NewDispatcherWithResult[string])

// Obtener métricas cuando las necesites
metrics := workerPool.GetMetrics()
```

### 4. Tipos de Dispatchers

Dos implementaciones disponibles:

- **DispatcherWithResult**: Para jobs que retornan resultados a través de un channel
- **NoResponseDispatcher**: Para jobs fire-and-forget que no necesitan retornar nada

Ambos reciben:
- `context.Context`: Contexto base del cual crean un context cancelable
- `MetricsTracker`: Interface para tracking de métricas

## Casos de Uso

### Caso 1: Jobs con Respuesta (API Requests)

Este es el caso de uso más común cuando necesitas procesar múltiples requests y obtener sus respuestas.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/dispatcher"
    "clean-arq-layout/internal/workers/jobs"
)

func main() {
    // 1. Crear context
    ctx := context.Background()

    // 2. Crear workers (métricas se inicializan automáticamente)
    workerPool := workers.NewWorkers(5, ctx, dispatcher.NewDispatcherWithResult[string])

    // 3. Crear un slice de jobs
    jobList := make([]jobs.Job[string], 0)
    for i := 1; i <= 10; i++ {
        jobList = append(jobList, &jobs.ExampleJob{
            ID:   i,
            Data: fmt.Sprintf("Task %d", i),
        })
    }

    // 4. Despachar los jobs
    resultChan := workers.Dispatch(workerPool, jobList)

    // 5. Procesar resultados
    errorCount := 0
    for result := range resultChan {
        if result.Error != nil {
            errorCount++
            log.Printf("Job failed: %v", result.Error)

            // Decidir si cancelar el resto de jobs
            if errorCount > 3 {
                workerPool.Cancel()
                log.Println("Too many errors, canceling remaining jobs")
                break
            }
        } else {
            // Ejecutar lógica de negocio con el resultado
            fmt.Printf("Success: %s\n", result.Result)
        }
    }

    // 6. Ver métricas
    finalMetrics := workerPool.GetMetrics()
    fmt.Printf("Total: %d, Completed: %d, Failed: %d\n",
        finalMetrics.TotalJobs, finalMetrics.CompletedJobs, finalMetrics.FailedJobs)
}
```

### Caso 2: Jobs sin Respuesta (Fire-and-Forget)

Para tareas que no necesitan retornar resultados, como enviar notificaciones o logs.

```go
package main

import (
    "context"
    "fmt"

    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/dispatcher"
    "clean-arq-layout/internal/workers/jobs"
)

func main() {
    // 1. Crear workers con dispatcher sin respuesta
    ctx := context.Background()
    workerPool := workers.NewWorkers(3, ctx, dispatcher.NewNoResponseDispatcher[string])

    // 2. Crear jobs
    jobList := make([]jobs.Job[string], 0)
    for i := 1; i <= 5; i++ {
        jobList = append(jobList, &jobs.ExampleJob{
            ID:   i,
            Data: fmt.Sprintf("Notification %d", i),
        })
    }

    // 3. Despachar y esperar completación
    resultChan := workers.Dispatch(workerPool, jobList)

    // Esperar a que todos los jobs terminen
    for range resultChan {
        // Channel se cierra cuando todos los jobs terminan
    }

    // Ver métricas finales
    finalMetrics := workerPool.GetMetrics()
    fmt.Printf("Completed: %d, Failed: %d\n",
        finalMetrics.CompletedJobs, finalMetrics.FailedJobs)
}
```

### Caso 3: Job Individual

Para ejecutar un solo job:

```go
package main

import (
    "context"
    "fmt"

    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/dispatcher"
    "clean-arq-layout/internal/workers/jobs"
)

func main() {
    ctx := context.Background()
    workerPool := workers.NewWorkers(1, ctx, dispatcher.NewDispatcherWithResult[string])

    job := &jobs.ExampleJob{
        ID:   1,
        Data: "Single task",
    }

    resultChan := workers.DispatchSingle(workerPool, job)

    // Obtener el resultado
    result := <-resultChan
    if result.Error != nil {
        fmt.Printf("Error: %v\n", result.Error)
    } else {
        fmt.Printf("Result: %s\n", result.Result)
    }
}
```

### Caso 4: Implementar un Job Personalizado

Ejemplo de un job que hace una request HTTP:

```go
package jobs

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type UserData struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type FetchUserJob struct {
    UserID int
    Client *http.Client
}

func (f *FetchUserJob) Execute(ctx context.Context) (UserData, error) {
    var userData UserData

    url := fmt.Sprintf("https://api.example.com/users/%d", f.UserID)
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return userData, err
    }

    resp, err := f.Client.Do(req)
    if err != nil {
        return userData, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return userData, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return userData, err
    }

    err = json.Unmarshal(body, &userData)
    return userData, err
}
```

Uso:

```go
func main() {
    ctx := context.Background()
    workerPool := workers.NewWorkers(10, ctx, dispatcher.NewDispatcherWithResult[UserData])

    client := &http.Client{}
    jobList := make([]jobs.Job[UserData], 0)

    // Crear jobs para obtener múltiples usuarios
    for userID := 1; userID <= 100; userID++ {
        jobList = append(jobList, &jobs.FetchUserJob{
            UserID: userID,
            Client: client,
        })
    }

    resultChan := workers.Dispatch(workerPool, jobList)

    users := make([]UserData, 0)
    for result := range resultChan {
        if result.Error != nil {
            log.Printf("Failed to fetch user: %v", result.Error)
            continue
        }
        users = append(users, result.Result)
    }

    fmt.Printf("Successfully fetched %d users\n", len(users))
}
```

## Arquitectura y Desacoplamiento

### Ventajas del Diseño Actual

1. **Separación de Responsabilidades**:
   - `Workers`: Gestiona el número de workers y delega al dispatcher
   - `Dispatcher`: Maneja el context, cancelación y ejecución de jobs
   - Cada dispatcher tiene su propio context cancelable

2. **Context Management**:
   - El dispatcher recibe un context en su constructor
   - Crea un context cancelable derivado (`WithCancel`)
   - Todos los jobs usan ese context

3. **Simplicidad en el API**:
   - No necesitas pasar context al llamar `Dispatch()`
   - El context ya está configurado en el dispatcher
   - Workers NO es genérico - el tipo se infiere del dispatcher

4. **Workers define la concurrencia**:
   - El número de workers se configura en `NewWorkers(numWorkers, dispatcher)`
   - Workers NO tiene parámetro de tipo - se infiere automáticamente

### Flujo de Datos

```
Usuario → Crea Context
           ↓
       Crea Dispatcher[T] (con context y metrics)
           ↓
       Inyecta en Workers (con numWorkers) - tipo T inferido
           ↓
       workers.Dispatch[T](workerPool, jobs) → Dispatcher.Dispatch(numWorkers, jobs)
                                                       ↓
                                             Ejecuta jobs en pool
                                                       ↓
                                             Retorna channel de resultados
```

## Métricas Disponibles

El sistema proporciona las siguientes métricas:

- **TotalJobs**: Total de jobs enviados
- **CompletedJobs**: Jobs completados exitosamente
- **FailedJobs**: Jobs que fallaron
- **ActiveJobs**: Jobs actualmente en ejecución
- **TotalDuration**: Duración total acumulada
- **AverageDuration**: Duración promedio por job

### Ejemplo 1: Métricas Básicas

```go
package main

import (
    "context"
    "fmt"

    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/dispatcher"
    "clean-arq-layout/internal/workers/jobs"
)

func main() {
    // Crear workers
    ctx := context.Background()
    workerPool := workers.NewWorkers(5, ctx, dispatcher.NewDispatcherWithResult[string])

    // Crear y despachar jobs
    jobList := make([]jobs.Job[string], 0)
    for i := 1; i <= 100; i++ {
        jobList = append(jobList, &jobs.ExampleJob{
            ID:   i,
            Data: fmt.Sprintf("Task %d", i),
        })
    }

    resultChan := workers.Dispatch(workerPool, jobList)

    // Procesar resultados
    for range resultChan {
        // Procesar cada resultado
    }

    // Obtener métricas finales
    finalMetrics := metrics.GetMetrics()
    fmt.Printf("\n=== Métricas Finales ===\n")
    fmt.Printf("Total de jobs: %d\n", finalMetrics.TotalJobs)
    fmt.Printf("Jobs completados: %d\n", finalMetrics.CompletedJobs)
    fmt.Printf("Jobs fallidos: %d\n", finalMetrics.FailedJobs)
    fmt.Printf("Jobs activos: %d\n", finalMetrics.ActiveJobs)
    fmt.Printf("Duración total: %v\n", finalMetrics.TotalDuration)
    fmt.Printf("Duración promedio: %v\n", finalMetrics.AverageDuration)
}
```

### Ejemplo 2: Monitoreo en Tiempo Real

```go
package main

import (
    "context"
    "fmt"
    "time"

    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/dispatcher"
    "clean-arq-layout/internal/workers/jobs"
)

func main() {
    ctx := context.Background()
    workerPool := workers.NewWorkers(10, ctx, dispatcher.NewDispatcherWithResult[string])

    // Crear jobs
    jobList := make([]jobs.Job[string], 0)
    for i := 1; i <= 1000; i++ {
        jobList = append(jobList, &jobs.ExampleJob{
            ID:   i,
            Data: fmt.Sprintf("Task %d", i),
        })
    }

    // Despachar jobs
    resultChan := workers.Dispatch(workerPool, jobList)

    // Goroutine para monitorear métricas en tiempo real
    done := make(chan bool)
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                currentMetrics := workerPool.GetMetrics()
                fmt.Printf("\r[%s] Total: %d | Completados: %d | Fallidos: %d | Activos: %d",
                    time.Now().Format("15:04:05"),
                    currentMetrics.TotalJobs,
                    currentMetrics.CompletedJobs,
                    currentMetrics.FailedJobs,
                    currentMetrics.ActiveJobs,
                )
            case <-done:
                return
            }
        }
    }()

    // Procesar resultados
    for range resultChan {
        // Procesar cada resultado
    }

    done <- true
    time.Sleep(100 * time.Millisecond)

    // Métricas finales
    finalMetrics := metrics.GetMetrics()
    fmt.Printf("\n\n=== Resumen Final ===\n")
    fmt.Printf("Total procesados: %d/%d\n",
        finalMetrics.CompletedJobs + finalMetrics.FailedJobs,
        finalMetrics.TotalJobs)
    fmt.Printf("Tasa de éxito: %.2f%%\n",
        float64(finalMetrics.CompletedJobs)/float64(finalMetrics.TotalJobs)*100)
    fmt.Printf("Duración promedio: %v\n", finalMetrics.AverageDuration)
}
```

### Ejemplo 3: Métricas para Decisiones de Cancelación

```go
package main

import (
    "context"
    "fmt"
    "log"

    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/dispatcher"
    "clean-arq-layout/internal/workers/jobs"
)

func main() {
    ctx := context.Background()
    workerPool := workers.NewWorkers(5, ctx, dispatcher.NewDispatcherWithResult[string])

    // Crear jobs
    jobList := make([]jobs.Job[string], 0)
    for i := 1; i <= 100; i++ {
        jobList = append(jobList, &jobs.ExampleJob{
            ID:   i,
            Data: fmt.Sprintf("Task %d", i),
        })
    }

    resultChan := workers.Dispatch(workerPool, jobList)

    // Procesar con lógica de cancelación basada en métricas
    for result := range resultChan {
        if result.Error != nil {
            log.Printf("Error: %v", result.Error)
        }

        // Obtener métricas actuales
        current := workerPool.GetMetrics()

        // Decidir si cancelar basado en tasa de fallos
        if current.CompletedJobs+current.FailedJobs > 10 {
            failureRate := float64(current.FailedJobs) / float64(current.CompletedJobs+current.FailedJobs)

            if failureRate > 0.5 { // Más del 50% de fallos
                log.Printf("⚠️  Tasa de fallos muy alta (%.2f%%), cancelando jobs restantes", failureRate*100)
                workerPool.Cancel()
                break
            }
        }
    }

    // Reporte final
    finalMetrics := metrics.GetMetrics()
    fmt.Printf("\n=== Reporte Final ===\n")
    fmt.Printf("Jobs enviados: %d\n", finalMetrics.TotalJobs)
    fmt.Printf("Jobs completados: %d\n", finalMetrics.CompletedJobs)
    fmt.Printf("Jobs fallidos: %d\n", finalMetrics.FailedJobs)

    if finalMetrics.TotalJobs > 0 {
        processed := finalMetrics.CompletedJobs + finalMetrics.FailedJobs
        fmt.Printf("Procesados: %d (%.2f%%)\n",
            processed,
            float64(processed)/float64(finalMetrics.TotalJobs)*100)
    }
}
```

### Ejemplo 4: Métricas de Múltiples Worker Pools

```go
package main

import (
    "context"
    "fmt"

    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/dispatcher"
    "clean-arq-layout/internal/workers/jobs"
)

func main() {
    ctx := context.Background()

    // Dispatcher 1: Para procesar usuarios
    userWorkerPool := workers.NewWorkers(5, ctx, dispatcher.NewDispatcherWithResult[string])

    // Dispatcher 2: Para procesar órdenes
    orderWorkerPool := workers.NewWorkers(3, ctx, dispatcher.NewDispatcherWithResult[string])

    // Crear jobs para usuarios
    userJobs := make([]jobs.Job[string], 0)
    for i := 1; i <= 50; i++ {
        userJobs = append(userJobs, &jobs.ExampleJob{
            ID:   i,
            Data: fmt.Sprintf("User %d", i),
        })
    }

    // Crear jobs para órdenes
    orderJobs := make([]jobs.Job[string], 0)
    for i := 1; i <= 30; i++ {
        orderJobs = append(orderJobs, &jobs.ExampleJob{
            ID:   i,
            Data: fmt.Sprintf("Order %d", i),
        })
    }

    // Despachar ambos
    userResultChan := workers.Dispatch(userWorkerPool, userJobs)
    orderResultChan := workers.Dispatch(orderWorkerPool, orderJobs)

    // Procesar resultados de usuarios
    for range userResultChan {
        // Procesar
    }

    // Procesar resultados de órdenes
    for range orderResultChan {
        // Procesar
    }

    // Obtener métricas de cada worker pool
    userMetrics := userWorkerPool.GetMetrics()
    orderMetrics := orderWorkerPool.GetMetrics()

    fmt.Printf("\n=== Métricas de Usuarios ===\n")
    fmt.Printf("Total: %d, Completados: %d, Fallidos: %d\n",
        userMetrics.TotalJobs, userMetrics.CompletedJobs, userMetrics.FailedJobs)

    fmt.Printf("\n=== Métricas de Órdenes ===\n")
    fmt.Printf("Total: %d, Completados: %d, Fallidos: %d\n",
        orderMetrics.TotalJobs, orderMetrics.CompletedJobs, orderMetrics.FailedJobs)

    // Total consolidado
    fmt.Printf("\n=== Total Consolidado ===\n")
    fmt.Printf("Total de jobs: %d\n",
        userMetrics.TotalJobs + orderMetrics.TotalJobs)
    fmt.Printf("Completados: %d\n",
        userMetrics.CompletedJobs + orderMetrics.CompletedJobs)
    fmt.Printf("Fallidos: %d\n",
        userMetrics.FailedJobs + orderMetrics.FailedJobs)
}
```

## Cancelación de Jobs

La cancelación es simple, solo llama a `Cancel()`:

```go
// Crear workers
ctx := context.Background()
workerPool := workers.NewWorkers(5, ctx, dispatcher.NewDispatcherWithResult[string])

// Despachar jobs
resultChan := workers.Dispatch(workerPool, jobList)

// En el loop de resultados
for result := range resultChan {
    if someCondition {
        workerPool.Cancel()  // Cancela el context interno del dispatcher
        break
    }
}
```

## Características de Go 1.25

Esta implementación usa `sync.WaitGroup` de manera moderna:
- Utiliza el método `wg.Done()` introducido en Go 1.25 para una sintaxis más limpia

## Notas Importantes

1. **Thread-Safety**: Todas las operaciones son seguras para uso concurrente
2. **Channel Closure**: Los channels de resultados se cierran automáticamente cuando todos los jobs terminan
3. **Context Management**: El dispatcher crea y maneja su propio context cancelable
4. **Error Handling**: Los errores no detienen otros jobs a menos que explícitamente canceles
5. **Buffered Channels**: Los channels están bufferizados para mejor rendimiento
6. **Dependency Injection**: Workers recibe el dispatcher, permitiendo flexibilidad y testing
7. **No Context Pollution**: No necesitas pasar context en cada `Dispatch()`, está en el dispatcher
