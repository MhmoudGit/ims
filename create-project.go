package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tarqeem/ims/ent"
	"github.com/tarqeem/ims/ent/project"
	"github.com/tarqeem/ims/ent/user"
)

type SuccessData struct {
	Msg string
}

var CreateProjectEnd = "/create-project"
var UpdateProjectEnd = "/edit-project"

type CreateProjectDTO struct {
	Name                string   `form:"name"`
	Owner               string   `form:"owner"`
	Location            string   `form:"location"`
	Type                string   `form:"type"`
	DeliveryStrategy    string   `form:"deliveryStrategy"`
	CurrentState        string   `form:"currentState"`
	ContractingStrategy string   `form:"contractingStrategy"`
	DollarValue         int      `form:"dollarValue"`
	ExecutionLocation   string   `form:"executionLocation"`
	Tlsp                int      `form:"tlsp"` // No. of top level scope packages
	Jvp                 int      `form:"jvp"`  // No. of joint-venture partners
	Ish                 int      `form:"ish"`  // No. of involved stackeholders
	ProjectNature       string   `form:"projectNature"`
	Leader              string   `form:"leader"`
	Members             []string `form:"members[]"`
	PageTitle           string
	Err                 string
}

func createProject() {
	E.GET("/create-project", func(c echo.Context) error {
		return c.Render(http.StatusOK, "create-project", nil)
	})

	E.POST(CreateProjectEnd, func(c echo.Context) error {

		p := "success"

		r := &CreateProjectDTO{}
		if err := c.Bind(r); err != nil {
			return c.String(http.StatusBadRequest, "bad request")
		}

		proj, err := Client.Project.Create().
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
			fmt.Print("Error creating project: " + err.Error())
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

		err = c.Render(http.StatusOK, p, SuccessData{Msg: "created"})
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusOK, DashboardEnd)
	})
}

func addCoordinator(c echo.Context, proj *ent.Project) error {
	u, err := getCurrentUserID(c)
	if err != nil {
		return err
	}
	if _, err := proj.Update().AddCoordinator(u).Save(c.Request().Context()); err != nil {
		return err
	}
	return nil
}

func addLeader(c echo.Context, proj *ent.Project, leader string) error {
	u, err := ensureUser(c, leader)
	if err != nil {
		return err
	}
	if _, err = proj.Update().AddLeader(u).Save(c.Request().Context()); err != nil {
		return err
	}
	return nil
}

func addMembers(c echo.Context, proj *ent.Project, members []string) error {
	for _, member := range members {
		u, err := ensureUser(c, member)
		if err != nil {
			return err
		}
		if _, err = u.Update().AddProjects(proj).Save(c.Request().Context()); err != nil {
			return err
		}
	}

	return nil
}

func ensureUser(c echo.Context, email string) (*ent.User, error) {
	u, err := Client.User.
		Query().
		Where(user.EmailEQ(email)). // Query using the unique email.
		Only(context.Background())
	if ent.IsNotFound(err) {
		u, err = Client.User.Create().
			SetEmail(email).
			SetUsername(fmt.Sprintf("user-%s", randSeq(4))).
			SetIsActive(false).
			SetType(user.TypeMember).Save(c.Request().Context())
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return u, nil
}
