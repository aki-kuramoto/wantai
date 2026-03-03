package wantai

import (
	"testing"
)

func TestNewGeneralDateFormat(t *testing.T) {
	gdf, err := NewGeneralDateFormat("YYYY-MM-DD")
	if err != nil {
		t.Fatalf("NewGeneralDateFormat returned unexpected error: %v", err)
	}
	if gdf == nil {
		t.Fatal("NewGeneralDateFormat returned nil")
	}
}

func TestString_GeneralDateFormat(t *testing.T) {
	format := "YYYY-MM-DD HH:mm:ss"
	gdf, err := NewGeneralDateFormat(format)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := gdf.String(); got != format {
		t.Errorf("String() = %q, want %q", got, format)
	}
}

func TestString_GeneralDateFormat_Empty(t *testing.T) {
	gdf, err := NewGeneralDateFormat("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := gdf.String(); got != "" {
		t.Errorf("String() = %q, want empty string", got)
	}
}

// ---------------------------------------------------------------------------
// RenderWithFormat token coverage
// Uses testTimestampNano = 2024-01-15 12:34:56.123456789 UTC (Monday)
// ---------------------------------------------------------------------------

func TestRenderWithFormat_AllTokens(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		// Year
		{"YYYY", "YYYY", "2024"},
		{"YY", "YY", "24"},
		// Month
		{"MMMM", "MMMM", "January"},
		{"MMM", "MMM", "Jan"},
		{"MM", "MM", "01"},
		{"M", "M", "1"},
		// Day
		{"DD", "DD", "15"},
		{"D", "D", "15"},
		// Day of week
		{"dddd", "dddd", "Monday"},
		{"ddd", "ddd", "Mon"},
		// Hour
		{"HH (24h)", "HH", "12"},
		{"hh (12h)", "hh", "12"},
		{"h (12h)", "h", "12"},
		// Minute aliases
		{"mm", "mm", "34"},
		{"ii", "ii", "34"},
		{"II", "II", "34"},
		{"m", "m", "34"},
		{"i", "i", "34"},
		{"I", "I", "34"},
		// Second aliases
		{"ss", "ss", "56"},
		{"SS", "SS", "56"},
		{"s", "s", "56"},
		{"S", "S", "56"},
		// AM/PM
		{"A (PM)", "A", "PM"},
		{"a (pm)", "a", "pm"},
		// Timezone
		{"ZZ", "ZZ", "+0000"},
		{"Z", "Z", "UTC"},
		// Fractional: individual segments
		{"cc", "cc", "12"},
		{"SSS", "SSS", "123"},
		{"fff", "fff", "456"},
		{"nnn", "nnn", "789"},
		// Fractional: shorthands
		{"ffffff", "ffffff", "123456"},
		{"nnnnnnnnn", "nnnnnnnnn", "123456789"},
		// Composite
		{"YYYY-MM-DD", "YYYY-MM-DD", "2024-01-15"},
		{"YYYY-MM-DD HH:mm:ss", "YYYY-MM-DD HH:mm:ss", "2024-01-15 12:34:56"},
		{"YYYY/MM/DD", "YYYY/MM/DD", "2024/01/15"},
		{"DD MMM YYYY", "DD MMM YYYY", "15 Jan 2024"},
		{"SSSfff", "SSSfff", "123456"},
		{"SSSfffnnn", "SSSfffnnn", "123456789"},
		// Quiz-style: hide the microsecond segment
		{"SSS'***'nnn", "SSS'***'nnn", "123***789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := testTimestampNano.RenderWithFormat("UTC", *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q) = %q, want %q", tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Escape mechanisms
// ---------------------------------------------------------------------------

func TestRenderWithFormat_Escapes(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{"T as literal separator", "YYYY-MM-DD'T'HH:mm:ss", "2024-01-15T12:34:56"},
		{"literal word", "YYYY-MM-DD 'at' HH:mm", "2024-01-15 at 12:34"},
		{"adjacent escapes", "'foo''bar'", "foobar"},
		{"interleaved", "YYYY'year'MM'month'DD", "2024year01month15"},
		{"token chars inside escape", "'MM-DD'", "MM-DD"},
		{"empty quote is not escaped", "''", "''"},
		{"backslash-backslash", `YYYY\\MM`, `2024\01`},
		{`\Y suppresses year`, `\YYYY`, `Y24Y`}, // \Y→Y, YY→24, Y→pass
		{`\M suppresses month`, `\MM`, `M1`},    // \M→M, M→1
		{`\' is literal quote`, `YYYY\'MM`, `2024'01`},
		{`\' inside quote block`, `'it\'s'`, `it's`},
		{`\n is literal n`, `HH\nmm`, `12n34`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := testTimestampNano.RenderWithFormat("UTC", *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q) = %q, want %q", tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Token interaction / pass-through
// ---------------------------------------------------------------------------

func TestRenderWithFormat_TokenInteraction(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"M:m", "1:34"}, // M=month, m=minute
		{"MM:mm", "01:34"},
		{"mM", "341"}, // m=minute(34), M=month(1)
		{"m:i:I", "34:34:34"},
		{"HH hh", "12 12"},
		{"YYYYmm", "202434"}, // YYYY=2024, mm=34(minute)
		{"DDddd", "15Mon"},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := testTimestampNano.RenderWithFormat("UTC", *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q) = %q, want %q", tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Pass-through characters
// ---------------------------------------------------------------------------

func TestRenderWithFormat_PassThrough(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"Y", "Y"},
		{"d", "d"},
		{"f", "f"},
		{"n", "n"},
		{"T", "T"},
		{"年MM月DD日", "年01月15日"},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := testTimestampNano.RenderWithFormat("UTC", *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q) = %q, want %q", tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Centiseconds
// ---------------------------------------------------------------------------

func TestRenderWithFormat_Centiseconds(t *testing.T) {
	// testTimestampNano = 2024-01-15 12:34:56.123456789 UTC
	// centiseconds = first 2 digits of fractional = "12"
	gdf, err := NewGeneralDateFormat("HH:mm:ss.cc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := testTimestampNano.RenderWithFormat("UTC", *gdf)
	want := "12:34:56.12"
	if got != want {
		t.Errorf("RenderWithFormat(cc) = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// 12-hour clock AM/PM cases
// ---------------------------------------------------------------------------

func TestRenderWithFormat_12Hour(t *testing.T) {
	// testTimestamp = 2024-01-15 12:34:56 UTC (noon = PM, h=12)
	tests := []struct {
		format   string
		expected string
	}{
		{"hh:mm:ss A", "12:34:56 PM"},
		{"hh:mm:ss a", "12:34:56 pm"},
		{"h:mm A", "12:34 PM"},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := testTimestamp.RenderWithFormat("UTC", *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q) = %q, want %q", tt.format, got, tt.expected)
			}
		})
	}
}
