package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tarqeem/ims/ent"
	"github.com/tarqeem/ims/ent/user"
)

type DashboardDTO struct {
	User     *ent.User
	Projects []*ent.Project
	Err      string
}

var DashboardEnd string = "/dashboard"

func dashboard() {
	E.GET(DashboardEnd, func(c echo.Context) error {
		p := "dashboard"
		data := DashboardDTO{}
		id, err := authenticated(c)
		if err != nil && errors.Is(err, UserNotFound) {
			return c.Redirect(http.StatusTemporaryRedirect, LoginEnd)
		} else if err != nil {
			E.Logger.Errorf("authentication failed: %s", err.Error())
			return c.Render(http.StatusInternalServerError,
				"fail",
				&DashboardDTO{Err: err.Error()})
		}

		u, err := Client.User.Query().Where(user.ID(id)).Only(c.Request().Context())
		if err != nil {
			E.Logger.Error(err)
			return c.Render(http.StatusInternalServerError, "fail",
				&DashboardDTO{Err: err.Error()})
		}
		data.User = u
		// memberProjects, err := u.
		// 	QueryProjects().
		// 	All(context.Background())
		// if err != nil {
		// 	return c.Render(http.StatusInternalServerError, "fail",
		// 		&DashboardDTO{Err: err.Error()})
		// }
		// data.Projects = append(data.Projects, memberProjects...)

		coordinatorProjects, err := u.
			QueryCoordinatorOfProject().
			All(context.Background())
		if err != nil {
			return c.Render(http.StatusInternalServerError, "fail",
				&DashboardDTO{Err: err.Error()})
		}
		data.Projects = append(data.Projects, coordinatorProjects...)

		return c.Render(http.StatusOK, p, data)
	})

}
