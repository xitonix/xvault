package assert

import (
	"fmt"
	"strings"
)

type Fields map[string]interface{}

func (f Fields) String() string {
	if len(f) == 0 {
		return ""
	}
	var parts []string
	for k, v := range f {
		parts = append(parts, fmt.Sprintf("%v:%v", k, v))
	}
	return strings.Join(parts, ",")
}
