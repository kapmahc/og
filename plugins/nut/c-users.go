package nut

import (
	"net/http"
	"time"

	"github.com/SermoDigital/jose/jws"
	"github.com/gin-gonic/gin"
)

type fmSignUp struct {
	Name                 string `json:"name" binding:"required,max=255"`
	Email                string `json:"email" binding:"required,email"`
	Password             string `json:"password" binding:"min=6,max=32"`
	PasswordConfirmation string `json:"passwordConfirmation" binding:"eqfield=Password"`
}

func postUsersSignUp(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)
	var fm fmSignUp
	if err := c.BindJSON(&fm); err != nil {
		return err
	}

	var count int
	if err := DB().
		Model(&User{}).
		Where("email = ?", fm.Email).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return E(l, "auth.errors.user.email-already-exists")
	}

	user, err := AddEmailUser(fm.Name, fm.Email, fm.Password)
	if err != nil {
		return err
	}

	AddLog(user.ID, c.ClientIP(), T(l, "auth.logs.user.sign-up"))
	sendEmail(l, c.Request, user, actConfirm)

	c.JSON(http.StatusOK, gin.H{})
	return nil
}

type fmSignIn struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"rememberMe"`
}

func postUsersSignIn(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)
	var fm fmSignIn
	if err := c.BindJSON(&fm); err != nil {
		return err
	}

	user, err := SignIn(l, fm.Email, fm.Password, c.ClientIP())
	if err != nil {
		return err
	}

	cm := jws.Claims{}
	cm.Set(UID, user.UID)
	cm.Set("name", user.Name)
	cm.Set("admin", Is(user.ID, RoleAdmin))
	tkn, err := SumJwtToken(cm, time.Hour*24*7)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"token": string(tkn),
	})
	return nil
}

type fmEmail struct {
	Email string `json:"email" binding:"required,email"`
}

func getUsersConfirm(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)
	token := c.Param("token")
	user, err := parseToken(l, token, actConfirm)
	if err != nil {
		return err
	}
	if user.IsConfirm() {
		return E(l, "auth.errors.user.already-confirm")
	}
	DB().Model(user).Update("confirmed_at", time.Now())
	AddLog(user.ID, c.ClientIP(), T(l, "auth.logs.user.confirm"))

	c.Redirect(http.StatusFound, _signInURL(c.Request))
	return nil
}

func postUsersConfirm(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)
	var fm fmEmail
	if err := c.BindJSON(&fm); err != nil {
		return err
	}
	user, err := GetByEmail(fm.Email)
	if err != nil {
		return err
	}

	if user.IsConfirm() {
		return E(l, "auth.errors.user.already-confirm")
	}

	sendEmail(l, c.Request, user, actConfirm)

	c.JSON(http.StatusOK, gin.H{})
	return nil
}

func getUsersUnlock(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)
	token := c.Param("token")
	user, err := parseToken(l, token, actUnlock)
	if err != nil {
		return err
	}
	if !user.IsLock() {
		return E(l, "auth.errors.user.not-lock")
	}

	DB().Model(user).Update(map[string]interface{}{"locked_at": nil})
	AddLog(user.ID, c.ClientIP(), T(l, "auth.logs.user.unlock"))

	c.Redirect(http.StatusFound, _signInURL(c.Request))
	return nil
}

func postUsersUnlock(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)

	var fm fmEmail
	if err := c.BindJSON(&fm); err != nil {
		return err
	}
	user, err := GetByEmail(fm.Email)
	if err != nil {
		return err
	}
	if !user.IsLock() {
		return E(l, "auth.errors.user.not-lock")
	}
	sendEmail(l, c.Request, user, actUnlock)
	c.JSON(http.StatusOK, gin.H{})
	return nil
}

func postUsersForgotPassword(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)
	var fm fmEmail
	if err := c.BindJSON(&fm); err != nil {
		return err
	}
	var user *User
	user, err := GetByEmail(fm.Email)
	if err != nil {
		return err
	}
	sendEmail(l, c.Request, user, actResetPassword)

	c.JSON(http.StatusOK, gin.H{})
	return nil
}

type fmResetPassword struct {
	Token                string `json:"token" binding:"required"`
	Password             string `json:"password" binding:"min=6,max=32"`
	PasswordConfirmation string `json:"passwordConfirmation" binding:"eqfield=Password"`
}

func postUsersResetPassword(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)

	var fm fmResetPassword
	if err := c.BindJSON(&fm); err != nil {
		return err
	}
	user, err := parseToken(l, fm.Token, actResetPassword)
	if err != nil {
		return err
	}
	DB().Model(user).Update("password", Sum([]byte(fm.Password)))
	AddLog(user.ID, c.ClientIP(), T(l, "auth.logs.user.reset-password"))
	c.JSON(http.StatusOK, gin.H{})
	return nil
}

func deleteUsersSignOut(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)
	user := c.MustGet(CurrentUser).(*User)
	AddLog(user.ID, c.ClientIP(), T(l, "auth.logs.user.sign-out"))
	c.JSON(http.StatusOK, gin.H{})
	return nil
}

func getUsersInfo(c *gin.Context) error {
	user := c.MustGet(CurrentUser).(*User)
	c.JSON(http.StatusOK, gin.H{"name": user.Name, "email": user.Email})
	return nil
}

type fmInfo struct {
	Name string `json:"name" binding:"required,max=255"`
}

func postUsersInfo(c *gin.Context) error {
	user := c.MustGet(CurrentUser).(*User)
	var fm fmInfo
	if err := c.BindJSON(&fm); err != nil {
		return err
	}

	if err := DB().Model(user).Updates(map[string]interface{}{
		"name": fm.Name,
	}).Error; err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{})
	return nil
}

type fmChangePassword struct {
	CurrentPassword      string `json:"currentPassword" binding:"required"`
	NewPassword          string `json:"newPassword" binding:"min=6,max=32"`
	PasswordConfirmation string `json:"passwordConfirmation" binding:"eqfield=NewPassword"`
}

func postUsersChangePassword(c *gin.Context) error {
	l := c.MustGet(LOCALE).(string)

	user := c.MustGet(CurrentUser).(*User)
	var fm fmChangePassword
	if err := c.BindJSON(&fm); err != nil {
		return err
	}
	if !Chk([]byte(fm.CurrentPassword), user.Password) {
		return E(l, "auth.errors.bad-password")
	}
	if err := DB().Model(user).
		Update("password", Sum([]byte(fm.NewPassword))).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{})
	return nil
}

func getUsersLogs(c *gin.Context) error {
	user := c.MustGet(CurrentUser).(*User)
	var logs []Log
	if err := DB().
		Select([]string{"id", "ip", "message", "created_at"}).
		Where("user_id = ?", user.ID).
		Order("id DESC").Limit(120).
		Find(&logs).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, logs)
	return nil
}

func indexUsers(c *gin.Context) error {
	var users []User
	if err := DB().
		Select([]string{"name", "logo", "home"}).
		Order("last_sign_in_at DESC").
		Find(&users).Error; err != nil {
		return err
	}
	c.JSON(http.StatusOK, users)
	return nil
}

func _signInURL(r *http.Request) string {
	return r.URL.Host + "/users/sign-in"
}

func init() {

	ung := Router().Group("/users")
	ung.GET("/", Wrap(indexUsers))
	ung.POST("/sign-in", Wrap(postUsersSignIn))
	ung.POST("/sign-up", Wrap(postUsersSignUp))
	ung.POST("/confirm", Wrap(postUsersConfirm))
	ung.GET("/confirm/:token", Wrap(getUsersConfirm))
	ung.POST("/unlock", Wrap(postUsersUnlock))
	ung.GET("/unlock/:token", Wrap(getUsersUnlock))
	ung.POST("/forgot-password", Wrap(postUsersForgotPassword))
	ung.POST("/reset-password", Wrap(postUsersResetPassword))

	umg := Router().Group("/users", Wrap(MustSignInMiddleware))
	umg.GET("/info", Wrap(getUsersInfo))
	umg.POST("/info", Wrap(postUsersInfo))
	umg.GET("/logs", Wrap(getUsersLogs))
	umg.POST("/change-password", Wrap(postUsersChangePassword))
	umg.DELETE("/sign-out", Wrap(deleteUsersSignOut))

}
