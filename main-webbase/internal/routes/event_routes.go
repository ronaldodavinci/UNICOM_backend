package routes

import (
	"github.com/gofiber/fiber/v2"
	"main-webbase/internal/controllers"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupRoutesEvent(app *fiber.App, client *mongo.Client) {
	event := app.Group("/event")

	event.Post("/", controllers.CreateEventHandler())                         
	event.Get("/", controllers.GetAllVisibleEventHandler())                   
	event.Get("/:event_id", controllers.GetEventDetailHandler())                 
	event.Delete("/:event_id", controllers.DeleteEventHandler)  
	event.Post("/participate/:event_id", controllers.ParticipateEventWithNoFormHandler()) 

	event.Post("/:eventId/qa", controllers.CreateEventQAHandler(client))

	// Form Section
	form := app.Group("/event/:eventId/form")
	form.Post("/initialize", controllers.InitializeFormHandler())
	form.Post("/disable", controllers.DisableFormHandler())
	form.Post("/questions", controllers.CreateFormQuestionHandler())
	form.Get("/questions", controllers.GetFormQuestionHandler())
	form.Post("/answers", controllers.CreateUserAnswerHandler())
	form.Get("/matrix", controllers.GetAllUserAnswerandQuestionHandler())

	// Participant Status
	participant := event.Group("/participant")
	participant.Put("/status", controllers.UpdateParticipantStatusHandler())
	participant.Get("/mystatus/:eventId", controllers.GetMyParticipantStatusHandler())
}
