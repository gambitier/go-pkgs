package commonobservability

import (
	fiberotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/fiber/v3"
)

// FiberMiddleware returns Fiber middleware that records server spans via contrib/v3/otel.
// Health and swagger paths are skipped (see isHealthCheckPath).
func FiberMiddleware(serviceName string) fiber.Handler {
	if !IsEnabled() {
		return func(c fiber.Ctx) error { return c.Next() }
	}
	_ = serviceName // service.name is set on the OTel resource in Init()
	return fiberotel.Middleware(
		fiberotel.WithNext(func(c fiber.Ctx) bool {
			return isHealthCheckPath(c.Path())
		}),
	)
}
