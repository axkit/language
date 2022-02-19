package language

import (
	"fmt"
	"testing"


/*
func TestNewBitSet(t *testing.T) {

	var cases = []struct {
		result language.BitSet
		ids    []Index
	}{
		{1, []Index{0}},
		{2, []Index{1}},
		{3, []Index{0, 1}},
		{0, nil},
		{0, []Index{-1}},
	}

	for i := range cases {
		if cases[i].result != NewBitSet(cases[i].ids) {
			t.Log("Failed case: ", i, NewBitSet(cases[i].ids))
			t.Fail()
		}
	}
}

func Test_mergeTranslations(t *testing.T) {
	tcase := []struct {
		line string
		exp  string
	}{
		{"", ""},
		{"noCustomlongSL={Key}", "noCustomlongSL={Key}"},
		{"key=", "key="},
		{`noCustomlongML={Key
is probably			
expired }`, `noCustomlongML={Key
is probably			
expired }`},
		{"# comment", "# comment"},
		{"customlongSL={KeyA}", "customlongSL={KeyB}"},
		{"noCustom=Yes", "noCustom=Yes"},
		{`customlongML={Key
is probably			
expiredA }`, `customlongML={Key
is probably			
expiredB }`},
		{"withCustom=A", "withCustom=B"},
		{`lastCustomlongML={Key
expiredA }`, `lastCustomlongML={Key
expiredB }`},
	}

	custom := []string{
		"withCustom=B",
		"customlongSL={KeyB}",
		`customlongML={Key
is probably			
expiredB }`,
		`lastCustomlongML={Key
expiredB }`,
	}

	var orig []string
	for i := range tcase {
		orig = append(orig, tcase[i].line)
	}

	res := mergeTranslations(orig, custom)

	if len(res) != len(orig) {
		t.Errorf("result length is different")
	}

	for i := range res {
		if res[i] != tcase[i].exp {
			t.Errorf("expected: '%s', got '%s', line: %d", tcase[i].exp, res[i], i)
		}
		fmt.Printf("%s\n", res[i])
	}

}
*/