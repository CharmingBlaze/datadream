package typecheck

import (
	"fmt"
	"sort"
	"strings"
)

func (c *checker) hintUnknownIdentifier(name string) string {
	if c.fns[name] {
		return fmt.Sprintf("function %q is declared — call it as %s(...)", name, name)
	}
	if namespaceRoots[name] {
		return fmt.Sprintf("%q is a namespace — use %s.method(...), e.g. draw.text(...)", name, name)
	}
	if suggest := c.nearestName(name, c.knownNames()); suggest != "" {
		return fmt.Sprintf("did you mean %q?", suggest)
	}
	return "declare it with `let " + name + " = ...;` before use"
}

func (c *checker) knownNames() []string {
	seen := map[string]bool{}
	var names []string
	add := func(n string) {
		if n != "" && !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	for n := range c.globals {
		add(n)
	}
	for n := range c.fns {
		add(n)
	}
	for n := range namespaceRoots {
		add(n)
	}
	for n := range builtinFns {
		add(n)
	}
	for n := range c.modules {
		add(n)
	}
	return names
}

func (c *checker) nearestName(want string, candidates []string) string {
	wantLower := strings.ToLower(want)
	var best string
	bestDist := 3
	for _, c := range candidates {
		cl := strings.ToLower(c)
		d := editDistance(wantLower, cl)
		if d < bestDist {
			bestDist = d
			best = c
		}
	}
	return best
}

func editDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

func listMapKeys(m map[string]bool) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func listMethodKeys(m map[string]methodSpec) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func hintNamespaceMethod(ns, method string) string {
	methods, ok := namespaces[ns]
	if !ok {
		return ""
	}
	if _, ok := methods[method]; ok {
		return ""
	}
	return fmt.Sprintf("available %s.* methods: %s", ns, listMethodKeys(methods))
}

func hintOptionFields(ns, method string) string {
	spec, ok := namespaces[ns][method]
	if !ok || spec.optionFields == nil {
		return "option objects use commas between fields: { position: vec2(...), size: 32 }"
	}
	return fmt.Sprintf("valid options for %s.%s: %s", ns, method, listMapKeys(spec.optionFields))
}

func hintStructFields(fields map[string]bool) string {
	if len(fields) == 0 {
		return ""
	}
	return "valid fields: " + listMapKeys(fields)
}

func argRangeHint(spec methodSpec) string {
	if spec.minArgs == spec.maxArgs {
		return fmt.Sprintf("%d argument(s)", spec.minArgs)
	}
	if spec.maxArgs < 0 {
		return fmt.Sprintf("at least %d argument(s)", spec.minArgs)
	}
	return fmt.Sprintf("%d–%d argument(s)", spec.minArgs, spec.maxArgs)
}
