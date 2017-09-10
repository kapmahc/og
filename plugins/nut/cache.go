package nut

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/garyburd/redigo/redis"
)

// CacheSet set
func CacheSet(k string, v interface{}, ttl time.Duration) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return err
	}

	c := Redis().Get()
	defer c.Close()
	_, err := c.Do("SET", k, buf.Bytes(), "EX", int(ttl/time.Second))
	return err
}

// CacheGet get
func CacheGet(k string, v interface{}) error {
	c := Redis().Get()
	defer c.Close()
	bys, err := redis.Bytes(c.Do("GET", k))

	if err != nil {
		return err
	}
	var buf bytes.Buffer
	dec := gob.NewDecoder(&buf)
	buf.Write(bys)
	return dec.Decode(v)
}
