package stringcompare

import (
	"regexp"
	"strings"
	"sync"
)

var (
	mu    sync.Mutex
	cache = map[string]*regexp.Regexp{}
)

const maxCached = 256

func compile(pattern string) (*regexp.Regexp, error) {
	parts := strings.Split(pattern, "*")

	var sb strings.Builder
	sb.WriteString("(?is)^")
	for i, part := range parts {
		if i > 0 {
			sb.WriteString(".*")
		}
		sb.WriteString(regexp.QuoteMeta(part))
	}
	sb.WriteString("$")

	return regexp.Compile(sb.String())
}

func matcher(pattern string) (*regexp.Regexp, error) {
	mu.Lock()
	defer mu.Unlock()

	if re, ok := cache[pattern]; ok {
		return re, nil
	}
	re, err := compile(pattern)
	if err != nil {
		return nil, err
	}
	if len(cache) >= maxCached {
		cache = map[string]*regexp.Regexp{}
	}
	cache[pattern] = re
	return re, nil
}

func CompareWithWildcard(str string, pattern string) (bool, error) {
	re, err := matcher(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(str), nil
}
