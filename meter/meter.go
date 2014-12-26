// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"time"
)

// Meter represents a meter used to record various metrics which will be read
// periodically. The time.Duration parameter is used to normalize the values
// read by the meter so that they represent a per second values.
type Meter interface {
	ReadMeter(time.Duration) map[string]float64
}

// Join concatenates the given items with a '.' character where necessary.
func Join(items ...string) string {
	result := ""

	for _, item := range items {

		if result == "" {
			result = item

		} else if item != "" {
			result = result + "." + item
		}

	}

	return result
}
