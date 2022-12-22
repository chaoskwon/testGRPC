package controller

import (
	"errors"
	"fmt"
	"testGRPC/gw/entity"
	"testGRPC/gw/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func MemberController(c echo.Context) ([]entity.Member, error) {
	logrus.Trace("")

	param := c.QueryParam("ids")

	if param == "" {
		msg := "ids가 존재하지 않습니다"

		fmt.Print(msg)
		return nil, errors.New(msg)
	}

	ids, err := service.ValidateRequesst(param)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	list, err := service.GetList(c.Request().Context(), ids)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	return list, nil
}
