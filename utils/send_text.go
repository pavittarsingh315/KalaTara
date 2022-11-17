package utils

import (
	"fmt"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"

	"nerajima.com/NeraJima/configs"
)

func SendRegistrationText(code string, number string) {
	var accountID, authToken, fromNumber = configs.EnvTwilioIDKeyFrom()
	var twilioClient = twilio.NewRestClientWithParams(twilio.ClientParams{AccountSid: accountID, Password: authToken})

	message := fmt.Sprintf("Here is your NeraJima verification code: %s. Code expires in 5 minutes!", code)

	params := &openapi.CreateMessageParams{}
	params.SetTo(number)
	params.SetFrom(fromNumber)
	params.SetBody(message)

	res, err := twilioClient.Api.CreateMessage(params)
	_ = res
	_ = err
}

func SendPasswordResetText(code string, number string) {
	var accountID, authToken, fromNumber = configs.EnvTwilioIDKeyFrom()
	var twilioClient = twilio.NewRestClientWithParams(twilio.ClientParams{AccountSid: accountID, Password: authToken})

	message := fmt.Sprintf("Here is your NeraJima password reset code: %s. Code expires in 5 minutes!", code)

	params := &openapi.CreateMessageParams{}
	params.SetTo(number)
	params.SetFrom(fromNumber)
	params.SetBody(message)

	res, err := twilioClient.Api.CreateMessage(params)
	_ = res
	_ = err
}
