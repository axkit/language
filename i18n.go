package language

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	FileExtension  = ".i18n"
	HintSeparator  = "//"
	NotFoundMarker = "\u2638"
)

type RequestStrategy int8

const (
	ReturnNotFoundVariable RequestStrategy = iota
	ReturnEmptyString
	ReturnInPrimaryLanguage
)

type key struct {
	lang Index
	// custom suffix en.{custom}.i18n
	// empty for default resource file.
	custom string
}

type file struct {
	key
	name     string
	fullName string // path + file name
}

// Item represents a row in a .i18n file.
type Item struct {
	Key   string
	Value string
	Hint  string
}

// ResponseItem represents a row to be returned to the client.
type ResponseItem struct {
	Value string `json:"v"`
	Hint  string `json:"h,omitempty"`
}

// Set holds a set of items.
type Set struct {
	items []Item
	index map[string]int // key -> index in Items
}

// Container stores for all translation resources.
type Container struct {
	cfg          Option
	translations map[key]Set
	files        []file
	customDirs   []string
}

// Option defines options for Container.
type Option struct {
	// Default language
	primaryLanguage Index

	// suffixPriority maps text and it's priority
	suffixPriority map[string]int

	bracketSymbol string
}

// WithPrimaryLanguage assigns a primary language.
func WithPrimaryLanguage(li Index) func(o *Option) {
	return func(o *Option) {
		o.primaryLanguage = li
	}
}

// WithSuffixes assigns suffixes of translation files in the order of applying priority.
// The first suffix has the highest priority.
func WithSuffixes(suffix ...string) func(o *Option) {
	return func(o *Option) {
		for i, s := range suffix {
			o.suffixPriority[s] = i
		}
	}
}

// WithBrackets assignss wrapping symbol used by .
func WithBrackets(bracketSymbol string) func(o *Option) {
	return func(o *Option) {
		o.bracketSymbol = bracketSymbol
	}
}

// New creates a new translations container.
func New(fn ...func(o *Option)) *Container {
	c := Container{
		translations: make(map[key]Set),
		cfg: Option{
			primaryLanguage: -1, // option is not set
			suffixPriority:  make(map[string]int),
		},
	}

	for _, f := range fn {
		f(&c.cfg)
	}
	return &c
}

// AddFiles registers .i18n files in the container.
// Returns error if even one could not be found or it's a directory.
func (c *Container) AddFiles(filenames ...string) error {
	for _, filename := range filenames {
		fi, err := os.Stat(filename)
		if err != nil {
			return err
		}

		pfi, err := parseFileInfo(fi)
		if err != nil {
			return err
		}
		pfi.fullName = filename
		c.files = append(c.files, pfi)
	}
	return nil
}

func parseFileInfo(fi fs.FileInfo) (file, error) {
	var res file

	if fi.IsDir() {
		return res, errors.New(fi.Name() + " not a file")
	}

	res = file{
		name: fi.Name(),
	}

	res.lang, res.custom = parseFileName(fi.Name())
	return res, nil
}

// AddFileByMask registers .i18n files matching mask from the path specified by path.
func (c *Container) AddFileByMask(dir string, mask string) error {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}

		if len(mask) > 0 && mask != "*" {
			ok, err := filepath.Match(mask, de.Name())
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
		}

		fi, err := de.Info()
		if err != nil {
			return err
		}

		pfi, err := parseFileInfo(fi)
		if err != nil {
			return err
		}
		pfi.fullName = filepath.Join(dir, pfi.name)
		c.files = append(c.files, pfi)
	}
	return nil
}

// AddCustomDir registers a directory with custom translation files.
func (c *Container) AddCustomDir(dirs ...string) error {
	for _, d := range dirs {
		f, err := os.Stat(d)
		if err != nil {
			return err
		}
		if !f.IsDir() {
			return errors.New(f.Name() + " not a directory")
		}
		c.customDirs = append(c.customDirs, d)
	}
	return nil
}

func (c *Container) ListenFileChange() error {
	return nil
}

func (c *Container) sortFilesBySuffixPriority() {
	sort.Slice(c.files, func(i, j int) bool {
		if c.files[i].lang == c.files[j].lang {
			return c.cfg.suffixPriority[c.files[i].custom] < c.cfg.suffixPriority[c.files[j].custom]
		}
		return c.files[i].lang < c.files[j].lang
	})
}

// ReadRegisteredFiles reads content of all registered files and stores items in the container.
func (c *Container) ReadRegisteredFiles() error {

	c.sortFilesBySuffixPriority()

	for i := range c.files {
		items, err := c.loadFile(c.files[i].fullName)
		if err != nil {
			return err
		}

		key := key{
			lang:   c.files[i].lang,
			custom: "",
		}

		if ti, ok := c.translations[key]; ok {
			// replace
			for j := range items {
				if idx, ok := ti.index[items[j].Key]; ok {
					ti.items[idx] = items[j]
				} else {
					ti.items = append(ti.items, items[j])
					ti.index[items[j].Key] = len(ti.items) - 1
				}
			}
		} else {
			x := Set{index: make(map[string]int)}
			x.items = items
			for i, item := range items {
				x.index[item.Key] = i
			}
			c.translations[key] = x
		}
	}
	return nil
}

func parseFileName(filename string) (li Index, suffix string) {
	from := strings.Index(filename, ".")
	to := strings.LastIndex(filename, ".")
	if from == -1 && to == -1 {
		// it also covers the case when filename has not "."
		return Unknown, ""
	}

	if from == to {
		return ToIndex(filename[0:from]), ""
	}

	return ToIndex(filename[0:from]), filename[from+1 : to]
}

func (c *Container) loadFile(filename string) ([]Item, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var res []Item
	for {
		if !scanner.Scan() {
			break
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		line := scanner.Text()
		line = strings.TrimLeft(line, " ")
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		item := c.parseLine(line)
		if item != nil {
			res = append(res, *item)
		}
	}

	return res, nil
}

func (c *Container) parseLine(line string) *Item {
	var res Item

	vx := strings.Index(line, "=")
	if vx == -1 {
		return nil
	}

	res.Key = strings.TrimSpace(line[0:vx])
	val := strings.TrimSpace(line[vx+1:])
	hx := strings.Index(val, HintSeparator)
	if hx != -1 {
		res.Hint = strings.TrimSpace(val[hx+1:])
		res.Value = strings.TrimSpace(val[0:hx])
	} else {
		res.Value = val
	}

	return &res
}

type ContainerRequest struct {
	lang Index
	c    *Container
}

func (c *Container) Lang(li Index) ContainerRequest {
	return ContainerRequest{
		lang: li,
		c:    c,
	}
}

func (c *ContainerRequest) Value(id string) string {
	res, ok := c.item(id)
	if !ok {
		return id + NotFoundMarker
	}
	return res.Value
}

func (c *ContainerRequest) Hint(id string) string {
	res, ok := c.item(id)
	if !ok {
		return ""
	}
	return res.Hint
}

func (cr *ContainerRequest) item(id string) (Item, bool) {

	if len(id) > 2 &&
		cr.c.cfg.bracketSymbol != "" &&
		strings.HasPrefix(id, cr.c.cfg.bracketSymbol) &&
		strings.HasSuffix(id, cr.c.cfg.bracketSymbol) {
		id = id[1 : len(id)-1]
	}

	rsi, ok := cr.c.translations[key{lang: cr.lang, custom: ""}]
	if ok {
		var idx int
		if idx, ok = rsi.index[id]; ok {
			return rsi.items[idx], true
		}
	}

	if cr.c.cfg.primaryLanguage == Unknown {
		return Item{}, false
	}

	rsi, ok = cr.c.translations[key{lang: cr.c.cfg.primaryLanguage, custom: ""}]
	if ok {
		var idx int
		if idx, ok = rsi.index[id]; ok {
			return rsi.items[idx], true
		}
	}
	return Item{}, false
}

func (c *ContainerRequest) ValueWithDefault(id string, notFoundValue string) string {
	res, ok := c.item(id)
	if !ok {
		return notFoundValue
	}
	return res.Value
}

// JSON returns translation in JSON format.
func (cr *ContainerRequest) JSON() ([]byte, error) {
	kv := make(map[string]ResponseItem)
	set, ok := cr.c.translations[key{lang: cr.lang}]
	if !ok {
		if cr.c.cfg.primaryLanguage != Unknown {
			set, ok = cr.c.translations[key{lang: cr.c.cfg.primaryLanguage}]
		}
	}
	if !ok {
		return nil, errors.New("no translation found")
	}

	for _, item := range set.items {
		kv[cr.c.genKey(item.Key)] = ResponseItem{
			Value: item.Value,
			Hint:  item.Hint,
		}
	}

	if cr.c.cfg.primaryLanguage != Unknown && cr.c.cfg.primaryLanguage != cr.lang {
		if set, ok = cr.c.translations[key{lang: cr.c.cfg.primaryLanguage}]; ok {
			for _, item := range set.items {
				k := cr.c.genKey(item.Key)
				if _, ok := kv[k]; ok {
					continue
				}
				kv[k] = ResponseItem{
					Value: item.Value,
					Hint:  item.Hint,
				}
			}
		}
	}

	buf, err := json.Marshal(kv)
	return buf, err
}

// genKey generates resource key for JSON response.
func (c *Container) genKey(id string) string {
	if c.cfg.bracketSymbol == "" {
		return id
	}
	return c.cfg.bracketSymbol + id + c.cfg.bracketSymbol
}
