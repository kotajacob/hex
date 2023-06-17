package hb

import (
	"html/template"
	"strings"
	"time"
)

// Timestamp implements a fancy HTML timstamp renderer for the hb types.
func Timestamp(i interface{}) template.HTML {
	switch v := i.(type) {
	case Comment:
		var b strings.Builder
		if v.Updated != nil {
			b.WriteString("<time title=\"")
			b.WriteString(v.Updated.String())
			b.WriteString("\">")
			b.WriteString(v.Updated.Since(time.Now()))
		} else {
			b.WriteString("<time title=\"")
			b.WriteString(v.Published.String())
			b.WriteString("\">")
			b.WriteString(v.Published.Since(time.Now()))
		}
		b.WriteString("</time>")
		return template.HTML(b.String())
	case *Comment:
		var b strings.Builder
		if v.Updated != nil {
			b.WriteString("<time title=\"")
			b.WriteString(v.Updated.String())
			b.WriteString("\">")
			b.WriteString(v.Updated.Since(time.Now()))
		} else {
			b.WriteString("<time title=\"")
			b.WriteString(v.Published.String())
			b.WriteString("\">")
			b.WriteString(v.Published.Since(time.Now()))
		}
		b.WriteString("</time>")
		return template.HTML(b.String())
	case Post:
		var b strings.Builder
		if v.Updated != nil {
			b.WriteString("<time title=\"")
			b.WriteString(v.Updated.String())
			b.WriteString("\">")
			b.WriteString(v.Updated.Since(time.Now()))
		} else {
			b.WriteString("<time title=\"")
			b.WriteString(v.Published.String())
			b.WriteString("\">")
			b.WriteString(v.Published.Since(time.Now()))
		}
		b.WriteString("</time>")
		return template.HTML(b.String())
	default:
		return ""
	}
}
