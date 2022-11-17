package utils

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"nerajima.com/NeraJima/configs"
)

func SendRegistrationEmail(name, email string, code string) {
	var apiKey, emailSender string = configs.EnvSendGridKeyAndFrom()
	var sendgridClient *sendgrid.Client = sendgrid.NewSendClient(apiKey)

	from := mail.NewEmail("NeraJima", emailSender)
	tos := []*mail.Email{ // list of emails to send this email to
		mail.NewEmail(name, email),
	}

	m := mail.NewV3Mail()
	m.SetFrom(from)
	m.SetTemplateID("d-bccd2db8db3e4699b3e636b78bddb90e")

	p := mail.NewPersonalization()
	p.SetDynamicTemplateData("full_name", name)
	p.SetDynamicTemplateData("verification_code", code)
	p.AddTos(tos...)

	m.AddPersonalizations(p)

	res, err := sendgridClient.Send(m)
	_ = res
	_ = err
}

func SendPasswordResetEmail(name, email string, code string) {
	var apiKey, emailSender string = configs.EnvSendGridKeyAndFrom()
	var sendgridClient *sendgrid.Client = sendgrid.NewSendClient(apiKey)

	from := mail.NewEmail("NeraJima", emailSender)
	tos := []*mail.Email{ // list of emails to send this email to
		mail.NewEmail(name, email),
	}

	m := mail.NewV3Mail()
	m.SetFrom(from)
	m.SetTemplateID("d-7333e78a73e946638808809e4020df8b")

	p := mail.NewPersonalization()
	p.SetDynamicTemplateData("verification_code", code)
	p.AddTos(tos...)

	m.AddPersonalizations(p)

	res, err := sendgridClient.Send(m)
	_ = res
	_ = err
}
