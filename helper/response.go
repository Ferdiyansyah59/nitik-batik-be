package helper

import "strings"

type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    interface{} `json:"data"`
	Token   string      `json:"token"`
}

type EmptyObj struct{}

func BuildResponse(status bool, message string, data interface{}) Response {
	res := Response{
		Status:  status,
		Message: message,
		Error:   nil,
		Data:    data,
	}
	return res
}
func BuildResponseLogin(status bool, message string, data interface{}, token string) Response {
	res := Response{
		Status:  status,
		Message: message,
		Error:   nil,
		Data:    data,
		Token:   token,
	}
	return res
}

func BuildErrorResponse(message string, err string, data interface{}) Response {
	splittedError := strings.Split(err, "\n")
	res := Response{
		Status:  false,
		Message: message,
		Error:   splittedError,
		Data:    data,
	}
	return res
}
