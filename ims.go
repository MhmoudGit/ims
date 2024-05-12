package main

import (
	"context"
	"embed"
	"errors"
	"io"
	"log"
	"net/http"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tarqeem/ims/ent"
	. "github.com/tarqeem/ims/translate"
	. "github.com/tarqeem/template/utl"
)

//go:embed pages/*
var views embed.FS

const debug = true

var executor TemplateExecutor

var E *echo.Echo
var Client *ent.Client

func main() {
	// Views
	Views = views
	TemplateFuncs = template.FuncMap{
		"message": Message,
	}
	ts, err := GetTemplates()

	if err != nil {
		log.Fatal(err)
	}
	if debug {

		if err != nil {
			log.Fatal(err)
		}
		executor = DebugTemplateExecutor{}

	} else {
		executor = ts
	}

	// Database
	Client, err = ent.Open("sqlite3", "file:ent.db?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer Client.Close()
	// Run the auto migration tool.
	if err := Client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	// middleware
	E = echo.New()
	E.Use(session.Middleware(sessions.NewCookieStore([]byte("ssecret"))))
	E.Use(middleware.Logger())
	E.Use(middleware.Recover())
	E.Use(middleware.Static("static"))
	E.Renderer = &etmplt{templates: ts}

	// controllers
	E.GET("/", func(c echo.Context) error {
		_, err := authenticated(c)
		if err != nil && errors.Is(err, UserNotFound) {
			return c.Redirect(http.StatusTemporaryRedirect, LoginEnd)
		} else if err != nil {
			E.Logger.Errorf("authentication failed: %s", err.Error())
			return c.Render(http.StatusInternalServerError,
				"dashboard",
				&DashboardDTO{Err: err.Error()})
		}
		return c.Redirect(http.StatusTemporaryRedirect, DashboardEnd)
	})
	register()
	dashboard()
	login()

	createProject()
	projectView()
	createIssue()
	registerIssueRoutes()

	E.GET("/404", func(c echo.Context) error {
		return c.Render(http.StatusOK, "notFound", nil)
	})

	E.Logger.Fatal(E.Start("localhost:8000"))
}

type etmplt struct {
	templates *template.Template
}

func (t *etmplt) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if debug {
		return executor.ExecuteTemplate(w, name, data)
	}
	return t.templates.ExecuteTemplate(w, name, data)
}
