package typecheck

import (
	"fmt"

	"datadream/internal/ast"
)

func (c *checker) registerUseWhitelist(path string, symbols []string) {
	if c.useWhitelist == nil {
		c.useWhitelist = map[string]map[string]bool{}
	}
	if len(symbols) == 0 {
		c.useWhitelist[path] = nil
		return
	}
	allowed := map[string]bool{}
	for _, s := range symbols {
		allowed[s] = true
	}
	c.useWhitelist[path] = allowed
}

func (c *checker) hasSelectiveUsingImport() bool {
	for _, mod := range c.usingMods {
		if wl, ok := c.useWhitelist[mod]; ok && wl != nil {
			return true
		}
	}
	return false
}

func (c *checker) symbolAllowedUnqualified(name string) bool {
	for _, mod := range c.usingMods {
		wl, ok := c.useWhitelist[mod]
		if !ok || wl == nil {
			if isRaylibSymbol(name) || isRaylibConstant(name) {
				return true
			}
			continue
		}
		if wl[name] {
			return true
		}
	}
	return false
}

func (c *checker) checkSelectiveImport(name string, pos ast.Position) {
	if !c.usesRaylib || !c.hasSelectiveUsingImport() {
		return
	}
	if c.fns[name] || namespaceRoots[name] || builtinFns[name] {
		return
	}
	if _, ok := c.lookup(name); ok {
		return
	}
	if isRaylibSymbol(name) || isRaylibConstant(name) {
		if !c.symbolAllowedUnqualified(name) {
			c.errorAtHint(pos,
				fmt.Sprintf("symbol %q is not imported", name),
				"add it to the use raylib { ... } list")
		}
	}
}

func (c *checker) importPathFor(name string) string {
	if name == "raylib" || name == "graphics" {
		return name
	}
	for path := range c.modules {
		if path == name {
			return path
		}
	}
	return name
}

func isRaylibConstant(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if r != '_' && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}
