package templates

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// Jinja2Template provides jinja2-like template processing
type Jinja2Template struct {
	template *template.Template
}

// NewJinja2Template creates a new jinja2 template processor
func NewJinja2Template(templateContent string) (*Jinja2Template, error) {
	// Convert jinja2 syntax to Go template syntax
	goTemplate := convertJinja2ToGo(templateContent)

	tmpl, err := template.New("jinja2").Parse(goTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &Jinja2Template{template: tmpl}, nil
}

// Render executes the template with the given data
func (jt *Jinja2Template) Render(data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := jt.template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}

// convertJinja2ToGo converts jinja2 syntax to Go template syntax
func convertJinja2ToGo(jinja2Content string) string {
	content := jinja2Content

	// Convert {% for %} loops - use proper Go template syntax
	content = regexp.MustCompile(`\{%\s*for\s+(\w+)\s+in\s+(\w+)\s*%\}`).ReplaceAllStringFunc(content, func(match string) string {
		re := regexp.MustCompile(`\{%\s*for\s+(\w+)\s+in\s+(\w+)\s*%\}`)
		matches := re.FindStringSubmatch(match)
		if len(matches) == 3 {
			return fmt.Sprintf("{{range .%s}}{{$1 := .}}", matches[2])
		}
		return match
	})
	content = regexp.MustCompile(`\{%\s*endfor\s*%\}`).ReplaceAllString(content, "{{end}}")

	// Convert {% if %} statements
	content = regexp.MustCompile(`\{%\s*if\s+(.+?)\s*%\}`).ReplaceAllString(content, "{{if $1}}")
	content = regexp.MustCompile(`\{%\s*else\s*%\}`).ReplaceAllString(content, "{{else}}")
	content = regexp.MustCompile(`\{%\s*endif\s*%\}`).ReplaceAllString(content, "{{end}}")

	// Convert {% set %} variable assignments
	content = regexp.MustCompile(`\{%\s*set\s+(\w+)\s*=\s*['"](.+?)['"]\s*%\}`).ReplaceAllString(content, "{{$1 = \"$2\"}}")
	content = regexp.MustCompile(`\{%\s*set\s+(\w+)\s*=\s*(.+?)\s*%\}`).ReplaceAllString(content, "{{$1 = $2}}")

	// Convert field access {{ dict.field }} before variable access
	content = regexp.MustCompile(`\{\{\s*(\w+)\.(\w+)\s*\}\}`).ReplaceAllString(content, "{{.$1.$2}}")

	// Convert dictionary access {{ dict['key'] }}
	content = regexp.MustCompile(`\{\{\s*(\w+)\['(.+?)'\]\s*\}\}`).ReplaceAllString(content, "{{index .$1 \"$2\"}}")

	// Convert variable access {{ variable }}
	content = regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`).ReplaceAllString(content, "{{$1}}")

	// Convert method calls {{ string.split(...) }}
	content = regexp.MustCompile(`\{\{\s*(\w+)\.split\(['"](.+?)['"]\)\s*\}\}`).ReplaceAllString(content, "{{split $1 \"$2\"}}")

	// Convert .strip() calls
	content = regexp.MustCompile(`\{\{\s*(\w+)\.strip\(\)\s*\}\}`).ReplaceAllString(content, "{{trim $1}}")

	// Convert .upper() and .lower() calls
	content = regexp.MustCompile(`\{\{\s*(\w+)\.upper\(\)\s*\}\}`).ReplaceAllString(content, "{{upper $1}}")
	content = regexp.MustCompile(`\{\{\s*(\w+)\.lower\(\)\s*\}\}`).ReplaceAllString(content, "{{lower $1}}")

	// Convert .get() method calls
	content = regexp.MustCompile(`\{\{\s*(\w+)\.get\(['"](.+?)['"]\)\s*\}\}`).ReplaceAllString(content, "{{get $1 \"$2\"}}")

	// Convert 'in' operator
	content = regexp.MustCompile(`\{%\s*if\s+['"](.+?)['"]\s+in\s+(.+?)\s*%\}`).ReplaceAllString(content, "{{if contains $2 \"$1\"}}")

	return content
}

// Template functions for Go templates
var TemplateFuncs = template.FuncMap{
	"split": func(s, sep string) []string {
		return strings.Split(s, sep)
	},
	"trim":  strings.TrimSpace,
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	"get": func(m map[string]interface{}, key string) interface{} {
		return m[key]
	},
	"contains": func(s, substr string) bool {
		return strings.Contains(s, substr)
	},
}
