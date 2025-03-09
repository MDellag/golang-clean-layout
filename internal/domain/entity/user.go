package entity

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	Password  string // Almacenado como hash
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Métodos de comportamiento de la entidad
func (u *User) ChangePassword(newPassword string) error {
	// Lógica para cambiar contraseña
	return nil
}
