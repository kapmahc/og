package nut

import (
	"bytes"
	"encoding/gob"
	"time"
)

// Setting k-v
type Setting struct {
	ID        uint      `gorm:"primary_key"`
	Key       string    `gorm:"not null;unique_index;type:VARCHAR(255)"`
	Val       []byte    `gorm:"not null"`
	Encode    bool      `gorm:"not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName table name
func (Setting) TableName() string {
	return "settings"
}

//Set save setting
func Set(k string, v interface{}, f bool) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return err
	}
	var val []byte
	if f {
		if val, err = Encrypt(buf.Bytes()); err != nil {
			return err
		}
	} else {
		val = buf.Bytes()
	}

	var m Setting
	if DB().Where("key = ?", k).First(&m).RecordNotFound() {
		return DB().Create(&Setting{
			Key:    k,
			Val:    val,
			Encode: f,
		}).Error
	}

	return DB().Model(&m).Updates(map[string]interface{}{
		"encode": f,
		"val":    val,
	}).Error
}

//Get get setting value by key
func Get(k string, v interface{}) error {
	var m Setting
	if err := DB().Where("key = ?", k).First(&m).Error; err != nil {
		return err
	}

	var buf bytes.Buffer
	dec := gob.NewDecoder(&buf)

	if m.Encode {
		vl, er := Decrypt(m.Val)
		if er != nil {
			return er
		}
		buf.Write(vl)
	} else {
		buf.Write(m.Val)
	}

	return dec.Decode(v)
}
