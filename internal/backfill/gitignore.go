package backfill

import (
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type ignoreRule struct {
	dir     string
	negate  bool
	dirOnly bool
	re      *regexp.Regexp
}

type treeIgnore struct {
	byDir map[string][]ignoreRule
}

func newTreeIgnore() *treeIgnore {
	return &treeIgnore{byDir: make(map[string][]ignoreRule)}
}

func (t *treeIgnore) add(dir string, rules []ignoreRule) {
	if dir == "." {
		dir = ""
	}
	t.byDir[dir] = append(t.byDir[dir], rules...)
}

func (t *treeIgnore) ignored(rel string, isDir bool) bool {
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." {
		return false
	}
	parts := strings.Split(rel, "/")
	acc := ""
	for i := 0; i < len(parts)-1; i++ {
		if acc == "" {
			acc = parts[i]
		} else {
			acc += "/" + parts[i]
		}
		if t.matched(acc, true) {
			return true
		}
	}
	return t.matched(rel, isDir)
}

func (t *treeIgnore) matched(rel string, isDir bool) bool {
	ignored := false
	parts := strings.Split(rel, "/")
	dirs := []string{""}
	for i := 0; i < len(parts)-1; i++ {
		dirs = append(dirs, strings.Join(parts[:i+1], "/"))
	}
	for _, d := range dirs {
		for _, rule := range t.byDir[d] {
			if rule.matches(rel, isDir) {
				ignored = !rule.negate
			}
		}
	}
	return ignored
}

func (r ignoreRule) matches(fullRel string, isDir bool) bool {
	if r.dirOnly && !isDir {
		return false
	}
	name := fullRel
	if r.dir != "" {
		if fullRel == r.dir || !strings.HasPrefix(fullRel, r.dir+"/") {
			return false
		}
		name = fullRel[len(r.dir)+1:]
	}
	return r.re.MatchString(name)
}

func parseGitignore(dir, content string) ([]ignoreRule, error) {
	dir = filepath.ToSlash(dir)
	if dir == "." {
		dir = ""
	}
	var rules []ignoreRule
	sc := bufio.NewScanner(strings.NewReader(content))
	first := true
	lineNo := 0
	for sc.Scan() {
		line := sc.Text()
		lineNo++
		if first {
			line = strings.TrimPrefix(line, "\ufeff")
			first = false
		}
		line = normalizeIgnoreLine(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rule, err := compileRule(dir, line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if rule.re == nil {
			continue
		}
		rules = append(rules, rule)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

func normalizeIgnoreLine(line string) string {
	line = strings.TrimRight(line, "\r")
	if strings.HasSuffix(line, `\ `) {
		return strings.TrimSuffix(line, `\ `) + " "
	}
	return strings.TrimRight(line, " \t")
}

func compileRule(dir, line string) (ignoreRule, error) {
	rule := ignoreRule{dir: dir}
	if strings.HasPrefix(line, `\!`) {
		line = line[1:]
	} else if strings.HasPrefix(line, "!") {
		rule.negate = true
		line = line[1:]
	}
	if line == "" {
		return ignoreRule{}, nil
	}
	if strings.HasSuffix(line, "/") {
		rule.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}
	leadingSlash := strings.HasPrefix(line, "/")
	line = strings.TrimPrefix(line, "/")
	if line == "" {
		return ignoreRule{}, nil
	}
	anchored := leadingSlash || strings.Contains(line, "/")
	expr, err := globToRegex(line, anchored)
	if err != nil {
		return ignoreRule{}, err
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return ignoreRule{}, err
	}
	rule.re = re
	return rule, nil
}

func globToRegex(glob string, anchored bool) (string, error) {
	var b strings.Builder
	b.WriteString("^")
	if !anchored {
		b.WriteString(`(?:.*/)?`)
	}
	for i := 0; i < len(glob); {
		switch glob[i] {
		case '\\':
			if i+1 < len(glob) {
				b.WriteString(regexp.QuoteMeta(glob[i+1 : i+2]))
				i += 2
				continue
			}
			b.WriteString(`\\`)
			i++
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				atStart := i == 0
				i += 2
				hasSlashAfter := i < len(glob) && glob[i] == '/'
				atEnd := i == len(glob)
				switch {
				case atStart && hasSlashAfter:
					b.WriteString(`(?:.*/)?`)
					i++
				case atStart && atEnd:
					b.WriteString(`.*`)
				case hasSlashAfter:
					b.WriteString(`(?:.*/)?`)
					i++
				case atEnd:
					b.WriteString(`.*`)
				default:
					b.WriteString(`[^/]*`)
				}
				continue
			}
			b.WriteString(`[^/]*`)
			i++
		case '?':
			b.WriteString(`[^/]`)
			i++
		case '[':
			j, class, ok := parseCharClass(glob, i)
			if !ok {
				b.WriteString(`\[`)
				i++
				continue
			}
			b.WriteString(class)
			i = j
		default:
			b.WriteString(regexp.QuoteMeta(glob[i : i+1]))
			i++
		}
	}
	b.WriteString("$")
	return b.String(), nil
}

func parseCharClass(glob string, i int) (int, string, bool) {
	if i >= len(glob) || glob[i] != '[' {
		return i, "", false
	}
	j := i + 1
	if j < len(glob) && (glob[j] == '!' || glob[j] == '^') {
		j++
	}
	if j < len(glob) && glob[j] == ']' {
		j++
	}
	for j < len(glob) && glob[j] != ']' {
		j++
	}
	if j >= len(glob) {
		return i, "", false
	}
	inner := glob[i+1 : j]
	neg := false
	if strings.HasPrefix(inner, "!") || strings.HasPrefix(inner, "^") {
		neg = true
		inner = inner[1:]
	}
	var b strings.Builder
	b.WriteByte('[')
	if neg {
		b.WriteByte('^')
	}
	b.WriteString(inner)
	b.WriteByte(']')
	return j + 1, b.String(), true
}
