package actions

import (
	"log"

	"github.com/bfosberry/gamesocialid/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/i18n"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr"
	"github.com/markbates/goth/gothic"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.Automatic(buffalo.Options{
			Env:         ENV,
			SessionName: "_gamesocialid_session",
		})
		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
		}
		if ENV == "production" {
			app.Host = "https://id.gamesocial.co"
		}
		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(middleware.CSRF)

		app.Use(middleware.PopTransaction(models.DB))

		app.Use(DecorateUserID)

		// Setup and use translations:
		var err error
		T, err = i18n.New(packr.NewBox("../locales"), "en-US")
		if err != nil {
			log.Fatal(err)
		}
		app.Use(T.Middleware())

		app.GET("/", HomeHandler)

		app.ServeFiles("/assets", packr.NewBox("../public/assets"))

		userGroup := app.Group("/")
		userGroup.Use(UserLoggedIn)
		userGroup.Resource("/users", UsersResource{&buffalo.BaseResource{}})
		userGroup.Resource("/credentials", CredentialsResource{&buffalo.BaseResource{}})

		adminGroup := userGroup.Group("/")
		adminGroup.Use(Admin)
		adminGroup.Resource("/user_sessions", UserSessionsResource{&buffalo.BaseResource{}})

		auth := app.Group("/auth")
		auth.GET("/logout", Logout)
		auth.GET("/{provider}", buffalo.WrapHandlerFunc(gothic.BeginAuthHandler))
		auth.GET("/{provider}/callback", AuthCallback)
	}

	return app
}
