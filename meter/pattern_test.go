// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"testing"
)

func TestPattern(t *testing.T) {
	patternOk(t, "a", "a.b.c", []string{".b.c"})
	patternOk(t, "b", "a.b.c", []string{"a.", ".c"})
	patternOk(t, "c", "a.b.c", []string{"a.b."})
	patternFail(t, "d", "a.b.c")

	patternOk(t, "*.b.c", "a.b.c", []string{"a"})
	patternOk(t, "a.*.c", "a.b.c", []string{"b"})
	patternOk(t, "a.b.*", "a.b.c", []string{"c"})
	patternOk(t, "a.b.c*", "a.b.c", []string{})
	patternOk(t, "*a.b.c", "a.b.c", []string{})
	patternOk(t, "a.*b*.c", "a.b.c", []string{})
	patternFail(t, "b.a.*", "a.b.c")

	patternOk(t, "*.b.*.d", "a.b.c.d.e", []string{"a", "c", ".e"})
	patternOk(t, "a.*.c.*.e", "a.b.c.d.e", []string{"b", "d"})
	patternOk(t, "a.b.*.d.*", "a.b.c.d.e", []string{"c", "e"})

}

func patternOk(t *testing.T, pattern, key string, exp []string) {
	p := NewPattern(pattern)

	groups, ok := p.Match(key)
	if !ok {
		t.Errorf("FAIL(%v, %s): unmatched", p, key)
	}

	if len(groups) != len(exp) {
		t.Errorf("FAIL(%v, %s): group mismatch %v != %v", p, key, groups, exp)
		return
	}

	for i, group := range groups {
		if group != exp[i] {
			t.Errorf("FAIL(%v, %s): group mismatch %v != %v", p, key, groups, exp)
			return
		}
	}
}

func patternFail(t *testing.T, pattern, key string) {
	p := NewPattern(pattern)
	if groups, ok := p.Match(key); ok {
		t.Errorf("FAIL(%v, %s): unexpected success -> %v", p, key, groups)
	}
}
