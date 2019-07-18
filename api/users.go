package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"

	"github.com/udovin/solve/models"
)

func (v *View) CreateUser(c echo.Context) error {
	var userData struct {
		Login    string `json:""`
		Email    string `json:""`
		Password string `json:""`
	}
	if err := c.Bind(&userData); err != nil {
		c.Logger().Error(err)
		return err
	}
	user := models.User{
		Login: userData.Login,
	}
	if err := user.SetPassword(
		userData.Password, v.app.PasswordSalt,
	); err != nil {
		c.Logger().Error(err)
		return err
	}
	if err := v.app.UserStore.Create(&user); err != nil {
		c.Logger().Error(err)
		return err
	}
	return c.JSON(http.StatusOK, user)
}

func (v *View) GetUser(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("UserID"), 10, 60)
	if err != nil {
		c.Logger().Error(err)
		return err
	}
	user, ok := v.app.UserStore.Get(userID)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, user)
}

func (v *View) UpdateUser(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("UserID"), 10, 60)
	if err != nil {
		c.Logger().Error(err)
		return err
	}
	user, ok := v.app.UserStore.Get(userID)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	var userData struct {
		Password *string `json:""`
	}
	if err := c.Bind(&userData); err != nil {
		c.Logger().Error(err)
		return c.NoContent(http.StatusBadRequest)
	}
	if userData.Password != nil {
		if err := user.SetPassword(
			*userData.Password, v.app.PasswordSalt,
		); err != nil {
			c.Logger().Error(err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	if err := v.app.UserStore.Update(&user); err != nil {
		c.Logger().Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, user)
}

func (v *View) DeleteUser(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}