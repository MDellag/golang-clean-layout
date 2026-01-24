package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config estructura principal de configuración
type Config struct {
	AppName  string   `mapstructure:"app_name"`
	LogLevel string   `mapstructure:"log_level"`
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
	Swagger  Swagger  `mapstructure:"swagger"`
	Mongo    Mongo    `mapstructure:"mongo"`
}

// Server configuración del servidor HTTP
type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Database configuración de base de datos relacional
type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// Swagger configuración de Swagger/OpenAPI
type Swagger struct {
	HostName string `mapstructure:"hostname"`
	Enabled  bool   `mapstructure:"enabled"`
}

// Mongo configuración de MongoDB
type Mongo struct {
	URL string `mapstructure:"url"`
	DB  string `mapstructure:"db"`
}

// Load carga la configuración siguiendo este orden de prioridad:
// 1. Variables de entorno (máxima prioridad)
// 2. config_{env}.yaml (específico del ambiente)
// 3. config.yaml (base común, mínima prioridad)
//
// El parámetro env determina qué archivo específico cargar (local, test, prod).
// Si env está vacío, se usa "local" por defecto.
func Load(env string) (*Config, error) {
	// Usar "local" como ambiente por defecto
	if env == "" {
		env = "local"
	}

	v := viper.New()

	// Configurar para leer variables de entorno
	v.SetEnvPrefix("") // Sin prefijo, lee todas las variables de entorno
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Configurar directorio de configuración
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// 1. Cargar config.yaml (base común)
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error leyendo config.yaml: %w", err)
	}

	// 2. Hacer merge con config_{env}.yaml (específico del ambiente)
	v.SetConfigName(fmt.Sprintf("config_%s", env))

	if err := v.MergeInConfig(); err != nil {
		// Si el archivo específico del ambiente no existe, continuar
		// (no es crítico, podemos usar solo el base)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error leyendo config_%s.yaml: %w", env, err)
		}
	}

	// Expandir variables de entorno en todos los valores
	expandEnvVars(v)

	// Deserializar en la estructura Config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error deserializando configuración: %w", err)
	}

	return &cfg, nil
}

// expandEnvVars expande las variables de entorno en todos los valores de configuración
// Soporta sintaxis ${VAR_NAME} y ${VAR_NAME:default}
func expandEnvVars(v *viper.Viper) {
	for _, key := range v.AllKeys() {
		value := v.GetString(key)
		if value != "" && strings.Contains(value, "${") {
			// Primero expandir con defaults, luego con os.ExpandEnv para variables sin default
			expanded := expandWithDefaults(value)
			v.Set(key, expanded)
		}
	}
}

// expandWithDefaults maneja la sintaxis ${VAR_NAME:default}
// También maneja ${VAR_NAME} usando os.ExpandEnv si no tiene default
func expandWithDefaults(s string) string {
	// Buscar patrones ${VAR:default} o ${VAR}
	for {
		start := strings.Index(s, "${")
		if start == -1 {
			break
		}

		end := strings.Index(s[start:], "}")
		if end == -1 {
			break
		}
		end += start

		// Extraer el contenido entre ${ y }
		content := s[start+2 : end]

		// Verificar si tiene sintaxis VAR:default
		parts := strings.SplitN(content, ":", 2)
		varName := parts[0]
		defaultVal := ""
		hasDefault := false

		if len(parts) == 2 {
			defaultVal = parts[1]
			hasDefault = true
		}

		// Obtener el valor de la variable de entorno
		value := os.Getenv(varName)

		// Si la variable está vacía y hay default, usar el default
		if value == "" && hasDefault {
			value = defaultVal
		}

		// Reemplazar en el string
		s = s[:start] + value + s[end+1:]
	}

	return s
}
