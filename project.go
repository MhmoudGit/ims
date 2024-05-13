package main

import (
	"context"
	"fmt"
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
		err = Client.Project.DeleteOneID(ID).Exec(context.Background())
		if err != nil {
			return c.Render(http.StatusInternalServerError, "fail", nil)
		}

		return nil
	})

	E.GET(ProjectEnd+"-edit", func(c echo.Context) error {
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
		leaders := projectObject.QueryLeader().AllX(context.Background())
		members := projectObject.QueryMembers().AllX(context.Background())
		data.Project = projectObject
		data.Project.Edges.Members = members
		data.Project.Edges.Leader = leaders

		return c.Render(http.StatusOK, "edit-project", data.Project)
	})

	E.POST(UpdateProjectEnd+"/:id", func(c echo.Context) error {
		prID := c.Param("id")
		// Convert string to int
		ID, err := strconv.Atoi(prID)
		if err != nil {
			return c.String(http.StatusBadRequest, "bad request")
		}

		p := "success"

		r := &CreateProjectDTO{}
		if err := c.Bind(r); err != nil {
			return c.String(http.StatusBadRequest, "bad request")
		}

		proj, err := Client.Project.UpdateOneID(ID).
			SetName(r.Name).
			SetOwner(r.Owner).
			SetLocation(r.Location).
			SetType(project.Type(r.Type)).
			SetProjectNature(project.ProjectNature(r.ProjectNature)).
			SetDeliveryStrategies(r.DeliveryStrategy).
			SetState(r.CurrentState).
			SetContractingStrategies(r.ContractingStrategy).
			SetDollarValue(r.DollarValue).
			SetExecutionLocation(r.ExecutionLocation).
			SetTlsp(r.Tlsp).
			SetJvp(r.Jvp).
			SetIsh(r.Ish).
			Save(c.Request().Context())

		if err != nil {
			fmt.Print("Error updateing project: " + err.Error())
			return c.Render(http.StatusInternalServerError, "fail",
				&CreateProjectDTO{Err: err.Error()})
		}
		if err = addCoordinator(c, proj); err != nil {
			fmt.Print("Error adding coordinator: " + err.Error())
			return c.Render(http.StatusInternalServerError, "fail",
				&CreateProjectDTO{Err: err.Error()})
		}
		if err = addLeader(c, proj, r.Leader); err != nil {
			fmt.Print("Error adding leader: " + err.Error())
			return c.Render(http.StatusInternalServerError, "fail",
				&CreateProjectDTO{Err: err.Error()})
		}
		r.Members = append(r.Members, r.Leader)
		if err = addMembers(c, proj, r.Members); err != nil {
			fmt.Print("Error adding members: " + err.Error())
			return c.Render(http.StatusInternalServerError, "fail",
				&CreateProjectDTO{Err: err.Error()})
		}

		err = c.Render(http.StatusOK, p, SuccessData{Msg: "updated"})
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusOK, DashboardEnd)
	})
}
