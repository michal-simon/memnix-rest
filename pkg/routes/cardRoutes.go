package routes

import (
	"github.com/memnix/memnixrest/app/controllers"

	"github.com/gofiber/fiber/v2"
)

func registerCardRoutes(r fiber.Router) {
	// Get
	r.Get("/cards/today", controllers.GetAllTodayCard)                   // Get all Today's card
	r.Get("/cards/:deckID/training", controllers.GetTrainingCardsByDeck) // Get training card by deck

	r.Get("/mcqs/:deckID", controllers.GetMcqsByDeck) // Get MCQs by deckID

	// Post
	r.Post("/cards/response", controllers.PostResponse)                 // Post a response
	r.Post("/cards/selfresponse", controllers.PostSelfEvaluateResponse) // Post

	// ADMIN ONLY
	r.Get("/cards", controllers.GetAllCards)                   // Get all cards
	r.Get("/cards/id/:id", controllers.GetCardByID)            // Get card by ID
	r.Get("/cards/deck/:deckID", controllers.GetCardsFromDeck) // Get card by deckID

	r.Post("/cards/new", controllers.CreateNewCard) // Create a new card
	r.Post("/mcqs/new", controllers.CreateMcq)      // Create a mcq

	r.Put("/cards/:id/edit", controllers.UpdateCardByID) // Update a card by ID
	r.Put("/mcqs/:id/edit", controllers.UpdateMcqByID)   // Update a mcq by ID

	r.Delete("/cards/:id", controllers.DeleteCardByID) // Delete a card by ID
	r.Delete("/mcqs/:id", controllers.DeleteMcqByID)   // Delete a mcq by ID
}
