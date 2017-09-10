package nut

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SermoDigital/jose/jws"
)

const (
	actConfirm       = "confirm"
	actUnlock        = "unlock"
	actResetPassword = "reset-password"

	// SendEmailJob send email
	SendEmailJob = "send-email"
)

func sendEmail(lng string, req *http.Request, user *User, act string) error {
	cm := jws.Claims{}
	cm.Set("act", act)
	cm.Set("uid", user.UID)
	tkn, err := SumJwtToken(cm, time.Hour*6)
	if err != nil {
		return err
	}

	obj := struct {
		Home  string
		Token string
	}{
		Home:  req.URL.Host,
		Token: string(tkn),
	}
	subject, err := F(lng, fmt.Sprintf("auth.emails.%s.subject", act), obj)
	if err != nil {
		return err
	}
	body, err := F(lng, fmt.Sprintf("auth.emails.%s.body", act), obj)
	if err != nil {
		return err
	}

	// -----------------------
	buf, err := json.Marshal(map[string]string{
		"to":      user.Email,
		"subject": subject,
		"body":    body,
	})
	if err != nil {
		return err
	}
	return Send(1, SendEmailJob, buf)
}

func parseToken(lng, tkn, act string) (*User, error) {
	cm, err := ValidateJwtToken([]byte(tkn))
	if err != nil {
		return nil, err
	}
	if act != cm.Get("act").(string) {
		return nil, E(lng, "errors.bad-action")
	}
	return GetUserByUID(cm.Get(UID).(string))
}
