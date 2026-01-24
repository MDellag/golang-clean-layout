package main

import (
	"fmt"
	"log"
	"os"

	"clean-arq-layout/config"
)

// Este es un ejemplo de cómo usar el sistema de configuración
// Ejecutar con: go run examples/config_usage.go
//
// Prueba diferentes ambientes:
// - go run examples/config_usage.go
// - APP_ENV=test go run examples/config_usage.go
// - APP_ENV=prod go run examples/config_usage.go
//
// Prueba sobrescribir con variables de entorno:
// - DB_HOST=192.168.1.100 go run examples/config_usage.go
// - SERVER_PORT=9000 go run examples/config_usage.go

func main() {
	// Obtener ambiente desde variable de entorno
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	fmt.Printf("=== Cargando configuración para ambiente: %s ===\n\n", env)

	// Cargar configuración
	cfg, err := config.Load(env)
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	// Mostrar valores cargados
	fmt.Printf("Aplicación:\n")
	fmt.Printf("  Nombre: %s\n", cfg.AppName)
	fmt.Printf("  Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("\n")

	fmt.Printf("Servidor HTTP:\n")
	fmt.Printf("  Host: %s\n", cfg.Server.Host)
	fmt.Printf("  Port: %d\n", cfg.Server.Port)
	fmt.Printf("  URL completa: http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("\n")

	fmt.Printf("Base de Datos PostgreSQL:\n")
	fmt.Printf("  Host: %s\n", cfg.Database.Host)
	fmt.Printf("  Port: %d\n", cfg.Database.Port)
	fmt.Printf("  Username: %s\n", cfg.Database.Username)
	fmt.Printf("  Password: %s\n", maskPassword(cfg.Database.Password))
	fmt.Printf("  Database: %s\n", cfg.Database.Name)
	fmt.Printf("  Connection: postgresql://%s:%s@%s:%d/%s\n",
		cfg.Database.Username,
		maskPassword(cfg.Database.Password),
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name)
	fmt.Printf("\n")

	fmt.Printf("Swagger:\n")
	fmt.Printf("  Hostname: %s\n", cfg.Swagger.HostName)
	fmt.Printf("  Enabled: %v\n", cfg.Swagger.Enabled)
	fmt.Printf("\n")

	fmt.Printf("MongoDB:\n")
	fmt.Printf("  URL: %s\n", maskMongoURL(cfg.Mongo.URL))
	fmt.Printf("  Database: %s\n", cfg.Mongo.DB)
	fmt.Printf("\n")

	fmt.Println("=== Configuración cargada exitosamente ===")
}

// maskPassword enmascara la contraseña para mostrarla de forma segura
func maskPassword(password string) string {
	if len(password) == 0 {
		return "(vacía)"
	}
	if len(password) <= 3 {
		return "***"
	}
	return password[:2] + "***" + password[len(password)-1:]
}

// maskMongoURL enmascara la contraseña en la URL de MongoDB
func maskMongoURL(url string) string {
	// Buscar el patrón mongodb://user:password@
	// Esta es una implementación simple, en producción usar una librería más robusta
	if len(url) == 0 {
		return "(vacía)"
	}

	// Buscar ':' después de '//' y '@'
	startIdx := -1
	endIdx := -1

	for i := 0; i < len(url)-1; i++ {
		if url[i] == '/' && url[i+1] == '/' {
			// Encontrar el siguiente ':'
			for j := i + 2; j < len(url); j++ {
				if url[j] == ':' {
					startIdx = j + 1
					break
				}
			}
		}
		if url[i] == '@' && startIdx != -1 {
			endIdx = i
			break
		}
	}

	if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
		return url[:startIdx] + "***" + url[endIdx:]
	}

	return url
}
