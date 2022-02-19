package language

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func TranslationResource(dir string, langfilename string, customSuffixes ...string) ([]string, error) {

	orig, err := readi18File(filepath.Join(dir, langfilename))
	if err != nil {
		return nil, err
	}

	var (
		custom []string
		cname  string
	)

	for i := range customSuffixes {
		if customSuffixes[i] == "" {
			continue
		}

		cname = filepath.Join(dir, langfilename+"."+customSuffixes[i])
		_, err = os.Stat(cname)

		if err != nil && os.IsNotExist(err) {
			return orig, nil
		}

		custom, err = readi18File(cname)
		if err != nil {
			return orig, nil
		}

		orig = mergeTranslations(orig, custom)
	}
	/*
		if customFileSuffix != "" {
			cname = filepath.Join(dir, langfilename+"."+customFileSuffix)
			_, err = os.Stat(cname)
		} else {
			cname = filepath.Join(dir, langfilename+"."+defaultCustomTranslationFileSuffix)
			_, err = os.Stat(cname)
		}

		if err != nil && os.IsNotExist(err) {
			return orig, nil
		}

		custom, err = readi18File(cname)
		if err != nil {
			return orig, nil
		}

		return mergeTranslations(orig, custom), nil */
	return orig, nil
}

func readi18File(fname string) ([]string, error) {

	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var res []string
	for {
		if scanner.Scan() {
			res = append(res, scanner.Text())
		} else {
			break
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func mergeTranslations(orig, cust []string) []string {
	var res []string
	for i := range orig {
		a := strings.TrimLeft(orig[i], " ")
		if a == "" || strings.HasPrefix(a, "#") || strings.HasPrefix(a, "[") {
			res = append(res, a)
			continue
		}

		pair := strings.SplitN(a, "=", 2)
		if len(pair) < 2 {
			res = append(res, a)
			continue
		}

		key := strings.TrimSpace(pair[0]) + "="
		val := strings.TrimLeft(pair[1], " ")

		ismultiline := false
		if len(val) > 0 {
			ismultiline = (val[0] == '{')
		}

		found := false
		for j := 0; j < len(cust); j++ {
			b := strings.TrimLeft(cust[j], " ")
			if strings.HasPrefix(b, key) == false {
				continue
			}

			res = append(res, b)
			found = true
			if ismultiline == false {
				break
			}

			if strings.HasSuffix(strings.TrimRight(b, " "), "}") {
				break
			}

			// looking for close bracket in multiline
			for j++; j < len(cust); j++ {
				b = cust[j]
				res = append(res, b)
				if strings.HasSuffix(strings.TrimRight(b, " "), "}") {
					break
				}
			}
		}
		if !found {
			res = append(res, a)
		}
	}

	return res
}
