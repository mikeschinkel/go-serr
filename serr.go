package serr

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	ExcerptFormat      = "%s%s%s"
	LengthPrefixFormat = "[len=%d] %s"
	EllipsisRune       = "\u2026"
)

type SError interface {
	error
	Args(...any) SError
	Err(error, ...any) SError
	Unwrap() error
	ValidArgs(...string) SError
	NoArgs() SError
}

type sError struct {
	error
	err       error
	args      []any
	validArgs []string
}

func New(msg string) SError {
	return &sError{
		error: errors.New(msg),
	}
}

func Cast(err error) SError {
	return &sError{
		error: err,
	}
}

func Wrap(err error, msg string, args ...any) SError {
	sErr := New(msg).Err(err)
	if len(args) > 0 {
		return sErr.Args(args...)
	}
	return sErr
}

func (se *sError) String() string {
	return se.error.Error()
}

func (se *sError) Error() string {
	if se.err != nil {
		return fmt.Sprintf("%s%s; %s",
			se.error.Error(),
			se.argsString(),
			se.err.Error(),
		)
	}
	return se.error.Error() + se.argsString()
}

func (se *sError) ValidArgs(args ...string) SError {
	se.validArgs = args
	return se
}

func (se *sError) NoArgs() SError {
	return se
}

func (se *sError) argsString() string {
	sb := strings.Builder{}
	for i := 0; i < len(se.args)-1; i += 2 {
		sb.WriteString(" [")
		sb.WriteString(fmt.Sprintf("%v", se.args[i]))
		sb.WriteByte('=')
		switch value := se.args[i+1].(type) {
		case string:
			sb.WriteByte('\'')
			sb.WriteString(value)
			sb.WriteByte('\'')
		default:
			sb.WriteString(fmt.Sprintf("%v", value))
		}
		sb.WriteString("]")
	}
	return sb.String()
}

func (se *sError) Args(args ...any) SError {
	se.chkArgs(len(args))
	se.args = args
	return se
}

func (se *sError) Err(err error, args ...any) SError {
	se.err = err
	if len(args) > 0 {
		se.chkArgs(len(args))
		if se.args == nil {
			se.args = args
		} else {
			se.args = append(se.args, args)
		}
	}
	return se
}

func (se *sError) chkArgs(count int) {
	if count%2 != 0 {
		panicf("SError.Args() for '%s' must receive key-value pairs for args; received %d args instead",
			se.error.Error(), count)
	}
}

func (se *sError) Is(err error) bool {
	//goland:noinspection GoDirectComparisonOfErrors
	return se.error == err
}

func (se *sError) Unwrap() error {
	return se.err
}

func panicf(msg string, args ...any) {
	panic(fmt.Sprintf(msg, args...))
}

func Diff(s1, s2 string, n int) (_, _ string, start, end int) {

	// Convert strings to local byte slices for immutability
	b1 := []byte(s1)
	b2 := []byte(s2)

	// Scan from the beginning and look for the first runes that are not the same.
	// Continue slicing each rune off of both strings until you find a pair that are
	// different or that the byte slices are empty.
	for len(b1) > 0 && len(b2) > 0 {
		ch1, width1 := utf8.DecodeRune(b1)
		ch2, width2 := utf8.DecodeRune(b2)
		if ch1 != ch2 {
			break
		}
		b1 = b1[width1:]
		b2 = b2[width2:]
		start++
	}

	// If both byte slices are empty, the strings were the same and no need to
	// continue.
	if len(b1)+len(b2) == 0 {
		s1 = ""
		s2 = ""
		goto end
	}

	// Now scan from the end and look for the last runes that are not the same.
	// Continue slicing each rune off the end of both strings until you find a pair
	// that are different or that the byte slices are empty.
	// If you get to within `n` of each end, stop.
	for len(b1) > 0 && len(b2) > 0 {
		ch1, width1 := utf8.DecodeLastRune(b1)
		ch2, width2 := utf8.DecodeLastRune(b2)
		if ch1 != ch2 {
			break
		}
		b1 = b1[:len(b1)-width1]
		b2 = b2[:len(b2)-width2]
		end++
	}
	s1 = string(b1)
	s2 = string(b2)
	if len(b1) > n {
		s1 = Excerpt(s1, n)
	}
	if len(b2) > n {
		s2 = Excerpt(s2, n)
	}

end:
	return s1, s2, start, end
}

func ExcerptWithLen(s string, width int) string {
	return fmt.Sprintf(LengthPrefixFormat, len(s), Excerpt(s, width))
}

func Excerpt(s string, width int) string {
	var prefix, suffix int

	cnt := utf8.RuneCountInString(s)
	if cnt <= width {
		// String is shorter than allocated width. Clearly, there is no need to excerpt.
		goto end
	}

	// Start with half of the allocated width
	prefix = width / 2
	// Suffix also gets half
	suffix = prefix
	if cnt >= width && width%2 == 0 {
		// If the string is longer than we allocated width, and it is not an ODD width,
		// shave off one character for the ellipsis. For an ODD width the ellipses is
		// handled by the int truncation when divided by 2, leaving 1 remainder so no
		// need to subtract one. (P.S. If string is shorter than allocated width, we
		// don't need to make room for the ellipsis.🤷‍)
		suffix--
	}
	// Get the prefix runes, the suffix runes, and insert the ellipses rune in the middle.
	s = fmt.Sprintf(ExcerptFormat,
		prefixRunes(s, prefix),
		EllipsisRune,
		suffixRunes(s, suffix),
	)
end:
	return s
}

func prefixRunes(input string, n int) string {
	result := make([]rune, 0)
	for i, r := range input {
		if i >= n {
			break
		}
		result = append(result, r)
	}
	return string(result)
}

func suffixRunes(input string, n int) string {
	var result []rune
	b := []byte(input)
	count := 0
	for count < n && len(b) > 0 {
		r, size := utf8.DecodeLastRune(b)
		result = append(result, r)
		count++
		b = b[:len(b)-size]
	}
	slices.Reverse(result)
	return string(result)
}

func DevNull(x ...any) {
	// Do nothing
}
