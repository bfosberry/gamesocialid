package actions

import (
	"github.com/bfosberry/gamesocialid/models"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
)

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (UserSession)
// DB Table: Plural (UserSessions)
// Resource: Plural (UserSessions)
// Path: Plural (/user_sessions)
// View Template Folder: Plural (/templates/userSessions/)

// UserSessionsResource is the resource for the user_session model
type UserSessionsResource struct {
	buffalo.Resource
}

// List gets all UserSessions. This function is mapped to the the path
// GET /user_sessions
func (v UserSessionsResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	userSessions := &models.UserSessions{}
	// You can order your list here. Just change
	err := tx.All(userSessions)
	// to:
	// err := tx.Order("(case when completed then 1 else 2 end) desc, lower([sort_parameter]) asc").All(userSessions)
	// Don't forget to change [sort_parameter] to the parameter of
	// your model, which should be used for sorting.
	if err != nil {
		return err
	}
	// Make user_sessions available inside the html template
	c.Set("userSessions", userSessions)
	return c.Render(200, r.HTML("user_sessions/index.html"))
}

// Show gets the data for one UserSession. This function is mapped to
// the path GET /user_sessions/{user_session_id}
func (v UserSessionsResource) Show(c buffalo.Context) error {
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty UserSession
	userSession := &models.UserSession{}
	// To find the UserSession the parameter user_session_id is used.
	err := tx.Find(userSession, c.Param("user_session_id"))
	if err != nil {
		return err
	}
	// Make userSession available inside the html template
	c.Set("userSession", userSession)
	return c.Render(200, r.HTML("user_sessions/show.html"))
}

// New renders the formular for creating a new user_session.
// This function is mapped to the path GET /user_sessions/new
func (v UserSessionsResource) New(c buffalo.Context) error {
	// Make userSession available inside the html template
	c.Set("userSession", &models.UserSession{})
	return c.Render(200, r.HTML("user_sessions/new.html"))
}

// Create adds a user_session to the DB. This function is mapped to the
// path POST /user_sessions
func (v UserSessionsResource) Create(c buffalo.Context) error {
	// Allocate an empty UserSession
	userSession := &models.UserSession{}
	// Bind userSession to the html form elements
	err := c.Bind(userSession)
	if err != nil {
		return err
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(userSession)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		// Make userSession available inside the html template
		c.Set("userSession", userSession)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		// Render again the new.html template that the user can
		// correct the input.
		return c.Render(422, r.HTML("user_sessions/new.html"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "UserSession was created successfully")
	// and redirect to the user_sessions index page
	return c.Redirect(302, "/user_sessions/%s", userSession.ID)
}

// Edit renders a edit formular for a user_session. This function is
// mapped to the path GET /user_sessions/{user_session_id}/edit
func (v UserSessionsResource) Edit(c buffalo.Context) error {
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty UserSession
	userSession := &models.UserSession{}
	err := tx.Find(userSession, c.Param("user_session_id"))
	if err != nil {
		return err
	}
	// Make userSession available inside the html template
	c.Set("userSession", userSession)
	return c.Render(200, r.HTML("user_sessions/edit.html"))
}

// Update changes a user_session in the DB. This function is mapped to
// the path PUT /user_sessions/{user_session_id}
func (v UserSessionsResource) Update(c buffalo.Context) error {
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty UserSession
	userSession := &models.UserSession{}
	err := tx.Find(userSession, c.Param("user_session_id"))
	if err != nil {
		return err
	}
	// Bind user_session to the html form elements
	err = c.Bind(userSession)
	if err != nil {
		return err
	}
	verrs, err := tx.ValidateAndUpdate(userSession)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		// Make userSession available inside the html template
		c.Set("userSession", userSession)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		// Render again the edit.html template that the user can
		// correct the input.
		return c.Render(422, r.HTML("user_sessions/edit.html"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "UserSession was updated successfully")
	// and redirect to the user_sessions index page
	return c.Redirect(302, "/user_sessions/%s", userSession.ID)
}

// Destroy deletes a user_session from the DB. This function is mapped
// to the path DELETE /user_sessions/{user_session_id}
func (v UserSessionsResource) Destroy(c buffalo.Context) error {
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty UserSession
	userSession := &models.UserSession{}
	// To find the UserSession the parameter user_session_id is used.
	err := tx.Find(userSession, c.Param("user_session_id"))
	if err != nil {
		return err
	}
	err = tx.Destroy(userSession)
	if err != nil {
		return err
	}
	// If there are no errors set a flash message
	c.Flash().Add("success", "UserSession was destroyed successfully")
	// Redirect to the user_sessions index page
	return c.Redirect(302, "/user_sessions")
}
