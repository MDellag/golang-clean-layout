package services

import (
	"clean-arq-layout/internal/dependencies"
	"clean-arq-layout/internal/domain/entity"
	"clean-arq-layout/internal/services"
	"clean-arq-layout/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestFindByEmail(t *testing.T) {
	usersRepo := new(mocks.UsersRepository)
	usersRepo.On("FindByUsername", "test-user-1").Return(&entity.User{
		ID:       uuid.New(),
		Name:     "test-user-1",
		Email:    "test-user@mail.com",
		Password: "123",
	}, nil)

	dependencies.MockApp(t,
		fx.Provide(func() services.UsersRepository { return usersRepo }),
		fx.Provide(services.NewUsersService),
		fx.Invoke(func(svc *services.UsersService) {
			testUser, err := svc.FindByUsername("test-user-1")
			assert.NoError(t, err)
			assert.Equal(t, "test-user-1", testUser.Name)
			assert.Equal(t, "test-user@mail.com", testUser.Email)
		}),
	).RequireStart().RequireStop()
}
