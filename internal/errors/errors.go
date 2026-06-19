package errors

import (
	"fmt"
	"strings"
)

// Level is the severity of a diagnostic
type Level int

const (
	LevelError Level = iota
	LevelWarning
	LevelInfo
)

func (l Level) String() string {
	switch l {
	case LevelError:
		return "error"
	case LevelWarning:
		return "warning"
	case LevelInfo:
		return "info"
	}
	return "?"
}

// Diagnostic is a single compiler message
type Diagnostic struct {
	Level   Level
	Message string
	File    string
	Line    int
	Col     int
	Source  string // the source line (if available)
	Hint    string // suggested fix
}

// Reporter collects and formats diagnostics
type Reporter struct {
	Diagnostics []Diagnostic
	sourceLines map[string][]string
}

func NewReporter() *Reporter {
	return &Reporter{
		sourceLines: map[string][]string{},
	}
}

func (r *Reporter) LoadSource(file, content string) {
	r.sourceLines[file] = strings.Split(content, "\n")
}

func (r *Reporter) Error(file string, line, col int, msg string) {
	source := r.getLine(file, line)
	r.Diagnostics = append(r.Diagnostics, Diagnostic{
		Level:   LevelError,
		Message: msg,
		File:    file,
		Line:    line,
		Col:     col,
		Source:  source,
	})
}

func (r *Reporter) ErrorHint(file string, line, col int, msg, hint string) {
	source := r.getLine(file, line)
	r.Diagnostics = append(r.Diagnostics, Diagnostic{
		Level:   LevelError,
		Message: msg,
		File:    file,
		Line:    line,
		Col:     col,
		Source:  source,
		Hint:    hint,
	})
}

func (r *Reporter) Warning(file string, line, col int, msg string) {
	source := r.getLine(file, line)
	r.Diagnostics = append(r.Diagnostics, Diagnostic{
		Level:   LevelWarning,
		Message: msg,
		File:    file,
		Line:    line,
		Col:     col,
		Source:  source,
	})
}

func (r *Reporter) WarningHint(file string, line, col int, msg, hint string) {
	source := r.getLine(file, line)
	r.Diagnostics = append(r.Diagnostics, Diagnostic{
		Level:   LevelWarning,
		Message: msg,
		File:    file,
		Line:    line,
		Col:     col,
		Source:  source,
		Hint:    hint,
	})
}

func (r *Reporter) WarningCount() int {
	count := 0
	for _, d := range r.Diagnostics {
		if d.Level == LevelWarning {
			count++
		}
	}
	return count
}

func (r *Reporter) HasErrors() bool {
	for _, d := range r.Diagnostics {
		if d.Level == LevelError {
			return true
		}
	}
	return false
}

func (r *Reporter) ErrorCount() int {
	count := 0
	for _, d := range r.Diagnostics {
		if d.Level == LevelError {
			count++
		}
	}
	return count
}

// Format produces human-friendly output (like Rust/Elm style errors)
func (r *Reporter) Format(useColor bool) string {
	var sb strings.Builder
	for _, d := range r.Diagnostics {
		sb.WriteString(r.formatDiag(d, useColor))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (r *Reporter) formatDiag(d Diagnostic, color bool) string {
	var sb strings.Builder

	// ── Header ──
	prefix := d.Level.String()
	if color {
		switch d.Level {
		case LevelError:
			prefix = "\033[1;31merror\033[0m"
		case LevelWarning:
			prefix = "\033[1;33mwarning\033[0m"
		case LevelInfo:
			prefix = "\033[1;34minfo\033[0m"
		}
	}

	sb.WriteString(fmt.Sprintf("%s: %s\n", prefix, d.Message))

	// ── Location ──
	if d.File != "" {
		loc := fmt.Sprintf("  --> %s", d.File)
		if d.Line > 0 {
			loc += fmt.Sprintf(":%d", d.Line)
		}
		if d.Col > 0 {
			loc += fmt.Sprintf(":%d", d.Col)
		}
		sb.WriteString(loc + "\n")
	}

	// ── Source snippet ──
	if d.Source != "" && d.Line > 0 {
		lineNum := fmt.Sprintf("%d", d.Line)
		pad := strings.Repeat(" ", len(lineNum))

		sb.WriteString(fmt.Sprintf("   %s |\n", pad))
		sb.WriteString(fmt.Sprintf("  %s | %s\n", lineNum, d.Source))

		// Caret indicator
		if d.Col > 0 {
			caretPos := d.Col - 1
			if caretPos < 0 {
				caretPos = 0
			}
			caret := strings.Repeat(" ", caretPos) + "^"
			if color {
				switch d.Level {
				case LevelError:
					caret = strings.Repeat(" ", caretPos) + "\033[1;31m^\033[0m"
				case LevelWarning:
					caret = strings.Repeat(" ", caretPos) + "\033[1;33m^\033[0m"
				}
			}
			sb.WriteString(fmt.Sprintf("   %s | %s\n", pad, caret))
		}
	}

	// ── Hint ──
	if d.Hint != "" {
		hintLabel := "hint"
		if color {
			hintLabel = "\033[1;32mhint\033[0m"
		}
		sb.WriteString(fmt.Sprintf("   = %s: %s\n", hintLabel, d.Hint))
	}

	return sb.String()
}

func (r *Reporter) getLine(file string, line int) string {
	lines, ok := r.sourceLines[file]
	if !ok || line <= 0 || line > len(lines) {
		return ""
	}
	return lines[line-1]
}

// ─── Common error messages ─────────────────────────────────────────────────────

func MsgUndefinedVar(name string) string {
	return fmt.Sprintf("undefined variable %q — did you forget to declare it with 'let'?", name)
}

func MsgTypeMismatch(expected, got string) string {
	return fmt.Sprintf("type mismatch: expected %s but got %s", expected, got)
}

func MsgMissingBrace(context string) string {
	return fmt.Sprintf("missing closing } after %s block", context)
}

func MsgUnexpectedToken(got, expected string) string {
	return fmt.Sprintf("unexpected %q — expected %s", got, expected)
}

func MsgUnterminatedString() string {
	return "string is not closed — did you forget the closing \"?"
}

func MsgFnReturnMissing(fnName string) string {
	return fmt.Sprintf("function %q should return a value but has no return statement", fnName)
}
