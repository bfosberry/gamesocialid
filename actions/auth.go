package actions

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bfosberry/gamesocialid/models"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/steam"
	"github.com/markbates/pop"
)

func init() {
	gothic.Store = App().SessionStore

	goth.UseProviders(
		steam.New(os.Getenv("STEAM_API_KEY"), fmt.Sprintf("%s%s", App().Host, "/auth/steam/callback")),
	)
}

func AuthCallback(c buffalo.Context) error {
	userData, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.Error(401, err)
	}

	tx := c.Value("tx").(*pop.Connection)
	user_id_str := ""
	user_id := c.Value("user_id")
	var user *models.User

	if user_id != nil {
		user_id_str = user_id.(string)
		user = &models.User{}
		if err := tx.Find(user, user_id_str); err != nil && !strings.Contains(err.Error(), "no rows in result set") {
			return err
		}
	}

	credential := &models.Credential{}
	err = tx.Where("provider = ?", userData.Provider).Where("uid = ?", userData.UserID).First(credential)
	if err == nil {

		if user == nil {
			user = &models.User{}
			if err := tx.Find(user, credential.UserID); err != nil {
				return err
			}
			// TODO handle login for user
		} else if user.ID != credential.UserID {
			credential.UserID = user.ID
			if err := tx.Save(credential); err != nil {
				return err
			}
		}
	} else {
		if user == nil {
			user = &models.User{
				Username:   userData.NickName,
				RealName:   userData.Name,
				AvatarUrl:  userData.AvatarURL,
				Email:      userData.Email,
				Visibility: true,
				Admin:      false,
			}
			if err := tx.Create(user); err != nil {
				return err
			}
		}
		if err := createCredential(tx, userData, user); err != nil {
			return err
		}

		// if user is logged in associate credential with user
		// if user is not logged in create a new user and associate it with this credential
		c.Flash().Add("success", fmt.Sprintf("Successfully signed into %s", userData.Provider))
	}
	// TODO UserID       uuid.UUID `json:"user_id" db:"user_id"`

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func createCredential(tx *pop.Connection, userData goth.User, user *models.User) error {
	credential := &models.Credential{}
	credential.Provider = userData.Provider
	credential.Uid = userData.UserID
	credential.Name = userData.Name
	credential.Nickname = userData.NickName
	credential.Email = userData.Email
	credential.ImageUrl = userData.AvatarURL
	credential.ProfileUrl = "unknown"
	credential.AccessToken = userData.AccessToken
	credential.RefreshToken = userData.RefreshToken
	credential.TokenExpiry = userData.ExpiresAt.String()
	credential.UserID = user.ID
	return tx.Create(credential)
}
