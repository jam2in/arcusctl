package internal

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	minPasswordLen     = 12
	minPasswordClasses = 3
)

var passwordClasses = []struct {
	name     string
	contains func(rune) bool
}{
	{"uppercase", unicode.IsUpper},
	{"lowercase", unicode.IsLower},
	{"digits", unicode.IsDigit},
	{"symbols", func(r rune) bool { return unicode.IsPunct(r) || unicode.IsSymbol(r) }},
}

func validatePassword(password string) error {
	if length := utf8.RuneCountInString(password); length < minPasswordLen {
		return fmt.Errorf(
			"password is %d characters, must be at least %d",
			length,
			minPasswordLen,
		)
	}

	missing := make([]string, 0, len(passwordClasses))
	found := 0
	for _, class := range passwordClasses {
		if strings.ContainsFunc(password, class.contains) {
			found++
			continue
		}
		missing = append(missing, class.name)
	}
	if found < minPasswordClasses {
		return fmt.Errorf(
			"password has %d character classes, must include at least %d (missing: %s)",
			found,
			minPasswordClasses,
			strings.Join(missing, ", "),
		)
	}

	return nil
}
