package language

import (
	"testing"
)

func TestContainer(t *testing.T) {

	t.Run("AddFileByMaskEmpty", func(t *testing.T) {
		c := New(WithSuffixes("reports", "prj", "reports.prj"))
		if err := c.AddFileByMask("testdata", ""); err != nil {
			t.Fatal(err)
		}
		if len(c.files) != 5 {
			t.Fatalf("expected %d files, got %d", 5, len(c.files))
		}
	})

	t.Run("AddFileByMask", func(t *testing.T) {
		c := New(WithSuffixes("reports", "prj", "reports.prj"))
		if err := c.AddFileByMask("testdata", "*.xxx"); err != nil {
			t.Fatal(err)
		}
		if len(c.files) != 0 {
			t.Errorf("expected 0 files, got %d", len(c.files))
		}
	})

	c := New(WithSuffixes("reports", "prj", "reports.prj"))

	t.Run("ReadRegisteredFiles", func(t *testing.T) {
		if err := c.AddFileByMask("testdata", ""); err != nil {
			t.Fatal(err)
		}

		if err := c.ReadRegisteredFiles(); err != nil {
			t.Fatal(err)
		}

		if len(c.translations) != 2 {
			t.Fatalf("expected %d translations, got %d", 2, len(c.translations))
		}
	})

	t.Run("ValueWithoutPrimaryLanguage", func(t *testing.T) {
		cr := c.Lang(ToIndex("en"))
		if cr.Value("Save") != "Persist" {
			t.Fatal("translation not found")
		}

		cr = c.Lang(ToIndex("de"))
		if cr.Value("Cancel") != "Abbrechen" {
			t.Fatal("translation not found")
		}

		if v := cr.Value("Exit"); v != "Exit"+NotFoundMarker {
			t.Fatalf("expected 'Exit%s', got '%s'", NotFoundMarker, v)
		}
	})
	t.Run("JSONWithoutPrimaryLanguage", func(t *testing.T) {
		cr := c.Lang(ToIndex("en"))
		buf, err := cr.JSON()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(buf))
	})

	t.Run("ValueWithPrimaryLanguage", func(t *testing.T) {
		c.cfg.primaryLanguage = ToIndex("en")

		cr := c.Lang(ToIndex("de"))
		if v := cr.Value("Exit"); v != "Sign out" {
			t.Fatalf("expected 'Sign out', got '%s'", v)
		}
	})

	t.Run("JSONWithPrimaryLanguage", func(t *testing.T) {
		c.cfg.primaryLanguage = ToIndex("en")
		cr := c.Lang(ToIndex("de"))
		buf, err := cr.JSON()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(buf))
	})

	t.Run("JSONWithPrimaryLanguageAndBrackets", func(t *testing.T) {
		c.cfg.primaryLanguage = ToIndex("en")
		c.cfg.bracketSymbol = "%"
		cr := c.Lang(ToIndex("de"))
		buf, err := cr.JSON()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(buf))
	})
}

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
