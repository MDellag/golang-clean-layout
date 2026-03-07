package dependencies

import (
	"clean-arq-layout/internal/app/modules"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// MockApp creates a minimal test fx application with only the config module
// pre-wired. Pass extra fx.Option values to provide services, repositories,
// and invoke test logic.
//
// Example:
//
//	dependencies.MockApp(t,
//	    fx.Provide(func() services.UsersRepository { return myMock }),
//	    fx.Provide(services.NewUsersService),
//	    fx.Invoke(func(svc *services.UsersService) {
//	        // assertions here
//	    }),
//	).RequireStart().RequireStop()
func MockApp(t testing.TB, opts ...fx.Option) *fxtest.App {
	base := []fx.Option{modules.ConfigModule}
	return fxtest.New(t, append(base, opts...)...)
}
