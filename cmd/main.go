package main

import (
	"clean-arq-layout/internal/app"
)

func main() {
	// La configuración se carga automáticamente desde APP_ENV
	// (ver config.Load() para más detalles)
	app.Start()
}
