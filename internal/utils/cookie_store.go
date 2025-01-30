package utils

import (
	"fmt"
	"strings"
)

type CookieStore struct {
	Cookies map[string]string
}

func NewCookieStore() *CookieStore {
	return &CookieStore{
		Cookies: make(map[string]string),
	}
}

func (c *CookieStore) Get(name string) (string, bool) {
	value, ok := c.Cookies[name]
	return value, ok
}

func (c *CookieStore) GetAll() string {
	cookieHeader := ""
	for name, value := range c.Cookies {
		cookieHeader += fmt.Sprintf("%s=%s; ", name, value)
	}
	cookieHeader = strings.TrimRight(cookieHeader, "; ")
	return cookieHeader
}

func (c *CookieStore) Set(name, value string) {
	c.Cookies[name] = value
}

func (c *CookieStore) Delete(name string) {
	delete(c.Cookies, name)
}
