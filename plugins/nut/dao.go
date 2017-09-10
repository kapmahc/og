package nut

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// SignIn set sign-in info
func SignIn(lang, email, password, ip string) (*User, error) {
	user, err := GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if Chk([]byte(password), user.Password) {
		AddLog(user.ID, ip, T(lang, "nut.logs.user.sign-in.failed"))
		return nil, E(lang, "nut.errors.user.email-password-not-match")
	}

	if !user.IsConfirm() {
		return nil, E(lang, "nut.errors.user.not-confirm")
	}

	if user.IsLock() {
		return nil, E(lang, "nut.errors.user.is-lock")
	}

	AddLog(user.ID, ip, T(lang, "nut.logs.user.sign-in.success"))
	user.SignInCount++
	user.LastSignInAt = user.CurrentSignInAt
	user.LastSignInIP = user.CurrentSignInIP
	now := time.Now()
	user.CurrentSignInAt = &now
	user.CurrentSignInIP = ip
	if err = DB().Model(user).Updates(map[string]interface{}{
		"last_sign_in_at":    user.LastSignInAt,
		"last_sign_in_ip":    user.LastSignInIP,
		"current_sign_in_at": user.CurrentSignInAt,
		"current_sign_in_ip": user.CurrentSignInIP,
		"sign_in_count":      user.SignInCount,
	}).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByUID get user by uid
func GetUserByUID(uid string) (*User, error) {
	var u User
	err := DB().Where("uid = ?", uid).First(&u).Error
	return &u, err
}

// GetByEmail get user by email
func GetByEmail(email string) (*User, error) {
	var user User
	err := DB().
		Where("provider_type = ? AND provider_id = ?", UserTypeEmail, email).
		First(&user).Error
	return &user, err
}

// AddLog add log
func AddLog(user uint, ip, message string) {
	if err := DB().Create(&Log{UserID: user, IP: ip, Message: message}).Error; err != nil {
		log.Error(err)
	}
}

// AddEmailUser add email user
func AddEmailUser(name, email, password string) (*User, error) {

	user := User{
		Email:           email,
		Password:        Sum([]byte(password)),
		Name:            name,
		ProviderType:    UserTypeEmail,
		ProviderID:      email,
		Home:            "/users",
		LastSignInIP:    "0.0.0.0",
		CurrentSignInIP: "0.0.0.0",
	}
	user.SetUID()
	user.SetGravatarLogo()
	user.Home = fmt.Sprintf("/users/%s", user.UID)

	err := DB().Create(&user).Error
	return &user, err
}

// Authority get roles
func Authority(user uint, rty string, rid uint) []string {
	var items []Role
	if err := DB().
		Where("resource_type = ? AND resource_id = ?", rty, rid).
		Find(&items).Error; err != nil {
		log.Error(err)
	}
	var roles []string
	for _, r := range items {
		var pm Policy
		if err := DB().
			Where("role_id = ? AND user_id = ? ", r.ID, user).
			First(&pm).Error; err != nil {
			log.Error(err)
			continue
		}
		if pm.Enable() {
			roles = append(roles, r.Name)
		}
	}
	return roles
}

//Is is role ?
func Is(user uint, names ...string) bool {
	for _, name := range names {
		if Can(user, name, "-", 0) {
			return true
		}
	}
	return false
}

//Can can?
func Can(user uint, name string, rty string, rid uint) bool {
	var r Role
	if DB().
		Where("name = ? AND resource_type = ? AND resource_id = ?", name, rty, rid).
		First(&r).
		RecordNotFound() {
		return false
	}
	var pm Policy
	if DB().
		Where("user_id = ? AND role_id = ?", user, r.ID).
		First(&pm).
		RecordNotFound() {
		return false
	}

	return pm.Enable()
}

// GetRole create role if not exist
func GetRole(name string, rty string, rid uint) (*Role, error) {
	var e error
	r := Role{}
	db := DB()
	if db.
		Where("name = ? AND resource_type = ? AND resource_id = ?", name, rty, rid).
		First(&r).
		RecordNotFound() {
		r = Role{
			Name:         name,
			ResourceType: rty,
			ResourceID:   rid,
		}
		e = db.Create(&r).Error

	}
	return &r, e
}

//Deny deny permission
func Deny(role uint, user uint) error {
	return DB().
		Where("role_id = ? AND user_id = ?", role, user).
		Delete(Policy{}).Error
}

//Allow allow permission
func Allow(role, user uint, years, months, days int) error {
	begin := time.Now()
	end := begin.AddDate(years, months, days)
	var count int
	DB().
		Model(&Policy{}).
		Where("role_id = ? AND user_id = ?", role, user).
		Count(&count)
	if count == 0 {
		return DB().Create(&Policy{
			UserID:   user,
			RoleID:   role,
			StartUp:  begin,
			ShutDown: end,
		}).Error
	}
	return DB().
		Model(&Policy{}).
		Where("role_id = ? AND user_id = ?", role, user).
		UpdateColumns(map[string]interface{}{"begin": begin, "end": end}).Error

}

// ListUserByResource list users by resource
func ListUserByResource(role, rty string, rid uint) ([]uint, error) {
	ror, err := GetRole(role, rty, rid)
	if err != nil {
		return nil, err
	}

	var ids []uint
	var policies []Policy
	if err := DB().Where("role_id = ?", ror.ID).Find(&policies).Error; err != nil {
		return nil, err
	}
	for _, pm := range policies {
		if pm.Enable() {
			ids = append(ids, pm.UserID)
		}
	}
	return ids, nil
}

// Resources list resource ids by user and role
func Resources(user uint, role, rty string) ([]uint, error) {
	var ids []uint
	var policies []Policy
	if err := DB().Where("user_id = ?", user).Find(&policies).Error; err != nil {
		return nil, err
	}
	for _, pm := range policies {
		if pm.Enable() {
			var ror Role
			if err := DB().Where("id = ?", pm.RoleID).First(&ror).Error; err != nil {
				return nil, err
			}
			if ror.Name == role && ror.ResourceType == rty {
				ids = append(ids, ror.ResourceID)
			}
		}
	}
	return ids, nil
}
