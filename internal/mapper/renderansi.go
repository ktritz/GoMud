package mapper

import (
	"fmt"
	"strings"
	"unicode"
)

// RenderANSILine converts a raw map row into ansitag-decorated output in one pass.
// This avoids replacing symbols inside tags that were already injected earlier.
func RenderANSILine(line []rune, legend map[rune]string) string {
	var out strings.Builder

	for _, sym := range line {
		txtLegend, ok := legend[sym]
		if !ok {
			out.WriteRune(sym)
			continue
		}

		alias := sanitizeLegendAlias(txtLegend)
		if alias == "" {
			out.WriteRune(sym)
			continue
		}
		out.WriteString(fmt.Sprintf(`<ansi fg="map-%s" bg="mapbg-%s">%c</ansi>`, alias, alias, sym))
	}

	return out.String()
}

func sanitizeLegendAlias(legend string) string {
	var out strings.Builder
	prevHyphen := false

	for _, r := range strings.ToLower(strings.TrimSpace(legend)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
			prevHyphen = false
			continue
		}

		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '-' || r == '_' {
			if !prevHyphen {
				out.WriteByte('-')
				prevHyphen = true
			}
		}
	}

	return strings.Trim(out.String(), "-")
}
