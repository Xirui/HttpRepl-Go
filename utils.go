package main

import (
	"errors"
	"strings"
	"unicode"
)

// splitArgs splits a command line string into separate arguments,
// respecting single and double quotes and backslash escaping.
func splitArgs(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inDoubleQuotes := false
	inSingleQuotes := false
	escaped := false

	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			if inSingleQuotes {
				current.WriteRune(r)
			} else {
				escaped = true
			}
			continue
		}

		if r == '"' && !inSingleQuotes {
			inDoubleQuotes = !inDoubleQuotes
			continue
		}

		if r == '\'' && !inDoubleQuotes {
			inSingleQuotes = !inSingleQuotes
			continue
		}

		if unicode.IsSpace(r) && !inDoubleQuotes && !inSingleQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	if inDoubleQuotes || inSingleQuotes {
		return nil, errors.New("unclosed quotes")
	}
	if escaped {
		return nil, errors.New("trailing backslash")
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}
