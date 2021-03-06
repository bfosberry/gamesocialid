package actions

import (
	"strings"

	"github.com/bfosberry/gamesocialid/models"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/satori/go.uuid"
)

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (User)
// DB Table: Plural (Users)
// Resource: Plural (Users)
// Path: Plural (/users)
// View Template Folder: Plural (/templates/users/)

// UsersResource is the resource for the user model
type UsersResource struct {
	buffalo.Resource
}

// List gets all Users. This function is mapped to the the path
// GET /users
func (v UsersResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	if err := requireAdmin(c); err != nil {
		return err
	}
	tx := c.Value("tx").(*pop.Connection)
	users := &models.Users{}
	// You can order your list here. Just change
	err := tx.All(users)
	// to:
	// err := tx.Order("(case when completed then 1 else 2 end) desc, lower([sort_parameter]) asc").All(users)
	// Don't forget to change [sort_parameter] to the parameter of
	// your model, which should be used for sorting.
	if err != nil {
		return err
	}
	// Make users available inside the html template
	c.Set("users", users)
	return c.Render(200, r.HTML("users/index.html"))
}

// Show gets the data for one User. This function is mapped to
// the path GET /users/{user_id}
func (v UsersResource) Show(c buffalo.Context) error {
	userID := currentUserID(c)
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty User
	user := &models.User{}
	// To find the User the parameter user_id is used.
	userParam := c.Param("user_id")
	userUUID, err := uuid.FromString(userParam)
	if err != nil {
		if err := tx.Where("username = ?", userParam).First(user); err != nil {
			return err
		}
	} else {
		if err := tx.Find(user, userUUID); err != nil {
			return err
		}
	}

	if user.Visibility == false && user.ID != userID {
		return c.Error(404, ErrNotFound)
	}

	credentials := &models.Credentials{}
	// You can order your list here. Just change
	if err = tx.Where("user_id = ?", user.ID).All(credentials); err != nil {
		return err
	}
	// Make credentials available inside the html template
	c.Set("credentials", credentials)
	// Make user available inside the html template
	c.Set("user", user)
	c.Set("owner", user.ID == userID)
	return c.Render(200, r.HTML("users/show.html"))
}

// New renders the formular for creating a new user.
// This function is mapped to the path GET /users/new
func (v UsersResource) New(c buffalo.Context) error {
	if err := requireAdmin(c); err != nil {
		return err
	}
	// Make user available inside the html template
	c.Set("user", &models.User{})
	return c.Render(200, r.HTML("users/new.html"))
}

// Create adds a user to the DB. This function is mapped to the
// path POST /users
func (v UsersResource) Create(c buffalo.Context) error {
	if err := requireAdmin(c); err != nil {
		return err
	} // Allocate an empty User
	user := &models.User{}
	// Bind user to the html form elements
	err := c.Bind(user)
	if err != nil {
		return err
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(user)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		// Make user available inside the html template
		c.Set("user", user)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		// Render again the new.html template that the user can
		// correct the input.
		return c.Render(422, r.HTML("users/new.html"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "User was created successfully")
	// and redirect to the users index page
	return c.Redirect(302, "/users/%s", user.ID)
}

// Edit renders a edit formular for a user. This function is
// mapped to the path GET /users/{user_id}/edit
func (v UsersResource) Edit(c buffalo.Context) error {
	if err := requireLoggedIn(c); err != nil {
		return err
	}
	userID := currentUserID(c)
	if c.Param("user_id") != userID.String() {
		return c.Error(404, ErrNotFound)
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty User
	user := &models.User{}
	admin := isAdmin(c)
	if !admin {
		user.Admin = false
	}
	err := tx.Find(user, c.Param("user_id"))
	if err != nil {
		return err
	}
	// Make user available inside the html template
	c.Set("user", user)
	return c.Render(200, r.HTML("users/edit.html"))
}

// Update changes a user in the DB. This function is mapped to
// the path PUT /users/{user_id}
func (v UsersResource) Update(c buffalo.Context) error {
	if err := requireLoggedIn(c); err != nil {
		return err
	}
	userID := currentUserID(c)
	if c.Param("user_id") != userID.String() {
		return c.Error(404, ErrNotFound)
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty User
	user := &models.User{}
	err := tx.Find(user, c.Param("user_id"))
	if err != nil {
		return err
	}
	// Bind user to the html form elements
	err = c.Bind(user)
	if err != nil {
		return err
	}
	verrs, err := tx.ValidateAndUpdate(user)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		// Make user available inside the html template
		c.Set("user", user)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		// Render again the edit.html template that the user can
		// correct the input.
		return c.Render(422, r.HTML("users/edit.html"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "User was updated successfully")
	// and redirect to the users index page
	return c.Redirect(302, "/users/%s", user.ID)
}

// Destroy deletes a user from the DB. This function is mapped
// to the path DELETE /users/{user_id}
func (v UsersResource) Destroy(c buffalo.Context) error {
	if err := requireLoggedIn(c); err != nil {
		return err
	}
	userID := currentUserID(c)
	if c.Param("user_id") != userID.String() {
		return c.Error(404, ErrNotFound)
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty User
	user := &models.User{}
	// To find the User the parameter user_id is used.
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return err
	}

	if err := tx.Destroy(user); err != nil {
		return err
	}

	sessions := &models.UserSessions{}
	if err := tx.Where("user_id = ?", user.ID).All(sessions); err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return err
	}
	for _, s := range []models.UserSession(*sessions) {
		if err := tx.Destroy(&s); err != nil {
			return err
		}
	}
	credentials := &models.Credentials{}
	if err := tx.Where("user_id = ?", user.ID).All(credentials); err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return err
	}
	for _, cred := range []models.Credential(*credentials) {
		if err := tx.Destroy(&cred); err != nil {
			return err
		}
	}

	// If there are no errors set a flash message
	c.Flash().Add("success", "User was destroyed successfully")
	// Redirect to the users index page
	return c.Redirect(302, "/users")
}
