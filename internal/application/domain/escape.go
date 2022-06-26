package domain

import "strings"

var escape = strings.NewReplacer(".", `\.`, "#", `\#`, "=", `\=`, "+", `\+`, `-`, `\-`, `_`, `\_`)

func EscapeString(s string) string {
	return escape.Replace(s)
}
