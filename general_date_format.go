// Package wantai provides utilities for UTC nanosecond timestamps and general date format conversion.
package wantai

import (
	"fmt"
	"regexp"
	"strings"
)

// GeneralDateFormat holds a common date format string (e.g. "YYYY-MM-DD HH:mm:ss")
// and converts it to a layout string compatible with Go's time.Format.
type GeneralDateFormat struct {
	format       string
	goStyleCache string
}

// NewGeneralDateFormat creates a GeneralDateFormat from a common date format string
// (e.g. "YYYY-MM-DD HH:mm:ss"). Returns an error if the format conversion fails.
func NewGeneralDateFormat(format string) (*GeneralDateFormat, error) {
	goStyleCache, err := convertDateFormatToGoLayout(format)
	if err != nil {
		return nil, err
	}
	return &GeneralDateFormat{
		format,
		goStyleCache,
	}, nil
}

// GoLayout returns the layout string that can be passed to Go's time.Format.
func (gdf *GeneralDateFormat) GoLayout() string {
	return gdf.goStyleCache
}

// String implements fmt.Stringer.
// It returns the original format string (e.g. "YYYY-MM-DD HH:mm:ss").
func (gdf *GeneralDateFormat) String() string {
	return gdf.format
}

func convertDateFormatToGoLayout(format string) (string, error) {
	escapedBlocks := make(map[string]string)

	// nextID returns a unique NUL-delimited placeholder that cannot be
	// produced by any date format token, preventing accidental re-substitution.
	nextID := func() string {
		return fmt.Sprintf("\x00%d\x00", len(escapedBlocks))
	}

	// Step 1: Process backslash escapes.
	// \\ → literal backslash
	// \<any> → literal <any character>
	// This runs before single-quote parsing so that \' can appear inside
	// or outside quoted blocks.
	var backslashPass strings.Builder
	runes := []rune(format)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) {
			i++ // consume the next character as a literal
			id := nextID()
			escapedBlocks[id] = string(runes[i])
			backslashPass.WriteString(id)
		} else {
			backslashPass.WriteRune(runes[i])
		}
	}

	// Step 2: Replace single-quoted literal strings with NUL-based placeholders.
	// NUL (\x00) is never present in date format tokens,
	// so the replacer cannot accidentally transform the placeholder.
	re := regexp.MustCompile(`'([^']+)'`)
	processedLayout := re.ReplaceAllStringFunc(backslashPass.String(), func(s string) string {
		id := nextID()
		escapedBlocks[id] = s[1 : len(s)-1]
		return id
	})

	// Step 3: Replace format tokens with Go layout equivalents.
	replacer := strings.NewReplacer(
		// Year
		"YYYY", "2006",
		"YY", "06",
		// Month
		"MMMM", "January",
		"MMM", "Jan",
		"MM", "01",
		"M", "1",
		// Day
		"DD", "02",
		"D", "2",
		// Day of Week
		"dddd", "Monday",
		"ddd", "Mon",
		// Hour
		"HH", "15",
		"hh", "03",
		"h", "3",
		// Minute
		"mm", "04",
		"ii", "04",
		"II", "04",
		"m", "4",
		"i", "4",
		"I", "4",
		// Fractional second (longer tokens must come before shorter ones)
		"SSS", "000",
		"ffffff", "000000",
		"nnnnnn", "000000000",
		"cc", "00",
		// Second
		"ss", "05",
		"SS", "05",
		"s", "5",
		"S", "5",
		// AM/PM
		"A", "PM",
		"a", "pm",
		// Timezone
		"ZZ", "-0700",
		"Z", "MST",
	)

	result := replacer.Replace(processedLayout)

	// Step 4: Restore all escaped literal blocks in reverse insertion order.
	// Placeholders are numbered sequentially (0, 1, 2, ...).
	// Restoring in descending order ensures that a placeholder whose value
	// contains another placeholder (e.g. a quote block enclosing a \' escape)
	// is expanded first, allowing inner placeholders to be resolved afterwards.
	for i := len(escapedBlocks) - 1; i >= 0; i-- {
		id := fmt.Sprintf("\x00%d\x00", i)
		result = strings.ReplaceAll(result, id, escapedBlocks[id])
	}

	return result, nil
}
