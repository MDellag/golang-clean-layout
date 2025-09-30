package main

import (
	"context"
	"log"
	"os"

	"clean-arq-layout/internal/infrastructure/http/clients"
	"clean-arq-layout/internal/workers"
)

func main() {
	// Configurar logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Verificar argumentos
	if len(os.Args) < 4 {
		log.Fatal("Usage: go run main.go <service_url> <input_csv> <output_csv> [workers]")
	}

	serviceURL := os.Args[1]
	inputCSV := os.Args[2]
	outputCSV := os.Args[3]
	
	workers := 10
	if len(os.Args) > 4 {
		// Parsear número de workers si se proporciona
		// Para simplicidad, usando valor por defecto
	}

	log.Printf("Starting CSV offer cancellation processor")
	log.Printf("Service URL: %s", serviceURL)
	log.Printf("Input CSV: %s", inputCSV)
	log.Printf("Output CSV: %s", outputCSV)
	log.Printf("Workers: %d", workers)

	// Crear cliente del servicio de precios
	priceService := clients.NewPriceServiceHTTPClient(serviceURL, "batch_cancellation")

	// Crear procesador CSV con el cliente inyectado
	processor := worker.NewCSVProcessor(
		priceService,
		inputCSV,
		outputCSV,
		workers,
	)

	// Procesar CSV
	ctx := context.Background()
	if err := processor.ProcessCSV(ctx); err != nil {
		log.Fatalf("CSV processing failed: %v", err)
	}

	log.Println("CSV processing completed successfully!")
}