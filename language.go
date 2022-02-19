package language

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

var languages []string

// Index is integer representative of language short code.
//
type Index int

// Unknown used if language code is not found.
const Unknown Index = -1

// ToIndex returns index by language code: ru, en, sr, cz...
func ToIndex(lang string) Index {
	for i := range languages {
		if languages[i] == lang {
			return Index(i)
		}
	}
	languages = append(languages, lang)
	return Index(len(languages) - 1)
}

// IndexToCode returns language code.
func IndexToCode(index Index) string {
	return languages[index]
}

// Supported returns code of supported languages.
func Supported() []string {
	return languages
}

// TODO сделать определение ближайшего языка на основании заголовка Accepted-Language

// NameColumn is a type of column Name in regular reference table
type NameColumn []byte

// Name holds decoded names. Index of the slice calculates by ToIndex().
type Name []string

// Name decodes jsonb into array of strings
func (rn NameColumn) Name() (Name, error) {

	var n map[string]string

	err := json.Unmarshal(rn, &n)
	if err != nil {
		return Name{}, err
	}

	var result Name

	for lang, val := range n {
		if idx := ToIndex(lang); idx != Unknown {
			result[idx] = val
		} else {
			// TODO: error handling
		}
	}
	return result, nil
}

func (rn *Name) Byte() []byte {

	// TODO: переделать без использования строк и с использованием bytes.Buffer

	result := ""
	s := "{"
	for lang, val := range *rn {
		result += s + "\"" + languages[lang] + "\":\"" + val + "\""
		s = ","
	}
	result += "}"
	return []byte(result)
}

// Value implements interface sql.Valuer
func (n *Name) Value() (driver.Value, error) {
	return n.MarshalJSON()
}

// Scan implements database/sql Scanner interface.
func (n *Name) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Name.Scan: expected []byte, got %T (%q)", value, value)
	}

	var err error
	*n, err = ToName(v)
	return err
}

// Name decodes jsonb into array of strings
func ToName(b []byte) (Name, error) {

	var n map[string]string

	err := json.Unmarshal(b, &n)
	if err != nil {
		return Name{}, err
	}

	var result Name

	for lang, val := range n {
		if idx := ToIndex(lang); idx != Unknown {
			result[idx] = val
		} else {
			// TODO: error handling
		}
	}
	return result, nil
}

func (n *Name) MarshalJSON() ([]byte, error) {

	// TODO: переделать без использования строк и с использованием bytes.Buffer

	result := ""
	s := "{"
	for lang, val := range *n {
		result += s + "\"" + languages[lang] + "\":\"" + val + "\""
		s = ","
	}
	result += "}"
	return []byte(result), nil
}

func (n *Name) UnmarshalJSON(buf []byte) error {
	name, err := ToName(buf)
	if err != nil {
		return err
	}

	*n = name
	return nil
}

func IsLanguageName(b []byte) bool {

	if !json.Valid(b) {
		return false
	}

	var n map[string]string

	err := json.Unmarshal(b, &n)
	if err != nil {
		return false
	}

	for lang := range n {
		if idx := ToIndex(lang); idx == Unknown {
			return false
		}
	}
	return true
}
