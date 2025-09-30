package services

import (
	"clean-arq-layout/internal/dependencies"
	"clean-arq-layout/internal/domain/entity"
	"clean-arq-layout/internal/services"
	"clean-arq-layout/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
)

func WithUsersRepository() dependencies.Option {
	usersRepo := new(mocks.UsersRepository)

	usersRepo.On("FindByUsername", "test-user-1").Return(&entity.User{
		ID:       uuid.New(),
		Name:     "test-user-1",
		Email:    "test-user@mail.com",
		Password: "123",
	}, nil)

	return func(c *dig.Container) error {
		return c.Provide(func() services.UsersRepository {
			return usersRepo
		})
	}
}

func TestFindByEmail(t *testing.T) {
	container := dependencies.MockContainer(WithUsersRepository())

	container.Invoke(func(
		usersService services.UsersService,
	) {
		testUser, err := usersService.FindByUsername("test-user-1")
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, "test-user-1", testUser.Name)
		assert.Equal(t, "test-user@mail.com", testUser.Email)
	})

}
