package nut

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	yaml "gopkg.in/yaml.v2"
)

const (
	// LOCALE locale key
	LOCALE = "locale"
)

var (
	_locales       = make(map[string]map[string]string)
	_localeMatcher language.Matcher
)

//Locale locale
type Locale struct {
	ID        uint      `gorm:"primary_key"`
	Lang      string    `gorm:"not null;type:varchar(8);index;unique_index:idx_locales_lang_code"`
	Code      string    `gorm:"not null;index;type:VARCHAR(255);unique_index:idx_locales_lang_code"`
	Message   string    `gorm:"not null;type:varchar(1024)"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName table name
func (Locale) TableName() string {
	return "locales"
}

func getLocale(lang, code string) string {
	if items, ok := _locales[lang]; ok {
		return items[code]
	}
	return ""
}

// F format message
func F(lang, code string, obj interface{}) (string, error) {
	msg := getLocale(lang, code)
	if msg == "" {
		return code, nil
	}
	tpl, err := template.New("").Parse(msg)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, obj)
	return buf.String(), err
}

//E create an i18n error
func E(lang, code string, args ...interface{}) error {
	msg := getLocale(lang, code)
	if msg == "" {
		return errors.New(code)
	}
	return fmt.Errorf(msg, args...)
}

//T translate by lang tag
func T(lang, code string, args ...interface{}) string {
	msg := getLocale(lang, code)
	if msg == "" {
		return code
	}
	return fmt.Sprintf(msg, args...)
}

func setLocale(lang, code, message string) {
	if _, ok := _locales[lang]; !ok {
		_locales[lang] = make(map[string]string)
	}
	_locales[lang][code] = message
}

// loadLocales locales from  database and yaml files
func loadLocales(dir string) error {

	// load from files
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		const ext = ".yml"
		name := info.Name()
		if info.Mode().IsRegular() && filepath.Ext(name) == ext {
			log.Debugf("Find locale file %s", path)
			lang, err := language.Parse(name[:len(name)-len(ext)])
			if err != nil {
				return err
			}

			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			val := make(map[interface{}]interface{})
			if err = yaml.Unmarshal(buf, &val); err != nil {
				return err
			}

			return loopLocaleFileNode("", val, func(code string, message string) error {
				// log.Debugf("%s.%s = %s", lang.String(), code, message)
				setLocale(lang.String(), code, message)
				return nil
			})
		}
		return nil
	}); err != nil {
		return err
	}

	// load from database
	var items []Locale
	if err := DB().Select([]string{"lang", "code", "message"}).Find(&items).Error; err != nil {
		return err
	}
	for _, it := range items {
		setLocale(it.Lang, it.Code, it.Message)
	}

	// init matcher
	var tags []language.Tag
	for _, l := range Languages() {
		tags = append(tags, language.Make(l))
	}
	_localeMatcher = language.NewMatcher(tags)
	return nil
}

func loopLocaleFileNode(r string, m map[interface{}]interface{}, f func(string, string) error) error {
	for k, v := range m {
		ks, ok := k.(string)
		if ok {
			if r != "" {
				ks = r + "." + ks
			}
			vs, ok := v.(string)
			if ok {
				if e := f(ks, vs); e != nil {
					return e
				}
			} else {
				vm, ok := v.(map[interface{}]interface{})
				if ok {
					if e := loopLocaleFileNode(ks, vm, f); e != nil {
						return e
					}
				}
			}
		}
	}
	return nil
}

// Languages support languages
func Languages() []string {
	var items []string
	for k := range _locales {
		items = append(items, k)
	}
	return items
}
func detectLocale(r *http.Request) string {
	// 1. Check URL arguments.
	if lang := r.URL.Query().Get(LOCALE); lang != "" {
		return lang
	}

	// 2. Get language information from cookies.
	if ck, er := r.Cookie(LOCALE); er == nil {
		return ck.Value
	}

	// 3. Get language information from 'Accept-Language'.
	if al := r.Header.Get("Accept-Language"); len(al) > 4 {
		return al[:5] // Only compare first 5 letters.
	}

	return ""
}

// LocaleMiddleware detect locale
func LocaleMiddleware(c *gin.Context) {
	tag, _, _ := _localeMatcher.Match(language.Make(detectLocale(c.Request)))
	c.Set(LOCALE, tag.String())
}
