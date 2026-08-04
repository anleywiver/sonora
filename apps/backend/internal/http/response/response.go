// Package response implements the API-wide envelope:
// { "data": ... } on success, { "error": { code, message, request_id } } on failure.
package response

import "github.com/gofiber/fiber/v2"

func OK(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{"data": data})
}

func Fail(c *fiber.Ctx, status int, code, message string) error {
	requestID, _ := c.Locals("requestid").(string)
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{
			"code":       code,
			"message":    message,
			"request_id": requestID,
		},
	})
}
