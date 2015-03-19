// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"fmt"
	"strings"
)

type Pattern []string

func NewPattern(pattern string) Pattern {
	return Pattern(strings.Split(pattern, "*"))
}

func (pattern Pattern) Match(key string) (result []string, ok bool) {
	for _, entry := range pattern {

		i := strings.Index(key, entry)
		if i < 0 {
			return
		}

		if i > 0 {
			result = append(result, key[0:i])
		}

		key = key[i+len(entry):]
	}

	if key != "" {
		result = append(result, key)
	}

	ok = true
	return
}

func (pattern Pattern) String() string {
	return fmt.Sprintf("%s", []string(pattern))
}
