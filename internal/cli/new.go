package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var projectNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

func cmdNew(args []string) int {
	name, _, ok := requireArg(args, "Usage: datadream new <project-name>")
	if !ok {
		return 1
	}

	if !projectNamePattern.MatchString(name) {
		die("project name must start with a letter and contain only letters, digits, hyphens, or underscores")
		return 1
	}

	dir, err := filepath.Abs(name)
	if err != nil {
		die("invalid path: " + err.Error())
		return 1
	}

	if _, err := os.Stat(dir); err == nil {
		die("directory already exists: " + name)
		return 1
	} else if !os.IsNotExist(err) {
		die("cannot create project: " + err.Error())
		return 1
	}

	title := projectTitle(name)

	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0755); err != nil {
		die("cannot create assets/: " + err.Error())
		return 1
	}

	files := map[string]string{
		"game.dd":     gameTemplate(title),
		"README.md":   readmeTemplate(name, title),
		"assets/.gitkeep": "",
	}

	for rel, content := range files {
		path := filepath.Join(dir, rel)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			die("cannot write " + rel + ": " + err.Error())
			return 1
		}
	}

	fmt.Printf("✓ Created project %q\n", name)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  datadream run game.dd")
	return 0
}

func projectTitle(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, " ")
}

func gameTemplate(title string) string {
	return fmt.Sprintf(`// %s — edit game.dd and run: datadream run game.dd

app "%s";

window {
    size: 800, 600;
    title: "%s";
    fps: 60;
}

draw {
    clear(colors.black);

    draw.text("Hello, DataDream!", {
        position: vec2(220, 280),
        size: 32,
        color: colors.white
    });
}
`, title, title, title)
}

func readmeTemplate(name, title string) string {
	return fmt.Sprintf(`# %s

A DataDream game project.

## Run

`+"```"+`bash
datadream run game.dd
`+"```"+`

## Build

`+"```"+`bash
datadream build game.dd -o %s
`+"```"+`

Put images and sounds in `+"`assets/`"+`.

See [DataDream docs](https://github.com/CharmingBlaze/datadream/tree/main/docs) for tutorials and examples.
`, title, name)
}
