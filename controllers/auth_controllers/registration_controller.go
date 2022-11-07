package authcontrollers

import (
	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/responses"
)

/*
   Registration Flow:
   1. InitiateRegistration:
      a. Allow user to pass in contact and username to verify they aren't already taken. You can pass other stuff like birthday if u want but its not needed.
      b. Create a temporary object with the contact and a randomly generated code.
      c. Send the code to user's contact
   2. FinalizeRegistration
      a. Find temporary object with the contact. If it doesn't exist give an error message
      b. Assuming object is found, compare the code to the code the user passes in. If it doesn't match give an error message
      c. Create the user and if an error happens to occur where the username now becomes taken but wasn't at the moment of initiate registration, give a custom error message
*/

func InitiateRegistration(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(
		responses.NewSuccessResponse(
			fiber.StatusOK,
			&fiber.Map{
				"data": "Initiate Register",
			},
		),
	)
}

func FinalizeRegistration(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(
		responses.NewSuccessResponse(
			fiber.StatusOK,
			&fiber.Map{
				"data": "Finalize Register",
			},
		),
	)
}
