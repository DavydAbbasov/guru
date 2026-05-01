package customerrors

import "github.com/gofiber/fiber/v2"

func FiberWriteError(c *fiber.Ctx, err error) error {
	customErr := Resolve(err)
	return c.Status(customErr.StatusCode).JSON(customErr)
}
