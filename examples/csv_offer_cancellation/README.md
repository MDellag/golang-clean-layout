# CSV Offer Cancellation Example

Este ejemplo demuestra cómo usar el sistema de workers para procesar un archivo CSV de ofertas y cancelarlas masivamente a través de HTTP requests, recopilando los resultados en un archivo CSV de salida.

## Características

- **Procesamiento paralelo**: Utiliza múltiples workers para cancelar ofertas concurrentemente
- **Manejo de errores**: Captura y registra errores de cada cancelación
- **Resultados detallados**: Genera CSV de salida con estado, duración y timestamp de cada operación
- **Reintentos automáticos**: Incluye lógica de reintentos con backoff exponencial
- **Monitoreo**: Métricas y logs detallados del progreso

## Uso

### Ejecución Básica

```bash
go run main.go "https://api.example.com/offers/cancel" "sample_input.csv" "output/results.csv"
```

### Con Número Personalizado de Workers

```bash
go run main.go "https://api.example.com/offers/cancel" "sample_input.csv" "output/results.csv" 15
```

## Formato de Archivos

### CSV de Entrada

El archivo debe contener una columna `offer_id`:

```csv
offer_id,created_at,amount
OFFER001,2024-01-15T10:30:00Z,100.50
OFFER002,2024-01-16T14:25:30Z,250.00
OFFER003,2024-01-17T09:15:45Z,75.25
```

### CSV de Salida

El resultado incluye información detallada de cada cancelación:

```csv
offer_id,row,status,error_message,duration_ms,timestamp
OFFER001,1,SUCCESS,,1245.67,2024-01-18T10:30:45Z
OFFER002,2,ERROR,HTTP error 404,890.23,2024-01-18T10:30:46Z
OFFER003,3,SUCCESS,,1100.45,2024-01-18T10:30:47Z
```

**Campos del CSV de salida:**
- `offer_id`: ID de la oferta procesada
- `row`: Número de fila del CSV original
- `status`: SUCCESS, ERROR, o NOT_PROCESSED
- `error_message`: Descripción del error (si aplica)
- `duration_ms`: Tiempo de procesamiento en milisegundos
- `timestamp`: Momento de completación en formato RFC3339

## API del Servicio

El servicio debe aceptar requests POST con el siguiente formato:

**Request:**
```json
{
    "offer_id": "OFFER001",
    "reason": "batch_cancellation"
}
```

**Response exitosa (2xx):**
```json
{
    "offer_id": "OFFER001",
    "status": "cancelled",
    "message": "Offer cancelled successfully",
    "timestamp": "2024-01-18T10:30:45Z"
}
```

**Response de error (4xx/5xx):**
```json
{
    "offer_id": "OFFER001",
    "status": "error",
    "message": "Offer not found",
    "timestamp": "2024-01-18T10:30:45Z"
}
```

## Configuración

### Variables de Entorno

```bash
# Opcional: configurar timeout HTTP
export HTTP_TIMEOUT=30s

# Opcional: configurar número máximo de reintentos
export MAX_RETRIES=3

# Opcional: configurar timeout de procesamiento total
export PROCESSING_TIMEOUT=10m
```

### Personalización del HTTP Client

Para configuraciones más avanzadas, puedes modificar el cliente HTTP en `OfferCancelJob`:

```go
job := jobs.NewOfferCancelJob(id, offerID, serviceURL, reason, responseChannel)

// Configurar cliente personalizado
customClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:       100,
        MaxIdleConnsPerHost: 20,
        IdleConnTimeout:    90 * time.Second,
    },
}
job.SetHTTPClient(customClient)
```

## Monitoreo y Logs

El sistema genera logs detallados durante la ejecución:

```
2024/01/18 10:30:00 Starting CSV offer cancellation processor
2024/01/18 10:30:00 Service URL: https://api.example.com/offers/cancel
2024/01/18 10:30:00 Processing 10 offer cancellations from sample_input.csv
2024/01/18 10:30:00 Worker pool started with 10 workers
2024/01/18 10:30:01 Worker 1 processing job: offer-cancel-OFFER001
2024/01/18 10:30:01 Worker 2 processing job: offer-cancel-OFFER002
...
2024/01/18 10:30:45 Waiting for all cancellation jobs to complete...
2024/01/18 10:30:50 Processing complete. Results written to output/results.csv
2024/01/18 10:30:50 === PROCESSING SUMMARY ===
2024/01/18 10:30:50 Total processed: 10
2024/01/18 10:30:50 Successful: 8
2024/01/18 10:30:50 Failed: 2
2024/01/18 10:30:50 Success rate: 80.00%
2024/01/18 10:30:50 Average duration: 1.2s
```

## Manejo de Errores Comunes

### Error de Conexión
```
ERROR: HTTP request failed: dial tcp: connection refused
```
**Solución**: Verificar que el servicio esté disponible y la URL sea correcta.

### Error de Timeout
```
ERROR: context deadline exceeded
```
**Solución**: Aumentar el timeout del HTTP client o del contexto.

### Error de Formato CSV
```
ERROR: offer_id column not found in CSV
```
**Solución**: Verificar que el CSV tenga la columna `offer_id` en el header.

### Error de Permisos
```
ERROR: failed to create output directory: permission denied
```
**Solución**: Verificar permisos de escritura en el directorio de salida.

## Extensiones

### Agregar Autenticación

```go
// En offer_cancel_job.go, agregar headers de autorización
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("X-API-Key", apiKey)
```

### Procesamiento por Lotes

Para procesar múltiples ofertas por request:

```go
type BatchCancelJob struct {
    offers []string
    // ... otros campos
}

func (j *BatchCancelJob) Execute(ctx context.Context) error {
    request := BatchCancelRequest{
        OfferIDs: j.offers,
        Reason:   j.reason,
    }
    // ... lógica de batch
}
```

### Validación Previa

```go
type OfferValidationJob struct {
    offerID string
    // ... campos de validación
}

// Ejecutar validación antes de cancelación
func (j *OfferValidationJob) Execute(ctx context.Context) error {
    // Validar que la oferta existe y puede ser cancelada
    return validateOffer(ctx, j.offerID)
}
```

Este ejemplo proporciona una base sólida para el procesamiento masivo de cancelaciones de ofertas, con manejo robusto de errores y recopilación completa de resultados.