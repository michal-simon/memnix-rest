package models

// ResponseHTTP structure
type ResponseHTTP struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Count   int         `json:"count"`
}

type CardResponse struct {
	CardID   uint `json:"card_id" example:"1"`
	Card     Card
	Response string `json:"response" example:"42"`
}

type CardResponseValidation struct {
	Validate bool   `json:"validate" example:"true"`
	Message  string `json:"message" example:"Correct answer"`
}

type ResponseCard struct {
	Card    Card
	Answers []string
}

type ResponseAuth struct {
	Success bool
	User    User
	Message string
}

type Permission int64

const (
	PermUser Permission = iota
	PermMod
	PermAdmin
)

func (s Permission) ToString() string {
	switch s {
	case PermUser:
		return "PermUser"
	case PermMod:
		return "PermMod"
	case PermAdmin:
		return "PermAdmin"
	default:
		return "Unknown"
	}
}
