package app

import (
	"errors"
	"fmt"

	"github.com/beego/beego/v2/core/validation"
	"github.com/labstack/echo/v4"
)

func SendMessage(c echo.Context, httpCode int, msg string) error {
	status := Status{Kind: KIND_STATUS, Code: httpCode, Message: msg}
	return c.JSON(httpCode, status)
}

func Send(c echo.Context, httpCode int, json interface{}) error {
	return c.JSON(httpCode, json)
}

func Validate(c echo.Context, params []string) error {
	valid := validation.Validation{}

	for _, name := range params {
		valid.Required(c.Param(name), name)
	}

	if valid.HasErrors() {
		for _, err := range valid.Errors {
			return errors.New(fmt.Sprintf("[%s] %s", err.Key, err.Error()))
		}
	}
	return nil
}

func NewStatus(code int, message string) *Status {
	return &Status{
		Kind:    KIND_STATUS,
		Code:    code,
		Message: message,
	}
}
