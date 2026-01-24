package main

import (
	"fmt"
	"os"

	"clean-arq-layout/config"
)

// Ejemplo que demuestra el merge parcial de configuración
// Este ejemplo muestra cómo los valores se heredan de config.yaml
// y solo se sobrescriben los específicos de cada ambiente

func main() {
	fmt.Println("=== DEMOSTRACIÓN DE MERGE PARCIAL ===\n")

	// Cargar las 3 configuraciones
	environments := []string{"local", "test", "prod"}

	for _, env := range environments {
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("AMBIENTE: %s\n", env)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		// Para prod, necesitamos las ENV vars
		if env == "prod" {
			os.Setenv("DB_HOST", "prod-db.example.com")
			os.Setenv("DB_USERNAME", "prod_user")
			os.Setenv("DB_PASSWORD", "prod_secret")
			os.Setenv("DB_NAME", "production_db")
			os.Setenv("MONGO_URL", "mongodb://prod:secret@mongo:27017")
			os.Setenv("MONGO_DB", "prod_db")
		}

		// Configurar APP_ENV para este ambiente
		os.Setenv("APP_ENV", env)
		cfg := config.Load()

		printConfig(env, cfg)
		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("RESUMEN DEL MERGE PARCIAL")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Println("📝 config.yaml contiene:")
	fmt.Println("   - Todos los valores base/defaults")
	fmt.Println("   - ~30 líneas de configuración completa\n")
	fmt.Println("📝 config_local.yaml contiene:")
	fmt.Println("   - SOLO 2 overrides (log_level, server.host)")
	fmt.Println("   - ~3 líneas\n")
	fmt.Println("📝 config_test.yaml contiene:")
	fmt.Println("   - SOLO ~8 overrides necesarios para tests")
	fmt.Println("   - ~15 líneas\n")
	fmt.Println("📝 config_prod.yaml contiene:")
	fmt.Println("   - SOLO ~10 overrides para producción")
	fmt.Println("   - ~15 líneas\n")
	fmt.Println("✨ Resultado:")
	fmt.Println("   - Sin duplicación innecesaria")
	fmt.Println("   - Fácil de mantener")
	fmt.Println("   - Claro qué es diferente en cada ambiente")
}

func printConfig(env string, cfg *config.Config) {
	// Mostrar qué viene de config.yaml vs qué viene de config_{env}.yaml
	fmt.Printf("AppName:  %s", cfg.AppName)
	printSource(env, "app_name")

	fmt.Printf("LogLevel: %s", cfg.LogLevel)
	printSource(env, "log_level")

	fmt.Printf("\nServer:\n")
	fmt.Printf("  Host:   %s", cfg.Server.Host)
	printSource(env, "server.host")

	fmt.Printf("  Port:   %d", cfg.Server.Port)
	printSource(env, "server.port")

	fmt.Printf("\nDatabase:\n")
	fmt.Printf("  Host:   %s", cfg.Database.Host)
	printSource(env, "database.host")

	fmt.Printf("  Port:   %d", cfg.Database.Port)
	printSource(env, "database.port")

	fmt.Printf("  User:   %s", cfg.Database.Username)
	printSource(env, "database.username")

	fmt.Printf("  Name:   %s", cfg.Database.Name)
	printSource(env, "database.name")

	fmt.Printf("\nSwagger:\n")
	fmt.Printf("  Host:   %s", cfg.Swagger.HostName)
	printSource(env, "swagger.hostname")

	fmt.Printf("  Enabled: %v", cfg.Swagger.Enabled)
	printSource(env, "swagger.enabled")

	fmt.Printf("\nMongo:\n")
	fmt.Printf("  DB:     %s", cfg.Mongo.DB)
	printSource(env, "mongo.db")
}

func printSource(env string, field string) {
	// Indicar de dónde viene cada valor
	overrides := map[string]map[string]bool{
		"local": {
			"log_level":   true,
			"server.host": true,
		},
		"test": {
			"log_level":         true,
			"server.port":       true,
			"database.username": true,
			"database.name":     true,
			"swagger.enabled":   true,
			"mongo.db":          true,
		},
		"prod": {
			"log_level":         true,
			"server.port":       true,
			"database.host":     true,
			"database.username": true,
			"database.name":     true,
			"swagger.enabled":   true,
			"mongo.db":          true,
		},
	}

	if overrides[env][field] {
		fmt.Printf("  ✅ (de config_%s.yaml)\n", env)
	} else {
		fmt.Printf("  ← (heredado de config.yaml)\n")
	}
}
