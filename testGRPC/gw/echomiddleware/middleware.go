package echomiddleware

import (
	"context"
	"testGRPC/gw/common"

	"github.com/go-xorm/xorm"
	"github.com/labstack/echo/v4"
)

func ConfigueEcho(e *echo.Echo) {
	xormDB := common.ConfigureDatabase()

	e.Use(SetContextDB(xormDB))
}

func SetContextDB(xormDB *xorm.Engine) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			req = req.
				WithContext(context.WithValue(req.Context(), "xormDB", xormDB))
			c.SetRequest(req)

			if err := next(c); err != nil {
				return err
			}

			return nil
		}
	}
}
