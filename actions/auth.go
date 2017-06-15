package actions

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bfosberry/gamesocialid/models"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/steam"
	"github.com/markbates/pop"
	"github.com/satori/go.uuid"
)

const (
	SessionKey  = "GSID_USER_SESSION"
	UserIDKey   = "user_id"
	LoggedInKey = "is_logged_in"
	AdminKey    = "is_admin"
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
	user_id_uuid := uuid.Nil
	user_id := c.Value("user_id")
	var user *models.User

	if user_id != nil {
		user_id_uuid = user_id.(uuid.UUID)
		user = &models.User{}
		if err := tx.Find(user, user_id_uuid); err != nil && !strings.Contains(err.Error(), "no rows in result set") {
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
			if err := loginUser(tx, c.Session(), user); err != nil {
				return err
			}
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
			if err := loginUser(tx, c.Session(), user); err != nil {
				return err
			}
		}
		if err := createCredential(tx, userData, user); err != nil {
			return err
		}
		c.Flash().Add("success", fmt.Sprintf("Successfully signed into %s", userData.Provider))
	}

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

func DecorateUserID(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		loggedIn := false
		sessionKey := c.Session().Get(SessionKey)
		if sessionKey != nil {
			sessionKeyStr := sessionKey.(string)
			userSession := &models.UserSession{}
			tx := c.Value("tx").(*pop.Connection)
			if err := tx.Where("session_key = ?", sessionKeyStr).First(userSession); err == nil {
				user := &models.User{}
				if err := tx.Find(user, userSession.UserID); err == nil {
					c.Logger().WithField("session_key", sessionKeyStr).Info("user_is_logged_in")
					c.Set(UserIDKey, userSession.UserID)
					c.Set(AdminKey, user.Admin)
					loggedIn = true
				}
			}
		}
		c.Set(LoggedInKey, loggedIn)
		return next(c)
	}
}

func Logout(c buffalo.Context) error {
	return logoutUser(c)
}

func loginUser(tx *pop.Connection, s *buffalo.Session, user *models.User) error {
	now := time.Now()
	userSession := &models.UserSession{
		UserID:      user.ID,
		SessionKey:  uuid.NewV4().String(),
		LoginTime:   &now,
		LastSeeTime: &now,
	}

	if err := tx.Create(userSession); err != nil {
		return err
	}
	s.Set(SessionKey, userSession.SessionKey)
	return nil
}

func logoutUser(c buffalo.Context) error {
	sessionKey := c.Session().Get(SessionKey)
	c.Session().Delete(SessionKey)
	if sessionKey != nil {
		sessionKeyStr := sessionKey.(string)
		tx := c.Value("tx").(*pop.Connection)
		sess := &models.UserSession{}
		if err := tx.Where("session_key = ?", sessionKeyStr).First(sess); err == nil {
			if err := tx.Destroy(sess); err != nil {
				return err
			}
		}
	}
	c.Set(UserIDKey, nil)

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
