package nut

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	yaml "gopkg.in/yaml.v2"
)

const (
	_defaultTTL = time.Hour * 24 * 30
)

var (
	_locales = make(map[string]map[string]string)
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
	// load from database
	var items []Locale
	if err := DB().Select([]string{"lang", "code", "message"}).Find(&items).Error; err != nil {
		return err
	}
	for _, it := range items {
		setLocale(it.Lang, it.Code, it.Message)
	}
	// load from files
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		names := strings.Split(info.Name(), ".")
		if info.Mode().IsRegular() && len(names) == 3 && names[2] == "yaml" {
			if err != nil {
				return err
			}

			log.Debugf("Find locale file %s", path)
			lang, err := language.Parse(names[1])
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

			return loopLocaleFileNode(names[0], val, func(code string, message string) error {
				// log.Debugf("%s.%s = %s", lang.String(), code, message)
				setLocale(lang.String(), code, message)
				return nil
			})
		}
		return nil
	})
}

func loopLocaleFileNode(r string, m map[interface{}]interface{}, f func(string, string) error) error {
	for k, v := range m {
		ks, ok := k.(string)
		if ok {
			ks = r + "." + ks
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
