# Workers

Sistema de procesamiento de tareas en background usando goroutines y generics de Go.

## Uso rápido

```go
dispatcher := worker.NewGenericDispatcher[MiResultado](workers, bufferSize)
defer dispatcher.Stop()

results, err := dispatcher.DispatchAndWait(ctx, jobs)
```

## GenericDispatcher (recomendado)

El `GenericDispatcher` es la API principal. Acepta un slice de jobs y devuelve resultados tipados.

### 1. Implementar un Job

```go
type PrecioJob struct {
    productoID string
}

func (j *PrecioJob) Execute(ctx context.Context) (Precio, error) {
    return buscarPrecio(ctx, j.productoID)
}

func (j *PrecioJob) Name() string     { return "precio-" + j.productoID }
func (j *PrecioJob) Priority() int    { return 1 }
```

### 2. Dispatch y esperar todos los resultados

```go
dispatcher := worker.NewGenericDispatcher[Precio](5, 100)
defer dispatcher.Stop()

jobs := []types.GenericJob[Precio]{
    &PrecioJob{"prod-1"},
    &PrecioJob{"prod-2"},
    &PrecioJob{"prod-3"},
}

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

results, err := dispatcher.DispatchAndWait(ctx, jobs)
for _, r := range results {
    if r.Success {
        fmt.Printf("✓ %s: %.2f\n", r.JobName, r.Result.Valor)
    } else {
        fmt.Printf("✗ %s: %v\n", r.JobName, r.Error)
    }
}
```

### 3. Consumir resultados a medida que llegan

Útil cuando querés procesar cada resultado sin esperar al resto.

```go
resultChan, err := dispatcher.Dispatch(jobs)
if err != nil {
    log.Fatal(err)
}

for r := range resultChan {
    if r.Success {
        procesar(r.Result)
    }
}
```

## Interfaz GenericJob

```go
type GenericJob[T any] interface {
    Execute(ctx context.Context) (T, error)
    Name() string
    Priority() int
}
```

## Dispatcher clásico (sin generics)

Para jobs sin valor de retorno (fire-and-forget).

```go
dispatcher := worker.NewDispatcher(ctx, 5, 100)
dispatcher.Start()
defer dispatcher.Stop()

dispatcher.EnqueueJob(miJob)
```

### Job clásico

```go
type NotificacionJob struct {
    userID string
}

func (j *NotificacionJob) Execute(ctx context.Context) error {
    return enviarNotificacion(ctx, j.userID)
}

func (j *NotificacionJob) Name() string  { return "notif-" + j.userID }
func (j *NotificacionJob) Priority() int { return 1 }
```

## Parámetros del dispatcher

| Parámetro    | Descripción                                      |
|--------------|--------------------------------------------------|
| `maxWorkers` | Goroutines paralelas para procesar jobs          |
| `bufferSize` | Capacidad del canal de resultados / cola de jobs |

- **Más workers**: mejor para workloads I/O intensivos (HTTP, DB)
- **Más buffer**: evita bloqueos cuando los jobs se producen más rápido de lo que se consumen

## Cancelación y shutdown

Cada job recibe un `context.Context`. El dispatcher lo cancela cuando se llama a `Stop()` o cuando el contexto padre expira.

```go
func (j *MiJob) Execute(ctx context.Context) (MiResultado, error) {
    select {
    case <-ctx.Done():
        return MiResultado{}, ctx.Err()
    default:
    }
    // ... lógica del job
}
```
