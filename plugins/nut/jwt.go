package nut

import (
	"net/http"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// TOKEN token session key
	TOKEN = "token"
	// UID uid key
	UID = "uid"
	// CurrentUser current-user key
	CurrentUser = "currentUser"
	// IsAdmin is-admin key
	IsAdmin = "isAdmin"
)

var (
	_jwtKey    []byte
	_jwtMethod = crypto.SigningMethodHS512
)

//ValidateJwtToken check jwt
func ValidateJwtToken(buf []byte) (jwt.Claims, error) {
	tk, err := jws.ParseJWT(buf)
	if err != nil {
		return nil, err
	}
	if err = tk.Validate(_jwtKey, _jwtMethod); err != nil {
		return nil, err
	}
	return tk.Claims(), nil
}

func parseJwtToken(r *http.Request) (jwt.Claims, error) {
	tk, err := jws.ParseJWTFromRequest(r)
	if err != nil {
		return nil, err
	}
	if err = tk.Validate(_jwtKey, _jwtMethod); err != nil {
		return nil, err
	}
	return tk.Claims(), nil
}

//SumJwtToken create jwt token
func SumJwtToken(cm jws.Claims, exp time.Duration) ([]byte, error) {
	kid := uuid.New().String()
	now := time.Now()
	cm.SetNotBefore(now)
	cm.SetExpiration(now.Add(exp))
	cm.Set("kid", kid)
	//TODO using kid

	jt := jws.NewJWT(cm, _jwtMethod)
	return jt.Serialize(_jwtKey)
}

func getUserFromRequest(c *gin.Context) (*User, error) {
	lng := c.MustGet(LOCALE).(string)
	cm, err := parseJwtToken(c.Request)
	if err != nil {
		return nil, err
	}
	user, err := GetUserByUID(cm.Get(UID).(string))
	if err != nil {
		return nil, err
	}
	if !user.IsConfirm() {
		return nil, E(lng, "auth.errors.user.not-confirm")
	}
	if user.IsLock() {
		return nil, E(lng, "auth.errors.user.is-lock")
	}
	return user, nil
}

// CurrentUserMiddleware current-user middleware
func CurrentUserMiddleware(c *gin.Context) error {
	if user, err := getUserFromRequest(c); err == nil {
		c.Set(CurrentUser, user)
		c.Set(IsAdmin, Is(user.ID, RoleAdmin))
	}
	return nil
}

// MustSignInMiddleware must-sign-in middleware
func MustSignInMiddleware(c *gin.Context) error {
	if _, ok := c.MustGet(CurrentUser).(*User); ok {
		return nil
	}
	lng := c.MustGet(LOCALE).(string)
	return E(lng, "auth.errors.please-sign-in")
}

// MustAdminMiddleware must-admin middleware
func MustAdminMiddleware(c *gin.Context) error {
	if is, ok := c.MustGet(IsAdmin).(bool); ok && is {
		return nil
	}
	lng := c.MustGet(LOCALE).(string)
	return E(lng, "errors.not-allow")
}
