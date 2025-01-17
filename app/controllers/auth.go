package controllers

import (
	"fmt"
	"github.com/memnix/memnixrest/app/models"
	"github.com/memnix/memnixrest/app/queries"
	"github.com/memnix/memnixrest/pkg/database"
	"github.com/memnix/memnixrest/pkg/utils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var SecretKey string // SecretKey env variable

func Init() {
	SecretKey = os.Getenv("SECRET") // SecretKey env variable
}

// Register function to create a new user
// @Description Create a new user
// @Summary creates a new user
// @Tags Auth
// @Produce json
// @Param credentials body models.RegisterStruct true "Credentials"
// @Success 200 {object} models.User
// @Failure 403 "Forbidden"
// @Router /v1/register [post]
func Register(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	var data models.RegisterStruct // Data object

	if err := c.BodyParser(&data); err != nil {
		return err
	} // Parse body

	// Register checks
	if len(data.Password) > utils.MaxPasswordLen || len(data.Username) > utils.MaxUsernameLen || len(data.Email) > utils.MaxEmailLen {
		log := models.CreateLog(fmt.Sprintf("Error on register: %s - %s", data.Username, data.Email), models.LogBadRequest).SetType(models.LogTypeWarning).AttachIDs(0, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusForbidden, utils.ErrorRequestFailed)
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data.Password), 10) // Hash password
	user := models.User{
		Username: data.Username,
		Email:    strings.ToLower(data.Email),
		Password: password,
	} // Create object

	//TODO: manual checking for unique username and email
	if err := db.Create(&user).Error; err != nil {
		log := models.CreateLog(fmt.Sprintf("Error on register: %s - %s", data.Username, data.Email), models.LogAlreadyUsedEmail).SetType(models.LogTypeWarning).AttachIDs(user.ID, 0, 0)
		_ = log.SendLog()
		return queries.RequestError(c, http.StatusForbidden, utils.ErrorAlreadyUsedEmail)
	} // Add user to DB

	// Create log
	log := models.CreateLog(fmt.Sprintf("Register: %s - %s", user.Username, user.Email), models.LogUserRegister).SetType(models.LogTypeInfo).AttachIDs(user.ID, 0, 0)
	_ = log.SendLog() // Send log

	return c.JSON(user) // Return user
}

// Login function to log in a user and return access with fresh token
// @Description Login the user and return a fresh token
// @Summary logins user and return a fresh token
// @Tags Auth
// @Produce json
// @Param credentials body models.LoginStruct true "Credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 "Incorrect password or email"
// @Failure 500 "Internal error"
// @Router /v1/login [post]
func Login(c *fiber.Ctx) error {
	db := database.DBConn // DB Conn

	var data models.LoginStruct // Data object

	if err := c.BodyParser(&data); err != nil {
		return err
	} // Parse body

	var user models.User // User object

	db.Where("email = ?", strings.ToLower(data.Email)).First(&user) // Get user

	// handle error
	if user.ID == 0 { // default Id when return nil
		// Create log
		log := models.CreateLog(fmt.Sprintf("Error on login: %s", data.Email), models.LogIncorrectEmail).SetType(models.LogTypeWarning).AttachIDs(user.ID, 0, 0)
		_ = log.SendLog()                // Send log
		c.Status(fiber.StatusBadRequest) // BadRequest Status
		// return error message as Json object
		return c.JSON(models.LoginResponse{
			Message: "Incorrect email or password !",
			Token:   "",
		})
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data.Password)); err != nil {
		c.Status(fiber.StatusBadRequest) // BadRequest Status
		log := models.CreateLog(fmt.Sprintf("Error on login: %s", data.Email), models.LogIncorrectPassword).SetType(models.LogTypeWarning).AttachIDs(user.ID, 0, 0)
		_ = log.SendLog() // Send log
		// return error message as Json object
		return c.JSON(models.LoginResponse{
			Message: "Incorrect email or password !",
			Token:   "",
		})
	}

	// Create token
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    strconv.Itoa(int(user.ID)),
		ExpiresAt: time.Now().Add(time.Hour * 336).Unix(), // 14 day
	}) // expires after 2 weeks

	token, err := claims.SignedString([]byte(SecretKey)) // Sign token
	if err != nil {
		log := models.CreateLog(fmt.Sprintf("Error on login: %s", err.Error()), models.LogLoginError).SetType(models.LogTypeError).AttachIDs(user.ID, 0, 0)
		_ = log.SendLog()                         // Send log
		c.Status(fiber.StatusInternalServerError) // InternalServerError Status
		// return error message as Json object
		return c.JSON(models.LoginResponse{
			Message: "Incorrect email or password !",
			Token:   "",
		})
	}

	log := models.CreateLog(fmt.Sprintf("Login: %s - %s", user.Username, user.Email), models.LogUserLogin).SetType(models.LogTypeInfo).AttachIDs(user.ID, 0, 0)
	_ = log.SendLog() // Send log

	// return token as Json object
	return c.JSON(models.LoginResponse{
		Message: "Login Succeeded",
		Token:   token,
	})
}

// User function to get connected user
// @Description Get connected user
// @Summary  gets connected user
// @Tags Auth
// @Produce json
// @Success 200 {object} models.ResponseAuth
// @Failure 401 "Forbidden"
// @Security Beaver
// @Router /v1/user [get]
func User(c *fiber.Ctx) error {
	statusCode, response := IsConnected(c) // Check if connected

	user := new(models.PublicUser)

	user.Set(&response.User) // Set user

	responseUser := models.ResponsePublicAuth{
		Success: response.Success,
		Message: response.Message,
		User:    *user,
	}

	return c.Status(statusCode).JSON(responseUser) // Return response
}

// Logout function to log user logout
// @Description Logout the user and create a record in the log
// @Summary logouts the user
// @Tags Auth
// @Produce json
// @Success 200 "Success"
// @Failure 401 "Forbidden"
// @Security Beaver
// @Router /v1/logout [post]
func Logout(c *fiber.Ctx) error {
	auth := CheckAuth(c, models.PermUser) // Check auth
	if !auth.Success {
		// Return error
		return c.Status(http.StatusUnauthorized).JSON(models.ResponseHTTP{
			Success: false,
			Message: auth.Message,
			Data:    nil,
			Count:   0,
		})
	}

	// Create log
	log := models.CreateLog(fmt.Sprintf("Logout: %s - %s", auth.User.Username, auth.User.Email), models.LogUserLogout).SetType(models.LogTypeInfo).AttachIDs(auth.User.ID, 0, 0)
	_ = log.SendLog()

	// Return response with success
	return c.JSON(fiber.Map{
		"message": "successfully logged out !",
		"token":   "",
	})
}

// AuthDebugMode function to bypass auth in debug mode
func AuthDebugMode(c *fiber.Ctx) models.ResponseAuth {
	db := database.DBConn // DB Conn
	var user models.User  // User object

	// Get user
	if res := db.Where("id = ?", 6).First(&user); res.Error != nil {
		c.Status(fiber.StatusInternalServerError) // InternalServerError Status
		// return error message as Json object
		return models.ResponseAuth{
			Success: false,
			Message: "Failed to get the user. Try to logout/login. Otherwise, contact the support",
		}
	}

	return models.ResponseAuth{
		Success: true,
		Message: "Authenticated",
		User:    user,
	}
}

// CheckAuth function to check if user is connected
func CheckAuth(c *fiber.Ctx, p models.Permission) models.ResponseAuth {
	statusCode, response := IsConnected(c) // Check if connected

	// Check statusCode
	if statusCode != fiber.StatusOK {
		c.Status(statusCode)
		// Return response
		return response
	}

	user := response.User // Get user from response

	// Check permission
	if user.Permissions < p {
		// Log permission error
		log := models.CreateLog(fmt.Sprintf("Permission error: %s | had %s but tried %s", user.Email, user.Permissions.ToString(), p.ToString()), models.LogPermissionForbidden).SetType(models.LogTypeWarning).AttachIDs(user.ID, 0, 0)
		_ = log.SendLog()                  // Send log
		c.Status(fiber.StatusUnauthorized) // Unauthorized Status
		// Return response
		return models.ResponseAuth{
			Success: false,
			Message: "You don't have the right permissions to perform this request.",
		}
	}

	// Validate permissions
	return models.ResponseAuth{
		Success: true,
		Message: "Authenticated",
		User:    user,
	}
}

// IsConnected function to check if user is connected
func IsConnected(c *fiber.Ctx) (int, models.ResponseAuth) {
	db := database.DBConn          // DB Conn
	tokenString := extractToken(c) // Extract token
	var user models.User           // User object

	// Parse token
	token, err := jwt.Parse(tokenString, jwtKeyFunc)
	if err != nil {
		// Return error
		return fiber.StatusForbidden, models.ResponseAuth{
			Success: false,
			Message: "Failed to get the user. Try to logout/login. Otherwise, contact the support",
			User:    user,
		}
	}
	// Check if token is valid
	claims := token.Claims.(jwt.MapClaims)

	// Get user from token
	if res := db.Where("id = ?", claims["iss"]).First(&user); res.Error != nil {
		// Generate log
		log := models.CreateLog(fmt.Sprintf("Error on check auth: %s", res.Error), models.LogLoginError).SetType(models.LogTypeError).AttachIDs(user.ID, 0, 0)
		_ = log.SendLog()                         // Send log
		c.Status(fiber.StatusInternalServerError) // InternalServerError Status
		// return error
		return fiber.StatusInternalServerError, models.ResponseAuth{
			Success: false,
			Message: "Failed to get the user. Try to logout/login. Otherwise, contact the support",
			User:    user,
		}
	}

	// User is connected
	return fiber.StatusOK, models.ResponseAuth{
		Success: true,
		Message: "User is connected",
		User:    user,
	}
}

// extractToken function to extract token from header
func extractToken(c *fiber.Ctx) string {
	token := c.Get("Authorization") // Get token from header
	// Normally Authorization HTTP header.
	onlyToken := strings.Split(token, " ") // Split token
	if len(onlyToken) == 2 {
		return onlyToken[1] // Return only token
	}
	return "" // Return empty string
}

// jwtKeyFunc function to get the key for the token
func jwtKeyFunc(_ *jwt.Token) (interface{}, error) {
	return []byte(SecretKey), nil // Return secret key
}
