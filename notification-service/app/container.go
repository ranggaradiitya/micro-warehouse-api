package app

import (
	"micro-warehouse/notificaiton-service/controller"
	"micro-warehouse/notificaiton-service/pkg/email"
	"micro-warehouse/notificaiton-service/pkg/rabbitmq"
	"micro-warehouse/notificaiton-service/usecase"
)

type Container struct {
	EmailController *controller.EmailController
	EmailUseCase    *usecase.EmailUseCase
	RabbitMQService rabbitmq.RabbitMQServiceInterface
	EmailService    email.EmailServiceInterface
}

func BuildContainer(rabbitMQService rabbitmq.RabbitMQServiceInterface, emailService email.EmailServiceInterface) *Container {
	emailUseCase := usecase.NewEmailUsecase(emailService)
	emailController := controller.NewEmailController(emailUseCase)

	return &Container{
		EmailController: emailController,
		EmailUseCase:    emailUseCase,
		RabbitMQService: rabbitMQService,
		EmailService:    emailService,
	}
}
