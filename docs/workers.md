# Workers System

Sistema de workers asíncronos para el procesamiento de tareas en background, implementado utilizando las mejores prácticas de Go y aprovechando las nuevas funcionalidades de Go 1.25.

## Arquitectura

El sistema está compuesto por los siguientes componentes:

- **Job Interface**: Define el contrato para tareas que pueden ser ejecutadas
- **Worker**: Ejecuta trabajos individualmente con métricas y manejo de errores
- **Pool**: Administra un grupo de workers para procesar trabajos concurrentemente
- **Dispatcher**: Coordina y distribuye trabajos entre workers disponibles

## Mejoras con Go 1.25

Esta implementación utiliza la nueva funcionalidad `sync.WaitGroup.Go()` de Go 1.25, que simplifica el patrón común de crear goroutines rastreadas por WaitGroup:

**Antes (Go < 1.25):**
```go
wg.Add(1)
go func() {
    defer wg.Done()
    worker.run()
}()
```

**Ahora (Go 1.25+):**
```go
wg.Go(worker.run)
```

## Implementación de Jobs

### Job Interface

Todos los jobs deben implementar la interfaz `Job`:

```go
type Job interface {
    Execute(ctx context.Context) error
    Name() string
    Priority() int
}
```

### Ejemplo Básico: SimpleJob

```go
package jobs

import (
    "context"
    "fmt"
    "log"
    "time"
)

type SimpleJob struct {
    ID    string
    Data  string
    Delay time.Duration
}

func NewSimpleJob(id string, data string, delay time.Duration) *SimpleJob {
    return &SimpleJob{
        ID:    id,
        Data:  data,
        Delay: delay,
    }
}

func (j *SimpleJob) Execute(ctx context.Context) error {
    log.Printf("Starting job %s with data: %s", j.ID, j.Data)

    timer := time.NewTimer(j.Delay)
    defer timer.Stop()

    select {
    case <-timer.C:
        result := fmt.Sprintf("Processed: %s", j.Data)
        log.Printf("Completed job %s with result: %s", j.ID, result)
        return nil

    case <-ctx.Done():
        log.Printf("Job %s was canceled: %v", j.ID, ctx.Err())
        return ctx.Err()
    }
}

func (j *SimpleJob) Name() string {
    return fmt.Sprintf("simple-job-%s", j.ID)
}

func (j *SimpleJob) Priority() int {
    return 1
}
```

### Ejemplo Avanzado: EmailJob con Reintentos

```go
package jobs

import (
    "context"
    "fmt"
    "log"
    "time"
)

type EmailJob struct {
    ID           string
    Recipient    string
    Subject      string
    Body         string
    MaxRetries   int
    currentRetry int
}

func NewEmailJob(id, recipient, subject, body string) *EmailJob {
    return &EmailJob{
        ID:         id,
        Recipient:  recipient,
        Subject:    subject,
        Body:       body,
        MaxRetries: 3,
    }
}

func (j *EmailJob) Execute(ctx context.Context) error {
    log.Printf("Starting email job %s to %s", j.ID, j.Recipient)

    for j.currentRetry <= j.MaxRetries {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        if err := j.sendEmail(ctx); err != nil {
            j.currentRetry++
            if j.currentRetry <= j.MaxRetries {
                log.Printf("Email job %s failed (attempt %d/%d): %v, retrying...", 
                    j.ID, j.currentRetry, j.MaxRetries, err)
                
                // Backoff exponencial
                backoff := time.Duration(j.currentRetry*j.currentRetry) * time.Second
                timer := time.NewTimer(backoff)
                
                select {
                case <-timer.C:
                    continue
                case <-ctx.Done():
                    timer.Stop()
                    return ctx.Err()
                }
            } else {
                return fmt.Errorf("email job failed after %d attempts: %w", j.MaxRetries, err)
            }
        } else {
            log.Printf("Email job %s completed successfully", j.ID)
            return nil
        }
    }

    return fmt.Errorf("email job %s exhausted all retry attempts", j.ID)
}

func (j *EmailJob) sendEmail(ctx context.Context) error {
    // Lógica de envío de email
    select {
    case <-time.After(500 * time.Millisecond):
        // Simular 70% de éxito
        if time.Now().UnixNano()%10 < 7 {
            return nil
        }
        return fmt.Errorf("temporary email service error")
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (j *EmailJob) Name() string {
    return fmt.Sprintf("email-job-%s", j.ID)
}

func (j *EmailJob) Priority() int {
    return 2 // Mayor prioridad que trabajos simples
}
```

## Uso del Sistema

### Configuración Básica

```go
package main

import (
    "context"
    "log"
    "time"
    
    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/jobs"
)

func main() {
    ctx := context.Background()
    
    // Crear dispatcher con 5 workers y cola de 100 trabajos
    dispatcher := worker.NewDispatcher(ctx, 5, 100)
    
    // Iniciar el dispatcher
    if err := dispatcher.Start(); err != nil {
        log.Fatalf("Failed to start dispatcher: %v", err)
    }
    defer dispatcher.Stop()
    
    // Encolar trabajos
    job1 := jobs.NewSimpleJob("1", "Hello World", 2*time.Second)
    job2 := jobs.NewEmailJob("email-001", "user@example.com", "Test", "Body")
    
    if err := dispatcher.EnqueueJob(job1); err != nil {
        log.Printf("Failed to enqueue job1: %v", err)
    }
    
    if err := dispatcher.EnqueueJob(job2); err != nil {
        log.Printf("Failed to enqueue job2: %v", err)
    }
    
    // Esperar un poco para que se procesen los trabajos
    time.Sleep(10 * time.Second)
}
```

### Configuración Avanzada con Múltiples Tipos de Jobs

```go
package main

import (
    "context"
    "log"
    "time"
    
    "clean-arq-layout/internal/workers"
    "clean-arq-layout/internal/workers/jobs"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Crear dispatcher
    dispatcher := worker.NewDispatcher(ctx, 10, 200)
    
    if err := dispatcher.Start(); err != nil {
        log.Fatalf("Failed to start dispatcher: %v", err)
    }
    defer dispatcher.Stop()
    
    // Procesar múltiples trabajos
    go func() {
        for i := 0; i < 20; i++ {
            // Alternar entre tipos de trabajos
            if i%2 == 0 {
                job := jobs.NewSimpleJob(
                    fmt.Sprintf("simple-%d", i), 
                    fmt.Sprintf("Processing item %d", i), 
                    time.Duration(i)*100*time.Millisecond,
                )
                dispatcher.EnqueueJob(job)
            } else {
                job := jobs.NewEmailJob(
                    fmt.Sprintf("email-%d", i),
                    fmt.Sprintf("user%d@example.com", i),
                    "Notification",
                    fmt.Sprintf("You have received notification #%d", i),
                )
                dispatcher.EnqueueJob(job)
            }
            
            time.Sleep(100 * time.Millisecond)
        }
    }()
    
    // Monitorear estadísticas
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            stats := dispatcher.Stats()
            log.Printf("Stats: %+v", stats)
            
        case <-ctx.Done():
            log.Println("Shutting down...")
            return
        }
    }
}
```

## Monitoreo y Métricas

### Estadísticas del Dispatcher

```go
stats := dispatcher.Stats()
log.Printf("Workers: %d, Pending jobs: %d, Running: %t", 
    stats["workers"], stats["pending_jobs"], stats["is_running"])
```

### Métricas de Workers Individuales

Los workers mantienen métricas automáticamente:

- `JobsProcessed`: Número de trabajos completados exitosamente
- `Errors`: Número de trabajos que fallaron
- `LastJobTime`: Duración del último trabajo procesado

```go
// Acceder a métricas (esto requeriría exponerlas a través del pool)
for _, worker := range pool.Workers() {
    metrics := worker.Metrics()
    log.Printf("Worker %d: Processed=%d, Errors=%d, LastJobTime=%v",
        worker.ID, metrics.JobsProcessed, metrics.Errors, metrics.LastJobTime)
}
```

## Características Clave

### 1. Graceful Shutdown
- El sistema maneja correctamente la cancelación de contexto
- Los workers terminan de procesar trabajos actuales antes de cerrarse
- Utiliza `sync.WaitGroup` para esperar que todos los workers terminen

### 2. Timeout por Job
- Cada job ejecuta con un timeout de 5 minutos por defecto
- Contexto cancelable para interrumpir trabajos de larga duración

### 3. Pool de Workers Escalable
- Número configurable de workers
- Cola de trabajos con tamaño configurable
- Distribución automática de carga

### 4. Patrones de Retry
- Implementación de reintentos a nivel de job
- Backoff exponencial para espaciar reintentos
- Límite configurable de intentos

## Mejores Prácticas

### 1. Diseño de Jobs
- Mantener jobs idempotentes cuando sea posible
- Implementar timeout adecuado en operaciones de larga duración
- Manejar cancelación de contexto correctamente
- Evitar dependencias externas pesadas en constructores

### 2. Configuración del Pool
- Ajustar número de workers según CPU y tipo de workload
- Configurar tamaño de cola basado en patrones de tráfico
- Monitorear métricas para optimizar configuración

### 3. Manejo de Errores
- Logs detallados para debugging
- Diferenciación entre errores temporales y permanentes
- Implementación de reintentos solo para errores recuperables

### 4. Testing
- Crear jobs de prueba con delays controlables
- Usar contextos con timeout para tests
- Validar comportamiento de cancelación y shutdown

## Jobs con Canales de Respuesta

Para casos donde necesitas recopilar resultados de cada job (como procesamiento de CSV con estado de éxito/fallo), el sistema soporta jobs con canales de respuesta.

### Interfaces Extendidas

```go
// JobResult representa el resultado de la ejecución de un job
type JobResult struct {
    JobID     string
    JobName   string
    Success   bool
    Error     error
    Data      interface{}
    Duration  time.Duration
    Timestamp time.Time
}

// JobWithResponse extiende Job para incluir canal de respuesta
type JobWithResponse interface {
    Job
    ResponseChannel() chan<- JobResult
    ID() string
}
```

### Ejemplo: Job de Cancelación de Ofertas con Cliente Inyectado

```go
// Interfaz del cliente de servicio
type PriceServiceClient interface {
    Cancel(ctx context.Context, offerID string) error
}

// Job simplificado con cliente inyectado
type OfferCancelJob struct {
    id              string
    offerID         string
    priceService    interfaces.PriceServiceClient
    responseChannel chan<- worker.JobResult
    maxRetries      int
}

func NewOfferCancelJob(id, offerID string, priceService interfaces.PriceServiceClient, responseChannel chan<- worker.JobResult) *OfferCancelJob {
    return &OfferCancelJob{
        id:              id,
        offerID:         offerID,
        priceService:    priceService,
        responseChannel: responseChannel,
        maxRetries:      3,
    }
}

func (j *OfferCancelJob) Execute(ctx context.Context) error {
    for j.currentRetry <= j.maxRetries {
        // Llamar al método Cancel del cliente de servicio
        err := j.priceService.Cancel(ctx, j.offerID)
        if err != nil {
            j.currentRetry++
            if j.currentRetry <= j.maxRetries {
                // Backoff exponencial antes de reintentar
                backoff := time.Duration(j.currentRetry*j.currentRetry) * time.Second
                time.Sleep(backoff)
                continue
            }
            return fmt.Errorf("offer cancellation failed after %d attempts: %w", j.maxRetries, err)
        }
        return nil // Éxito
    }
    return fmt.Errorf("offer cancellation exhausted all retry attempts for offer %s", j.offerID)
}

func (j *OfferCancelJob) ID() string { return j.id }
func (j *OfferCancelJob) Name() string { return fmt.Sprintf("offer-cancel-%s", j.offerID) }
func (j *OfferCancelJob) Priority() int { return 3 }
func (j *OfferCancelJob) ResponseChannel() chan<- worker.JobResult { return j.responseChannel }
```

### Agregador de Respuestas

```go
// Crear agregador para recopilar resultados
aggregator := worker.NewResponseAggregator(1000)
aggregator.Start()
defer aggregator.Stop()

// Usar el canal del agregador para los jobs
responseChannel := aggregator.GetResultsChannel()

// Crear cliente de servicio
priceService := clients.NewPriceServiceHTTPClient("https://api.example.com/cancel", "batch_cancellation")

// Crear jobs con canal de respuesta
for i, offerID := range offerIDs {
    job := NewOfferCancelJob(
        fmt.Sprintf("cancel-%d", i),
        offerID,
        priceService,
        responseChannel,
    )
    dispatcher.EnqueueJob(job)
}

// Esperar resultados
results := aggregator.WaitForResults(len(offerIDs), 5*time.Minute)

// Procesar resultados
for _, result := range results {
    if result.Success {
        log.Printf("Offer %s canceled successfully", result.JobID)
    } else {
        log.Printf("Failed to cancel offer %s: %v", result.JobID, result.Error)
    }
}
```

## Procesamiento de CSV Completo

### Caso de Uso: Cancelación Masiva de Ofertas

```go
package main

import (
    "context"
    "log"
    
    "clean-arq-layout/internal/workers"
)

func main() {
    // Crear cliente del servicio de precios
    priceService := clients.NewPriceServiceHTTPClient(
        "https://api.example.com/offers/cancel", 
        "batch_cancellation",
    )

    // Crear procesador CSV con el cliente inyectado
    processor := worker.NewCSVProcessor(
        priceService,                           // Cliente de servicio
        "input/offers_to_cancel.csv",          // Archivo de entrada
        "output/cancellation_results.csv",     // Archivo de salida
        10,                                     // Número de workers
    )

    ctx := context.Background()
    if err := processor.ProcessCSV(ctx); err != nil {
        log.Fatalf("CSV processing failed: %v", err)
    }
}
```

### Formato de CSV de Entrada

```csv
offer_id,created_at,amount
OFFER001,2024-01-15,100.50
OFFER002,2024-01-16,250.00
OFFER003,2024-01-17,75.25
```

### Formato de CSV de Salida

```csv
offer_id,row,status,error_message,duration_ms,timestamp
OFFER001,1,SUCCESS,,1245.67,2024-01-18T10:30:45Z
OFFER002,2,ERROR,HTTP error 404,890.23,2024-01-18T10:30:46Z
OFFER003,3,SUCCESS,,1100.45,2024-01-18T10:30:47Z
```

## Patrones de Uso Comunes

### Procesamiento de Archivos con Resultado
```go
type FileProcessJob struct {
    id              string
    filePath        string
    outputPath      string
    responseChannel chan<- worker.JobResult
}

func (j *FileProcessJob) Execute(ctx context.Context) error {
    // Procesar archivo y retornar métricas en el resultado
    metrics, err := processFile(ctx, j.filePath, j.outputPath)
    
    // Los datos se enviarán automáticamente al canal de respuesta
    // junto con el estado de éxito/error
    
    return err
}
```

### Llamadas a APIs con Respuesta Estructurada
```go
type APICallJob struct {
    id              string
    url             string
    payload         []byte
    responseChannel chan<- worker.JobResult
}

func (j *APICallJob) Execute(ctx context.Context) error {
    response, err := makeAPICall(ctx, j.url, j.payload)
    
    // El resultado contendrá la respuesta en el campo Data
    // si el job implementa JobWithResponse
    
    return err
}
```

### Validación de Datos en Lotes
```go
type DataValidationJob struct {
    id              string
    records         []Record
    responseChannel chan<- worker.JobResult
}

func (j *DataValidationJob) Execute(ctx context.Context) error {
    validationResults := validateRecords(j.records)
    
    // Los resultados de validación se pueden incluir en JobResult.Data
    // para análisis posterior
    
    return nil
}
```

## Monitoreo Avanzado

### Métricas en Tiempo Real

```go
// Crear canal para monitorear resultados en tiempo real
resultsChan := make(chan worker.JobResult, 100)

go func() {
    for result := range resultsChan {
        // Enviar métricas a sistema de monitoreo
        if result.Success {
            metrics.IncrementCounter("jobs.success")
        } else {
            metrics.IncrementCounter("jobs.failed")
        }
        metrics.RecordDuration("jobs.duration", result.Duration)
    }
}()

// Usar el canal para jobs
jobs := createJobsWithResponseChannel(resultsChan)
```

### Dashboard de Progreso

```go
func monitorProgress(aggregator *worker.ResponseAggregator, totalJobs int) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            results := aggregator.GetResults()
            completed := len(results)
            successRate := calculateSuccessRate(results)
            
            log.Printf("Progress: %d/%d (%.1f%%) - Success Rate: %.1f%%",
                completed, totalJobs, 
                float64(completed)/float64(totalJobs)*100,
                successRate)
                
            if completed >= totalJobs {
                return
            }
        }
    }
}
```

Este sistema de workers con canales de respuesta proporciona una solución completa y escalable para el procesamiento asíncrono con seguimiento detallado de resultados, aprovechando las mejoras de Go 1.25 para un código más limpio y mantenible.