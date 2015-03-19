// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"fmt"
	"strings"
)

type TranslationHandler struct {
	input  []Pattern
	output []string
}

func NewTranslationHandler(patterns map[string]string) *TranslationHandler {
	handler := &TranslationHandler{}

	for input, output := range patterns {
		handler.input = append(handler.input, NewPattern(input))
		handler.output = append(handler.output, output)
	}

	return handler
}

func (handler *TranslationHandler) HandleMeters(values map[string]float64) {
	for key, value := range values {
		if newKey, ok := handler.apply(key); ok {
			values[newKey] = value
		}
	}
}

func (handler *TranslationHandler) apply(key string) (string, bool) {
	for i, in := range handler.input {
		groups, ok := in.Match(key)
		if !ok {
			continue
		}

		newKey := handler.output[i]
		for i, group := range groups {
			newKey = strings.Replace(newKey, fmt.Sprintf("{%d}", i), group, -1)
		}

		return newKey, true
	}

	return key, false
}
