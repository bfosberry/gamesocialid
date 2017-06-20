package actions

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bfosberry/gamesocialid/models"
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/battlenet"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/steam"
	"github.com/markbates/goth/providers/twitch"
	"github.com/markbates/goth/providers/twitter"
	"github.com/markbates/pop"
	"github.com/satori/go.uuid"
)

const (
	SessionKey  = "GSID_USER_SESSION"
	UserIDKey   = "user_id"
	LoggedInKey = "is_logged_in"
	AdminKey    = "is_admin"
)

var (
	ErrUnauthorized    = errors.New("Forbidden Brah")
	ErrUnauthenticated = errors.New("Login Brah")
	ErrNotFound        = errors.New("Missing Brah")
)

func init() {
	gothic.Store = App().SessionStore

	goth.UseProviders(
		steam.New(os.Getenv("STEAM_API_KEY"), fmt.Sprintf("%s%s", App().Host, "/auth/steam/callback")),
		twitter.New(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), fmt.Sprintf("%s%s", App().Host, "/auth/twitter/callback")),
		facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"), fmt.Sprintf("%s%s", App().Host, "/auth/facebook/callback")),
		twitch.New(os.Getenv("TWITCH_KEY"), os.Getenv("TWITCH_SECRET"), fmt.Sprintf("%s%s", App().Host, "/auth/twitch/callback")),
		battlenet.New(os.Getenv("BATTLENET_KEY"), os.Getenv("BATTLENET_SECRET"), "https://id.gamesocial.co/auth/battlenet/callback"), // fmt.Sprintf("%s%s", App().Host, "/auth/battlenet/callback")),
	)
}

func AuthCallback(c buffalo.Context) error {
	userData, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	c.Logger().Infof("UserData is %+v\n", userData)
	if err != nil {
		return c.Error(401, err)
	}

	flashMsg := fmt.Sprintf("Logged in via %s", userData.Provider)
	redirectURL := "/"

	user, err := currentUser(c)
	if err != nil {
		return err
	}

	tx := c.Value("tx").(*pop.Connection)
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
			if err := tx.Destroy(credential); err != nil {
				return err
			}

			if err := createCredential(tx, userData, user); err != nil {
				return err
			}
		}
		c.Flash().Add("success", flashMsg)
	} else {
		created := false
		if user == nil {
			created = true
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
		flashMsg := fmt.Sprintf("Logged in via %s", userData.Provider)
		if created {
			flashMsg = fmt.Sprintf("Successfully created user via %s", userData.Provider)
			redirectURL = fmt.Sprintf("/users/%s/edit", user.ID.String())
		}
		c.Flash().Add("success", flashMsg)

	}
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func createCredential(tx *pop.Connection, userData goth.User, user *models.User) error {
	credential := &models.Credential{}
	credential.Provider = userData.Provider
	credential.Uid = userData.UserID
	credential.Name = userData.Name
	credential.Nickname = userData.NickName
	credential.Email = userData.Email
	credential.ImageUrl = userData.AvatarURL
	credential.ProfileUrl = profileURL(userData.Provider, userData.UserID, userData.Name)
	credential.AccessToken = userData.AccessToken
	credential.RefreshToken = userData.RefreshToken
	credential.TokenExpiry = userData.ExpiresAt.String()
	credential.UserID = user.ID
	return tx.Create(credential)
}

func profileURL(provider, uid, name string) string {
	switch provider {
	case "steam":
		return steamProfileURL(uid)
	case "twitch":
		return twitchProfileURL(name)
	}
	return ""

}

func steamProfileURL(uid string) string {
	return fmt.Sprintf("http://steamcommunity.com/profiles/%s", uid)
}

func twitchProfileURL(name string) string {
	return fmt.Sprintf("https://www.twitch.tv/%s", name)
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

func Admin(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if err := requireAdmin(c); err != nil {
			return err
		}
		return next(c)
	}
}

func UserLoggedIn(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if err := requireLoggedIn(c); err != nil {
			return err
		}
		return next(c)
	}
}

func requireAdmin(c buffalo.Context) error {
	if !isAdmin(c) {
		return c.Error(403, ErrUnauthorized)
	}
	return nil
}

func requireLoggedIn(c buffalo.Context) error {
	if !isLoggedIn(c) {
		return c.Error(401, ErrUnauthenticated)
	}
	return nil
}

func isAdmin(c buffalo.Context) bool {
	isAdmin := c.Value(AdminKey)
	if isAdmin != nil && isAdmin.(bool) {
		return true
	}
	return false
}

func isLoggedIn(c buffalo.Context) bool {
	isLoggedIn := c.Value(LoggedInKey)
	if isLoggedIn != nil && isLoggedIn.(bool) {
		return true
	}
	return false
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
	c.Flash().Add("success", "Logged out")
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func currentUser(c buffalo.Context) (*models.User, error) {
	tx := c.Value("tx").(*pop.Connection)
	userID := currentUserID(c)
	if userID != uuid.Nil {
		user := &models.User{}
		if err := tx.Find(user, userID); err != nil && !strings.Contains(err.Error(), "no rows in result set") {
			return nil, err
		}
		return user, nil
	}
	return nil, nil
}

func currentUserID(c buffalo.Context) uuid.UUID {
	userID := c.Value(UserIDKey)
	if userID != nil {
		return userID.(uuid.UUID)
	}
	return uuid.Nil
}
