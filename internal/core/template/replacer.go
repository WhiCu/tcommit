package template

type Replacer interface {
	Get(key string) (string, bool)
}

type ReplacerFunc func(key string) (string, bool)

func (f ReplacerFunc) Get(key string) (string, bool) {
	return f(key)
}

func ReplacerFuncFromMap(m map[string]string) ReplacerFunc {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}
