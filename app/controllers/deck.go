package controllers

import (
	"fmt"
	"github.com/memnix/memnixrest/app/models"
	"github.com/memnix/memnixrest/app/queries"
	"github.com/memnix/memnixrest/pkg/database"
	"github.com/memnix/memnixrest/pkg/utils"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GET

// GetAllDecks function to get all decks
// @Description Get every deck. Shouldn't really be used !
// @Security Admin
// @Deprecated
// @Summary gets all decks
// @Tags Deck
// @Produce json
// @Success 200 {array} models.Deck
// @Router /v1/decks [get]
func GetAllDecks(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	auth := CheckAuth(c, models.PermAdmin) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	var decks []models.Deck

	if res := db.Find(&decks); res.Error != nil {
		return queries.RequestError(c, http.StatusInternalServerError, utils.ErrorRequestFailed)
	}

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Get all decks",
		Data:    decks,
		Count:   len(decks),
	})
}

// GetDeckByID function to get a deck
// @Description Get a deck by ID
// @Summary get a deck
// @Security Admin
// @Tags Deck
// @Produce json
// @Param id path int true "Deck ID"
// @Success 200 {model} models.Deck
// @Router /v1/decks/{deckID} [get]
func GetDeckByID(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	// Params
	id := c.Params("deckID")

	auth := CheckAuth(c, models.PermAdmin) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	deck := new(models.Deck)

	if err := db.First(&deck, id).Error; err != nil {
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Success get deck by ID.",
		Data:    *deck,
		Count:   1,
	})
}

// GetAllSubDecks method to get a deck
// @Description Get decks a user is sub to
// @Summary gets a list of deck
// @Tags Deck
// @Produce json
// @Security Beaver
// @Success 200 {array} models.ResponseDeck
// @Router /v1/decks/sub [get]
func GetAllSubDecks(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	auth := CheckAuth(c, models.PermUser) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	var accesses []models.Access // Accesses array

	if err := db.Joins("Deck").Joins("User").Where("accesses.user_id = ? AND accesses.permission >= ?", auth.User.ID, models.AccessStudent).Find(&accesses).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on GetAllSubDecks: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}
	responseDeck := make([]models.ResponseDeck, len(accesses))

	for i := range accesses {
		responseDeck[i] = queries.FillResponseDeck(&accesses[i].Deck, accesses[i].Permission, accesses[i].ToggleToday)
	}

	sort.Slice(responseDeck, func(i, j int) bool {
		return responseDeck[i].Deck.DeckName < responseDeck[j].Deck.DeckName
	})

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Get all sub decks",
		Data:    responseDeck,
		Count:   len(responseDeck),
	})
}

// GetAllEditorDecks method to get a deck
// @Description Get decks the user is an editor
// @Summary gets a list of deck
// @Tags Deck
// @Produce json
// @Security Beaver
// @Success 200 {array} models.ResponseDeck
// @Router /v1/decks/editor [get]
func GetAllEditorDecks(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	auth := CheckAuth(c, models.PermUser) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	var accesses []models.Access // Accesses array

	if err := db.Joins("Deck").Joins("User").Where("accesses.user_id = ? AND accesses.permission >= ?", auth.User.ID, models.AccessEditor).Find(&accesses).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on GetAllEditorDecks: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}

	responseDeck := make([]models.ResponseDeck, len(accesses))

	for i := range accesses {
		responseDeck[i] = queries.FillResponseDeck(&accesses[i].Deck, accesses[i].Permission, accesses[i].ToggleToday)
	}

	sort.Slice(responseDeck, func(i, j int) bool {
		return responseDeck[i].Deck.DeckName < responseDeck[j].Deck.DeckName
	})

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Get all sub decks",
		Data:    responseDeck,
		Count:   len(responseDeck),
	})
}

// GetAllSubUsers method to get a list of users
// @Description Get all the sub users to a deck
// @Summary gets a list of users
// @Tags Deck
// @Security Admin
// @Produce json
// @Success 200 {array} models.User
// @Router /v1/decks/{deckID}/users [get]
func GetAllSubUsers(c *fiber.Ctx) error { // Params
	deckID := c.Params("deckID")
	var result *models.ResponseHTTP
	auth := CheckAuth(c, models.PermAdmin) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	var users []models.User
	id, _ := strconv.ParseUint(deckID, 10, 32)

	if result = queries.GetSubUsers(uint(id)); !result.Success {
		log := models.CreateLog(fmt.Sprintf("Error from %s on GetAllSubUsers: %s", auth.User.Email, result.Message), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, utils.ErrorRequestFailed)
	}

	switch result.Data.(type) {
	default:
		return queries.RequestError(c, http.StatusInternalServerError, utils.ErrorRequestFailed)
	case []models.User:
		users = result.Data.([]models.User)
	}

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Get all sub users",
		Data:    users,
		Count:   len(users),
	})
}

// GetAllAvailableDecks method to get a list of deck
// @Description Get all public deck that you are not sub to
// @Summary get a list of deck
// @Tags Deck
// @Produce json
// @Security Beaver
// @Success 200 {array} models.ResponseDeck
// @Router /v1/decks/available [get]
func GetAllAvailableDecks(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	auth := CheckAuth(c, models.PermUser) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	var decks []models.Deck

	if err := db.Raw(
		"SELECT DISTINCT public.decks.* FROM public.decks LEFT JOIN public.accesses ON public.decks.id = public.accesses.deck_id AND public.accesses.user_id = ? WHERE "+
			"(public.accesses.deck_id IS NULL  OR public.accesses.permission = 0 OR public.accesses.deleted_at IS NOT NULL) AND public.decks.status = 3",
		auth.User.ID).Scan(&decks).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on GetAllAvailableDecks: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}

	responseDeck := make([]models.ResponseDeck, len(decks))

	for i := range decks {
		responseDeck[i] = queries.FillResponseDeck(&decks[i], models.AccessNone, false)
	}

	sort.Slice(responseDeck, func(i, j int) bool {
		return responseDeck[i].Deck.DeckName < responseDeck[j].Deck.DeckName
	})

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Get all available decks",
		Data:    responseDeck,
		Count:   len(responseDeck),
	})
}

// GetAllPublicDecks method to get a list of deck
// @Description Get all public deck
// @Summary gets a list of deck
// @Tags Deck
// @Security Beaver
// @Produce json
// @Success 200 {array} models.Deck
// @Router /v1/decks/public [get]
func GetAllPublicDecks(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	auth := CheckAuth(c, models.PermAdmin) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	var decks []models.Deck

	if err := db.Where("decks.status = ?", models.DeckPublic).Find(&decks).Error; err != nil {
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}
	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Get all public decks",
		Data:    decks,
		Count:   len(decks),
	})
}

// POST

// CreateNewDeck method
// @Description Create a new deck
// @Summary creates a deck
// @Tags Deck
// @Produce json
// @Security Beaver
// @Success 200 {object} models.Deck
// @Accept json
// @Param deck body models.Deck true "Deck to create"
// @Router /v1/decks/new [post]
func CreateNewDeck(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	auth := CheckAuth(c, models.PermUser) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	deck := new(models.Deck)

	if err := c.BodyParser(&deck); err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on CreateNewDeck: %s", auth.User.Email, err.Error()), models.LogBodyParserError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusBadRequest, err.Error())
	}

	if len(strings.TrimSpace(deck.Key)) != utils.DeckKeyLen {
		deck.Key = strings.ToUpper(strings.ReplaceAll(deck.DeckName, " ", "")[0:utils.DeckKeyLen])
	}

	if deck.NotValidate() {
		log := models.CreateLog(fmt.Sprintf("Error from %s on CreateNewDeck: BadRequest", auth.User.Email), models.LogBadRequest).SetType(models.LogTypeWarning).AttachIDs(auth.User.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusBadRequest, utils.ErrorDeckName)
	}

	if res := queries.CheckDeckLimit(&auth.User); !res {
		log := models.CreateLog(fmt.Sprintf("Forbidden from %s on CreateNewDeck: This user has reached his limit", auth.User.Email), models.LogUserDeckLimit).SetType(models.LogTypeWarning).AttachIDs(auth.User.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusBadRequest, "You can't create more deck !")
	}

	deck.Key = strings.ToUpper(deck.Key)
	deck.GenerateCode()

	i := 0
	for !queries.CheckCode(deck.Key, deck.Code) {
		if i > 10 {
			log := models.CreateLog(fmt.Sprintf("Error from %s on CreateNewDeck Generate Code: Couldn't generate code after 10 attempts", auth.User.Email), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, 0, 0)
			_ = log.SendLog()
			return queries.RequestError(c, http.StatusBadRequest, utils.ErrorRequestFailed)
		}
		deck.GenerateCode()
		i++
	}

	deck.Status = models.DeckPrivate
	db.Create(deck)

	log := models.CreateLog(fmt.Sprintf("Created: %d - %s", deck.ID, deck.DeckName), models.LogDeckCreated).SetType(models.LogTypeInfo).AttachIDs(auth.User.ID, deck.ID, 0)
	_ = log.SendLog()

	if err := queries.GenerateCreatorAccess(&auth.User, deck); !err.Success {
		log = models.CreateLog(fmt.Sprintf("Error from %s on CreateNewDeck Generate Creator Access: %s", auth.User.Email, err.Message), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusBadRequest, err.Message)
	}

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Success register a deck",
		Data:    *deck,
		Count:   1,
	})
}

// UnSubToDeck method
// @Description Unsubscribe to a deck
// @Summary unsub deck
// @Tags Deck
// @Security Beaver
// @Produce json
// @Success 200
// @Router /v1/decks/{deckID}/unsubscribe [post]
func UnSubToDeck(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	auth := CheckAuth(c, models.PermUser) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	// Params
	deckID := c.Params("deckID")
	deckidInt, _ := strconv.ParseUint(deckID, 10, 32)

	access := new(models.Access)
	if err := db.Joins("User").Joins("Deck").Where("accesses.user_id = ? AND accesses.deck_id = ?", auth.User.ID, deckID).Find(&access).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on UnSubToDeck to %d: %s", auth.User.Email, deckidInt, err.Error()), models.LogBadRequest).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusBadRequest, utils.ErrorNotSub)
	}

	if access.Permission == 0 {
		return queries.RequestError(c, http.StatusForbidden, utils.ErrorNotSub)
	}

	access.Permission = 0
	db.Preload("User").Preload("Deck").Save(access)

	log := models.CreateLog(fmt.Sprintf("Unsubscribed: User - %d (%s) | Deck - %d (%s)", access.UserID, access.User.Username, access.DeckID, access.Deck.DeckName), models.LogUnsubscribe).SetType(models.LogTypeInfo).AttachIDs(auth.User.ID, access.DeckID, 0)
	_ = log.SendLog()

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Success unsub to the deck",
		Data:    nil,
		Count:   0,
	})
}

// SubToPrivateDeck method
// @Description Subscribe to a private deck
// @Summary sub deck
// @Tags Deck
// @Produce json
// @Success 200
// @Param key path string true "Deck unique Key"
// @Param code path string true "Deck unique Code"
// @Security Beaver
// @Router /v1/decks/private/{key}/{code}/subscribe [post]
func SubToPrivateDeck(c *fiber.Ctx) error {
	db := database.DBConn
	key := c.Params("key")
	code := c.Params("code")

	deck := new(models.Deck)

	// Check auth
	auth := CheckAuth(c, models.PermUser)
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	if err := db.Where("decks.key = ? AND decks.code = ? AND decks.share IS TRUE", key, code).First(&deck).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on SubToPrivateDeck: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}

	if err := queries.GenerateAccess(&auth.User, deck); !err.Success {
		log := models.CreateLog(fmt.Sprintf("Error from %s on SubToPrivateDeck: %s", auth.User.Email, err.Message), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, deck.ID, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Message)
	}

	if err := queries.PopulateMemDate(&auth.User, deck); !err.Success {
		log := models.CreateLog(fmt.Sprintf("Error from %s on SubToPrivateDeck: %s", auth.User.Email, err.Message), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, deck.ID, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Message)
	}

	log := models.CreateLog(fmt.Sprintf("SubToPrivateDeck: User - %d (%s)| Deck - %d (%s)", auth.User.ID, auth.User.Username, deck.ID, deck.DeckName), models.LogSubscribe).SetType(models.LogTypeInfo).AttachIDs(auth.User.ID, deck.ID, 0)
	_ = log.SendLog()

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Success subscribing to deck",
		Data:    nil,
		Count:   0,
	})
}

// PublishDeckRequest method
// @Description Request to publish deck
// @Summary publishes a deck
// @Tags Deck
// @Produce json
// @Success 200
// @Param deckID path string true "Deck ID"
// @Security Beaver
// @Router /v1/decks/{deckID}/publish [post]
func PublishDeckRequest(c *fiber.Ctx) error {
	db := database.DBConn
	id := c.Params("deckID")
	deckidInt, _ := strconv.ParseUint(id, 10, 32)

	deck := new(models.Deck)

	// Check auth
	auth := CheckAuth(c, models.PermUser)
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	if err := db.First(&deck, id).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on PublishDeckRequest: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}

	if res := queries.CheckAccess(auth.User.ID, deck.ID, models.AccessOwner); !res.Success {
		log := models.CreateLog(fmt.Sprintf("Forbidden from %s on deck %d - PublishDeckRequest: %s", auth.User.Email, deckidInt, res.Message), models.LogPermissionForbidden).SetType(models.LogTypeWarning).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusForbidden, utils.ErrorForbidden)
	}

	deck.Status = models.DeckWaitingReview

	db.Save(deck)
	// TODO: Error handling

	log := models.CreateLog(fmt.Sprintf("PublishDeckRequest: User - %d (%s)| Deck - %d (%s)", auth.User.ID, auth.User.Username, deck.ID, deck.DeckName), models.LogPublishRequest).SetType(models.LogTypeInfo).AttachIDs(auth.User.ID, deck.ID, 0)
	_ = log.SendLog()

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Success sending a publish request to deck",
		Data:    nil,
		Count:   0,
	})
}

// SubToDeck method
// @Description Subscribe to a deck
// @Summary sub deck
// @Tags Deck
// @Produce json
// @Success 200
// @Param deckID path string true "Deck ID"
// @Security Beaver
// @Router /v1/decks/{deckID}/subscribe [post]
func SubToDeck(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn
	deckID := c.Params("deckID")
	deckidInt, _ := strconv.ParseUint(deckID, 10, 32)

	deck := new(models.Deck)

	// Check auth
	auth := CheckAuth(c, models.PermUser)
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	if err := db.First(&deck, deckID).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on SubToDeck: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}

	if err := queries.GenerateAccess(&auth.User, deck); !err.Success {
		log := models.CreateLog(fmt.Sprintf("Error from %s on SubToDeck: %s", auth.User.Email, err.Message), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Message)
	}

	if err := queries.PopulateMemDate(&auth.User, deck); !err.Success {
		log := models.CreateLog(fmt.Sprintf("Error from %s on SubToDeck: %s", auth.User.Email, err.Message), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Message)
	}

	log := models.CreateLog(fmt.Sprintf("Subscribed: User - %d (%s)| Deck - %d (%s)", auth.User.ID, auth.User.Username, deck.ID, deck.DeckName), models.LogSubscribe).SetType(models.LogTypeInfo).AttachIDs(auth.User.ID, deck.ID, 0)
	_ = log.SendLog()

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Success subscribing to deck",
		Data:    nil,
		Count:   0,
	})
}

// PUT

// UpdateDeckByID method
// @Description Edit a deck
// @Summary edits a deck
// @Tags Deck
// @Produce json
// @Success 200
// @Accept json
// @Param deck body models.Deck true "Deck to edit"
// @Param deckID path string true "Deck ID"
// @Security Beaver
// @Router /v1/decks/{deckID}/edit [put]
func UpdateDeckByID(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	// Params
	id := c.Params("deckID")
	deckidInt, _ := strconv.ParseUint(id, 10, 32)

	auth := CheckAuth(c, models.PermUser) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	deck := new(models.Deck)

	if err := db.First(&deck, id).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on UpdateDeckByID: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}

	if res := queries.CheckAccess(auth.User.ID, deck.ID, models.AccessOwner); !res.Success {
		log := models.CreateLog(fmt.Sprintf("Forbidden from %s on deck %d - UpdateDeckByID: %s", auth.User.Email, deckidInt, res.Message), models.LogPermissionForbidden).SetType(models.LogTypeWarning).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusForbidden, utils.ErrorForbidden)
	}

	if err := UpdateDeck(c, deck); !err.Success {
		log := models.CreateLog(fmt.Sprintf("Error on UpdateDeckByID: %s from %s", err.Message, auth.User.Email), models.LogBadRequest).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusBadRequest, err.Message)
	}

	log := models.CreateLog(fmt.Sprintf("Updated: %d - %s", deck.ID, deck.DeckName), models.LogDeckEdited).SetType(models.LogTypeInfo).AttachIDs(auth.User.ID, deck.ID, 0)
	_ = log.SendLog()

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Success update deck by ID",
		Data:    *deck,
		Count:   1,
	})
}

// UpdateDeck function
func UpdateDeck(c *fiber.Ctx, deck *models.Deck) *models.ResponseHTTP {
	db := database.DBConn

	deckStatus := deck.Status

	res := new(models.ResponseHTTP)

	if err := c.BodyParser(&deck); err != nil {
		res.GenerateError(err.Error())
		return res
	}

	if deck.Status != deckStatus {
		res.GenerateError(utils.ErrorBreak)
		return res
	}

	if deck.NotValidate() {
		res.GenerateError(utils.ErrorDeckName)
		return res
	}

	if len(strings.TrimSpace(deck.Key)) != utils.DeckKeyLen {
		deck.Key = strings.ToUpper(strings.ReplaceAll(deck.DeckName, " ", "")[0:utils.DeckKeyLen])
	}

	deck.Key = strings.ToUpper(deck.Key)
	deck.GenerateCode()

	i := 0
	for !queries.CheckCode(deck.Key, deck.Code) {
		if i > 10 {
			res.GenerateError(utils.ErrorRequestFailed)
			return res
		}
		deck.GenerateCode()
		i++
	}

	db.Save(deck)

	res.GenerateSuccess("Success update deck", nil, 0)
	return res
}

// DeleteDeckById method
// @Description Delete a deck (must be deck owner)
// @Summary delete a deck
// @Tags Deck
// @Produce json
// @Success 200
// @Param deckID path string true "Deck ID"
// @Security Beaver
// @Router /v1/decks/{deckID} [delete]
func DeleteDeckById(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn
	id := c.Params("deckID")
	deckidInt, _ := strconv.ParseUint(id, 10, 32)

	auth := CheckAuth(c, models.PermUser) // Check auth
	if !auth.Success {
		return queries.AuthError(c, &auth)
	}

	deck := new(models.Deck)

	if err := db.First(&deck, id).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on DeleteDeckById: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, err.Error())
	}

	if res := queries.CheckAccess(auth.User.ID, deck.ID, models.AccessOwner); !res.Success {
		log := models.CreateLog(fmt.Sprintf("Forbidden from %s on deck %d - DeleteDeckById: %s", auth.User.Email, deckidInt, res.Message), models.LogPermissionForbidden).SetType(models.LogTypeWarning).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusForbidden, utils.ErrorForbidden)
	}

	var memDates []models.MemDate
	var accesses []models.Access
	var cards []models.Card

	if err := db.Joins("Card").Where("mem_dates.deck_id = ?", deck.ID).Find(&memDates).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on DeleteDeckById: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, utils.ErrorRequestFailed)
		// TODO: Error
	}

	if err := db.Where("accesses.deck_id = ?", deck.ID).Find(&accesses).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on DeleteDeckById: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, utils.ErrorRequestFailed)
		// TODO: Error
	}

	if err := db.Where("cards.deck_id = ?", deck.ID).Find(&cards).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error from %s on DeleteDeckById: %s", auth.User.Email, err.Error()), models.LogQueryGetError).SetType(models.LogTypeError).AttachIDs(auth.User.ID, uint(deckidInt), 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusInternalServerError, utils.ErrorRequestFailed)
	}

	db.Unscoped().Delete(memDates)
	db.Unscoped().Delete(accesses)

	db.Delete(cards)
	db.Delete(deck)

	log := models.CreateLog(fmt.Sprintf("Deleted: %d - %s", deck.ID, deck.DeckName), models.LogDeckDeleted).SetType(models.LogTypeInfo).AttachIDs(auth.User.ID, deck.ID, 0)
	_ = log.SendLog()

	return c.Status(http.StatusOK).JSON(models.ResponseHTTP{
		Success: true,
		Message: "Success delete deck by ID",
		Data:    *deck,
		Count:   1,
	})
}
