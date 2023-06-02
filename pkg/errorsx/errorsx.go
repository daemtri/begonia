package errorsx

import "strings"

type Errors []error

func (m Errors) Error() string {
	errStrings := make([]string, 0, len(m))
	for i := range m {
		errStrings = append(errStrings, m[i].Error())
	}
	return strings.Join(errStrings, "; ")
}

func Append(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}
	me, ok := err1.(Errors)
	if ok {
		return append(me, err2)
	}
	return Errors([]error{err1, err2})
}
