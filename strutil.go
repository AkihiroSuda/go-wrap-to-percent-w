package main

import (
	"strconv"
)

func unquote(s string) string {
	u, err := strconv.Unquote(s)
	if err != nil {
		return s
	}
	return u
}
