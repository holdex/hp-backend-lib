package libstrings

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	EmailPattern = `^[a-zA-Z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
)

func RemoveEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if !IsEmpty(str) {
			r = append(r, str)
		}
	}
	return r
}

func Contains(ss []string, str string) bool {
	for _, s := range ss {
		if s == str {
			return true
		}
	}
	return false
}

func IsEmpty(s string) bool {
	return len(strings.Trim(s, " ")) == 0
}

func IsEqual(s1, s2 string) bool {
	return strings.Trim(s1, " ") == strings.Trim(s2, " ")
}

func IsEmail(s string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return re.MatchString(s)
}

func MatchPattern(pattern, value string) bool {
	re := regexp.MustCompile(pattern)
	return re.MatchString(value)
}

func HtmlEscape(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		// "&#34;" is shorter than "&quot;".
		`"`, "&#34;",
		// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
		"'", "&#39;",
	).Replace(s)
}

func SqlEscape(s string) string {
	return strings.NewReplacer(
		"'", "''",
	).Replace(s)
}

func SqlStringArray(ss []string) (str string) {
	ln := len(ss) - 1
	for i, s := range ss {
		str += SqlEscape(s)
		if i != ln {
			str += "', '"
		}
	}
	return "array['" + str + "']"
}

func SqlStrings(ss []string) (str string) {
	ln := len(ss) - 1
	for i, s := range ss {
		str += SqlEscape(s)
		if i != ln {
			str += "', '"
		}
	}
	return "('" + str + "')"
}

func SqlIntArray(ss []int) (str string) {
	ln := len(ss) - 1
	for i, s := range ss {
		str += strconv.Itoa(s)
		if i != ln {
			str += ","
		}
	}
	return "array[" + str + "]"
}

func Secure(p string) string {
	return fmt.Sprintf("%d*", len(p))
}
