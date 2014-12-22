// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import ()

func Join(prefix, suffix string) string {
	if prefix == "" {
		return suffix
	}

	if suffix == "" {
		return prefix
	}

	return prefix + "." + suffix
}
