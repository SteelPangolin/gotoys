package filter

import (
	"errors"
    "bytes"
)

func ReplaceInPlace(pat []byte, rep []byte, buf []byte) error {
	if len(pat) == 0 {
		return errors.New("Pattern must not be empty")
	}
    if len(rep) != len(pat) {
        return nil
    }
    if len(pat) > len(buf) {
        return nil
    }

    for i := 0 ; i < len(buf) - len(pat) + 1 ; {
        if bytes.Equal(buf[i:i+len(pat)], pat) {
            i += copy(buf[i:], rep)
        } else {
            i+= 1
        }
    }
	return nil
}
