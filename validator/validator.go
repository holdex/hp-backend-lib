package libvalidator

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/grpc/codes"
)

func New(errs ...error) V {
	v := V(nil)
	for _, err := range errs {
		if err != nil {
			v = append(v, err)
		}
	}
	return v
}

type V []error

func (v V) Code() string {
	return codes.InvalidArgument.String()
}

func (v V) Error() string {
	if len(v) == 0 {
		return ""
	}
	var msg string
	for _, err := range v {
		msg = fmt.Sprintln(msg, err.Error())
	}
	return msg
}

func (v V) Err(field, msg string, args ...interface{}) V {
	if len(args) > 0 {
		return append(v, fmt.Errorf("%s: %s", field, fmt.Sprintf(msg, args...)))
	}
	return append(v, fmt.Errorf("%s: %s", field, msg))
}

func (v V) Len(field, value string, min, max uint) V {
	if l := uint(len(strings.Trim(value, " "))); l < min {
		return v.Err(field, "too short: min length is %d", min)
	} else if l > max {
		return v.Err(field, "too long: max length is %d", max)
	}
	return v
}

func (v V) True(field string, value bool) V {
	if !value {
		return v.Err(field, "is false, should be true")
	}
	return v
}

func (v V) NotNil(field string, value interface{}) V {
	if value == nil {
		return v.Err(field, "is nil")
	}
	return v
}

func (v V) GT(field string, value, compare float64) V {
	if value <= compare {
		return v.Err(field, "is not greater than: %f", compare)
	}
	return v
}

func (v V) LT(field string, value, compare float64) V {
	if value >= compare {
		return v.Err(field, "is not lower than: %f", compare)
	}
	return v
}

func (v V) Regex(field, value string, pattern *regexp.Regexp) V {
	if ok := pattern.MatchString(value); !ok {
		return v.Err(field, "%s does not match regexp %s", value, pattern.String())
	}
	return v
}

var uuidRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)

func (v V) UUID(field, value string) V {
	return v.Regex(field, value, uuidRegex)
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]{1,60}@[a-z0-9.\-]{1,30}\.[a-z]{2,10}$`)

func (v V) Email(field, value string) V {
	return v.Regex(field, value, emailRegex)
}
