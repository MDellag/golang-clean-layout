package main

import (
	"clean-arq-layout/internal/app"
	"log"
	"os"
)

func main() {
	// Leer ambiente desde variable de entorno
	// Si APP_ENV no está definida, se usará "local" por defecto
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
		log.Printf("APP_ENV no definida, usando ambiente: %s", env)
	} else {
		log.Printf("Iniciando aplicación en ambiente: %s", env)
	}

	// Iniciar aplicación con el ambiente especificado
	app.Start(env)
}
