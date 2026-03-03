// Package wantai provides utilities for UTC nanosecond timestamps and general date format conversion.
package wantai

import (
	"fmt"
	"regexp"
	"strings"
)

// fcKind identifies the type of a single format component.
type fcKind int

const (
	fcFixed         fcKind = iota // literal string
	fcYear4                       // YYYY
	fcYear2                       // YY
	fcMonthFull                   // MMMM → January
	fcMonthAbbr                   // MMM  → Jan
	fcMonthPadded                 // MM   → 01
	fcMonthCompact                // M    → 1
	fcDayPadded                   // DD   → 02
	fcDayCompact                  // D    → 2
	fcWeekdayFull                 // dddd → Monday
	fcWeekdayAbbr                 // ddd  → Mon
	fcHour24                      // HH   → 00-23
	fcHour12Padded                // hh   → 01-12
	fcHour12Compact               // h    → 1-12
	fcMinutePadded                // mm / ii / II → 00-59
	fcMinuteCompact               // m / i / I   → 0-59
	fcSecondPadded                // ss / SS → 00-59
	fcSecondCompact               // s / S   → 0-59
	fcAmPmUpper                   // A → AM/PM
	fcAmPmLower                   // a → am/pm
	fcTzOffset                    // ZZ → +0900
	fcTzName                      // Z  → JST
	fcFracCS                      // cc       → centiseconds (2 digits)
	fcFracMS                      // SSS      → milliseconds, digits 1-3
	fcFracUS                      // fff      → sub-ms micro part, digits 4-6
	fcFracNS                      // nnn      → sub-us nano part,  digits 7-9
	fcFracTotal6                  // ffffff   → full microseconds (digits 1-6)
	fcFracTotal9                  // nnnnnnnnn → full nanoseconds (digits 1-9)
)

// formatComponent is a single parsed piece of a GeneralDateFormat string.
type formatComponent struct {
	kind  fcKind
	fixed string // used only when kind == fcFixed
}

// GeneralDateFormat holds a pre-parsed date format string ready for efficient rendering.
type GeneralDateFormat struct {
	format     string
	components []formatComponent
}

// NewGeneralDateFormat creates a GeneralDateFormat from a common date format string
// (e.g. "YYYY-MM-DD HH:mm:ss"). Returns an error if the format is invalid.
func NewGeneralDateFormat(format string) (*GeneralDateFormat, error) {
	comps, err := convertDateFormatToComponents(format)
	if err != nil {
		return nil, err
	}
	return &GeneralDateFormat{format: format, components: comps}, nil
}

// String implements fmt.Stringer.
// It returns the original format string (e.g. "YYYY-MM-DD HH:mm:ss").
func (gdf *GeneralDateFormat) String() string {
	return gdf.format
}

// ---------------------------------------------------------------------------
// Tokenizer
// ---------------------------------------------------------------------------

// tokenTable maps format prefixes to component kinds (longest match first).
var tokenTable = []struct {
	prefix string
	kind   fcKind
}{
	{"nnnnnnnnn", fcFracTotal9},
	{"ffffff", fcFracTotal6},
	{"MMMM", fcMonthFull},
	{"dddd", fcWeekdayFull},
	{"YYYY", fcYear4},
	{"MMM", fcMonthAbbr},
	{"ddd", fcWeekdayAbbr},
	{"SSS", fcFracMS},
	{"fff", fcFracUS},
	{"nnn", fcFracNS},
	{"HH", fcHour24},
	{"hh", fcHour12Padded},
	{"mm", fcMinutePadded},
	{"ii", fcMinutePadded},
	{"II", fcMinutePadded},
	{"ss", fcSecondPadded},
	{"SS", fcSecondPadded},
	{"DD", fcDayPadded},
	{"MM", fcMonthPadded},
	{"cc", fcFracCS},
	{"YY", fcYear2},
	{"ZZ", fcTzOffset},
	{"M", fcMonthCompact},
	{"D", fcDayCompact},
	{"h", fcHour12Compact},
	{"m", fcMinuteCompact},
	{"i", fcMinuteCompact},
	{"I", fcMinuteCompact},
	{"s", fcSecondCompact},
	{"S", fcSecondCompact},
	{"A", fcAmPmUpper},
	{"a", fcAmPmLower},
	{"Z", fcTzName},
}

func matchToken(s string) (formatComponent, int, bool) {
	for _, t := range tokenTable {
		if strings.HasPrefix(s, t.prefix) {
			return formatComponent{kind: t.kind}, len(t.prefix), true
		}
	}
	return formatComponent{}, 0, false
}

// appendFixed appends a literal string, merging into the previous fcFixed component if possible.
func appendFixed(comps []formatComponent, s string) []formatComponent {
	if len(comps) > 0 && comps[len(comps)-1].kind == fcFixed {
		comps[len(comps)-1].fixed += s
		return comps
	}
	return append(comps, formatComponent{kind: fcFixed, fixed: s})
}

// convertDateFormatToComponents parses a format string into a slice of formatComponents.
//
// Escape semantics (same as before):
//   - \\ → literal backslash
//   - \<any> → literal <any>
//   - 'text' → literal text
func convertDateFormatToComponents(format string) ([]formatComponent, error) {
	escapedBlocks := make(map[string]string)

	nextID := func() string {
		return fmt.Sprintf("\x00%d\x00", len(escapedBlocks))
	}

	// Step 1: Process backslash escapes.
	var backslashPass strings.Builder
	runes := []rune(format)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) {
			i++
			id := nextID()
			escapedBlocks[id] = string(runes[i])
			backslashPass.WriteString(id)
		} else {
			backslashPass.WriteRune(runes[i])
		}
	}

	// Step 2: Replace single-quoted literal strings with placeholders.
	re := regexp.MustCompile(`'([^']+)'`)
	processed := re.ReplaceAllStringFunc(backslashPass.String(), func(s string) string {
		id := nextID()
		escapedBlocks[id] = s[1 : len(s)-1]
		return id
	})

	// Step 2.5: Resolve nested placeholders (e.g. a quote block that captured a
	// backslash-escaped character placeholder inside it).
	// Restore in descending order so inner placeholders are expanded first.
	for i := len(escapedBlocks) - 1; i >= 0; i-- {
		id := fmt.Sprintf("\x00%d\x00", i)
		if v, ok := escapedBlocks[id]; ok {
			// Expand any nested placeholders within this value.
			expanded := v
			for j := i - 1; j >= 0; j-- {
				inner := fmt.Sprintf("\x00%d\x00", j)
				if w, ok2 := escapedBlocks[inner]; ok2 {
					expanded = strings.ReplaceAll(expanded, inner, w)
				}
			}
			escapedBlocks[id] = expanded
		}
	}

	// Step 3: Tokenize.
	var result []formatComponent
	rs := []rune(processed)
	i := 0
	for i < len(rs) {
		// Detect NUL-delimited placeholder.
		if rs[i] == '\x00' {
			j := i + 1
			for j < len(rs) && rs[j] != '\x00' {
				j++
			}
			if j < len(rs) {
				id := string(rs[i : j+1])
				result = appendFixed(result, escapedBlocks[id])
				i = j + 1
				continue
			}
		}

		// Try longest-match token.
		remaining := string(rs[i:])
		if fc, n, ok := matchToken(remaining); ok {
			result = append(result, fc)
			i += n
			continue
		}

		// Pass-through character.
		result = appendFixed(result, string(rs[i]))
		i++
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// Calendar arithmetic
// ---------------------------------------------------------------------------

var monthNames = [12]string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}
var monthAbbrs = [12]string{
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}
var weekdayNames = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var weekdayAbbrs = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// daysToDate converts days-since-Unix-epoch to (year, month [1-12], day [1-31], weekday [0=Sun]).
// Uses Howard Hinnant's civil_from_days algorithm.
func daysToDate(days int64) (year, month, day, wday int) {
	// Weekday: epoch (1970-01-01) was Thursday (4 in 0=Sun convention).
	w := (days + 4) % 7
	if w < 0 {
		w += 7
	}
	wday = int(w)

	// Shift to days since 0000-03-01.
	z := days + 719468
	era := z / 146097
	if z < 0 && z%146097 != 0 {
		era--
	}
	doe := z - era*146097                                  // day-of-era   [0, 146096]
	yoe := (doe - doe/1460 + doe/36524 - doe/146096) / 365 // year-of-era  [0, 399]
	y := yoe + era*400
	doy := doe - (365*yoe + yoe/4 - yoe/100) // day-of-year  [0, 365]
	mp := (5*doy + 2) / 153                  // month-period [0, 11]
	day = int(doy-(153*mp+2)/5) + 1          // [1, 31]
	month = int(mp) + 3
	if month > 12 {
		month -= 12
	}
	year = int(y)
	if month <= 2 {
		year++
	}
	return
}

// ---------------------------------------------------------------------------
// Renderer
// ---------------------------------------------------------------------------

// render formats the timestamp using the pre-parsed components.
//
//   - adjustedSec: Unix seconds adjusted for the target timezone offset
//   - absNano:     nanoseconds within the current second (0–999_999_999)
//   - tzOffsetSec: timezone offset in seconds (used for ZZ token)
//   - tzName:      timezone abbreviation (used for Z token)
func (gdf *GeneralDateFormat) render(adjustedSec, absNano int64, tzOffsetSec int, tzName string) string {
	// --- Calendar fields ---
	days := adjustedSec / 86400
	rem := adjustedSec % 86400
	if rem < 0 {
		days--
		rem += 86400
	}
	year, month, day, wday := daysToDate(days)
	hour := int(rem / 3600)
	minute := int((rem % 3600) / 60)
	second := int(rem % 60)

	// --- Fractional second fields ---
	cs := int(absNano / 10_000_000)              // centiseconds 0-99
	ms := int(absNano / 1_000_000)               // milliseconds 0-999
	fracUS := int((absNano % 1_000_000) / 1_000) // micro sub-part 0-999
	fracNS := int(absNano % 1_000)               // nano sub-part  0-999

	// --- AM/PM ---
	h12 := hour % 12
	if h12 == 0 {
		h12 = 12
	}

	var sb strings.Builder
	for _, fc := range gdf.components {
		switch fc.kind {
		case fcFixed:
			sb.WriteString(fc.fixed)
		case fcYear4:
			fmt.Fprintf(&sb, "%04d", year)
		case fcYear2:
			fmt.Fprintf(&sb, "%02d", year%100)
		case fcMonthFull:
			sb.WriteString(monthNames[month-1])
		case fcMonthAbbr:
			sb.WriteString(monthAbbrs[month-1])
		case fcMonthPadded:
			fmt.Fprintf(&sb, "%02d", month)
		case fcMonthCompact:
			fmt.Fprintf(&sb, "%d", month)
		case fcDayPadded:
			fmt.Fprintf(&sb, "%02d", day)
		case fcDayCompact:
			fmt.Fprintf(&sb, "%d", day)
		case fcWeekdayFull:
			sb.WriteString(weekdayNames[wday])
		case fcWeekdayAbbr:
			sb.WriteString(weekdayAbbrs[wday])
		case fcHour24:
			fmt.Fprintf(&sb, "%02d", hour)
		case fcHour12Padded:
			fmt.Fprintf(&sb, "%02d", h12)
		case fcHour12Compact:
			fmt.Fprintf(&sb, "%d", h12)
		case fcMinutePadded:
			fmt.Fprintf(&sb, "%02d", minute)
		case fcMinuteCompact:
			fmt.Fprintf(&sb, "%d", minute)
		case fcSecondPadded:
			fmt.Fprintf(&sb, "%02d", second)
		case fcSecondCompact:
			fmt.Fprintf(&sb, "%d", second)
		case fcAmPmUpper:
			if hour < 12 {
				sb.WriteString("AM")
			} else {
				sb.WriteString("PM")
			}
		case fcAmPmLower:
			if hour < 12 {
				sb.WriteString("am")
			} else {
				sb.WriteString("pm")
			}
		case fcTzOffset:
			sign := "+"
			off := tzOffsetSec
			if off < 0 {
				sign = "-"
				off = -off
			}
			fmt.Fprintf(&sb, "%s%02d%02d", sign, off/3600, (off%3600)/60)
		case fcTzName:
			sb.WriteString(tzName)
		case fcFracCS:
			fmt.Fprintf(&sb, "%02d", cs)
		case fcFracMS:
			fmt.Fprintf(&sb, "%03d", ms)
		case fcFracUS:
			fmt.Fprintf(&sb, "%03d", fracUS)
		case fcFracNS:
			fmt.Fprintf(&sb, "%03d", fracNS)
		case fcFracTotal6:
			fmt.Fprintf(&sb, "%06d", ms*1000+fracUS)
		case fcFracTotal9:
			fmt.Fprintf(&sb, "%09d", ms*1_000_000+fracUS*1000+fracNS)
		}
	}
	return sb.String()
}
