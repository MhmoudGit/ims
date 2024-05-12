package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tarqeem/ims/ent"
	"github.com/tarqeem/ims/ent/project"
)

type ProjectDTO struct {
	Project *ent.Project
	Issues  []*ent.Issue
	Err     string
}

var ProjectEnd = "/project"

func projectView() {
	E.GET(ProjectEnd, func(c echo.Context) error {

		prName := c.QueryParam("name")
		data := ProjectDTO{}

		projectObject, err := Client.Project.Query().
			Where(project.NameEQ(prName)).
			Only(context.Background())

		if err != nil {
			return c.Render(http.StatusInternalServerError, "fail", nil)
		}
		data.Project = projectObject

		issues, err := Client.Project.
			QueryIssues(projectObject).
			All(context.Background())
		if err != nil {
			return c.Render(http.StatusInternalServerError, "fail", nil)
		}

		data.Issues = issues

		sess, err := cookie(c)
		if err != nil {
			E.Logger.Errorf("ent: %s", err.Error())
			return c.Render(http.StatusInternalServerError, "fail", nil)
		}

		// sess.Values["auth"] = "true"
		sess.Values["project_name"] = prName

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			E.Logger.Errorf("ent: %s", err.Error())
			return c.Render(http.StatusInternalServerError, "fail", nil)
		}

		return c.Render(http.StatusOK, "project", data)
	})

	E.DELETE(ProjectEnd, func(c echo.Context) error {
		prID := c.QueryParam("id")
		// Convert string to int
		ID, err := strconv.Atoi(prID)
		if err != nil {
			return c.Render(http.StatusInternalServerError, "fail", nil)
		}
		data := ProjectDTO{}

		projectObject, err := Client.Project.Query().
			Where(project.IDEQ(ID)).
			Only(context.Background())

		if err != nil {
			return c.Render(http.StatusInternalServerError, "fail", nil)
		}
		data.Project = projectObject
		return c.JSON(200, data.Project)
	})
}
