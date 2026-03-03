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

func TestGoLayout_BasicTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Year
		{"YYYY", "YYYY", "2006"},
		{"YY", "YY", "06"},
		// Month
		{"MMMM", "MMMM", "January"},
		{"MMM", "MMM", "Jan"},
		{"MM", "MM", "01"},
		{"M", "M", "1"},
		// Day
		{"DD", "DD", "02"},
		{"D", "D", "2"},
		// Day of week
		{"dddd", "dddd", "Monday"},
		{"ddd", "ddd", "Mon"},
		// Hour
		{"HH (24h)", "HH", "15"},
		{"hh (12h)", "hh", "03"},
		{"h (12h)", "h", "3"},
		// Minute
		{"mm", "mm", "04"},
		{"ii", "ii", "04"},
		{"II", "II", "04"},
		// Second
		{"ss", "ss", "05"},
		{"SS", "SS", "05"},
		// Fractional second
		{"SSS (ms)", "SSS", "000"},
		{"ffffff (us)", "ffffff", "000000"},
		{"nnnnnn (ns)", "nnnnnn", "000000000"},
		// AM/PM
		{"A", "A", "PM"},
		{"a", "a", "pm"},
		// Timezone
		{"ZZ", "ZZ", "-0700"},
		{"Z", "Z", "MST"},
		// Composite
		{"YYYY-MM-DD", "YYYY-MM-DD", "2006-01-02"},
		{"YYYY-MM-DD HH:mm:ss", "YYYY-MM-DD HH:mm:ss", "2006-01-02 15:04:05"},
		{"YYYY/MM/DD", "YYYY/MM/DD", "2006/01/02"},
		{"DD MMM YYYY", "DD MMM YYYY", "02 Jan 2006"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := gdf.GoLayout(); got != tt.expected {
				t.Errorf("GoLayout() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGoLayout_EscapedBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "T as literal separator",
			input:    "YYYY-MM-DD'T'HH:mm:ss",
			expected: "2006-01-02T15:04:05",
		},
		{
			name:     "at as literal word",
			input:    "YYYY-MM-DD 'at' HH:mm",
			expected: "2006-01-02 at 15:04",
		},
		{
			name:     "multiple escaped blocks",
			input:    "'Year:' YYYY, 'Month:' MM",
			expected: "Year: 2006, Month: 01",
		},
		{
			name:     "escape only",
			input:    "'hello world'",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := gdf.GoLayout(); got != tt.expected {
				t.Errorf("GoLayout() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGoLayout_EmptyFormat(t *testing.T) {
	gdf, err := NewGeneralDateFormat("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := gdf.GoLayout(); got != "" {
		t.Errorf("GoLayout() = %q, want empty string", got)
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
// Category 1: Token repetition / boundary tests
// Verifies how strings.NewReplacer consumes tokens when the same character
// is repeated, left-to-right, non-overlapping.
// ---------------------------------------------------------------------------

func TestGoLayout_TokenRepetition(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Year
		{"YYY", "06Y"},     // YY→06, Y→pass-through
		{"YYYYY", "2006Y"}, // YYYY→2006, Y→pass-through
		// Day
		{"DDD", "022"},   // DD→02, D→2
		{"DDDD", "0202"}, // DD→02, DD→02
		// Minute (aliases)
		{"mmm", "044"}, // mm→04, m→4
		{"iii", "044"}, // ii→04, i→4
		{"III", "044"}, // II→04, I→4
		// Fractional second (SSS registered before SS/S)
		{"SSSS", "0005"},     // SSS→000, S→5
		{"SSSSS", "00005"},   // SSS→000, S→5 (only 2 chars remain, ≠ SS)
		{"SSSSSS", "000000"}, // SSS→000, SSS→000
		// Day-of-week
		{"ddddd", "Mondayd"},     // dddd→Monday, d→pass-through
		{"dddddd", "Mondaydd"},   // dddd→Monday, dd→pass-through
		{"ddddddd", "MondayMon"}, // dddd→Monday, ddd→Mon
		// Month
		{"MMMMM", "January1"},   // MMMM→January, M→1
		{"MMMMMM", "January01"}, // MMMM→January, MM→01
		// Hour
		{"hhh", "033"}, // hh→03, h→3
		// Second
		{"sss", "055"}, // ss→05, s→5
		// Timezone
		{"ZZZ", "-0700MST"}, // ZZ→-0700, Z→MST
		// Fractional-only tokens need exact length
		{"ffff", "ffff"},          // 4 chars: no token defined → pass-through
		{"fffff", "fffff"},        // 5 chars: pass-through
		{"fffffff", "000000f"},    // ffffff→000000, f→pass-through
		{"nnnnn", "nnnnn"},        // 5 chars: pass-through
		{"nnnnnnn", "000000000n"}, // nnnnnn→000000000, n→pass-through
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := gdf.GoLayout(); got != tt.expected {
				t.Errorf("GoLayout(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Category 2: Token interaction / mixed types
// Verifies correct conversion when different token types appear adjacent.
// ---------------------------------------------------------------------------

func TestGoLayout_TokenInteraction(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Month (M) vs Minute (m)
		{"M:m", "1:4"},
		{"MM:mm", "01:04"},
		{"mM", "41"}, // reversed: m(minute)→4, M(month)→1
		// Minute aliases mixed
		{"m:i:I", "4:4:4"},
		// Second + fractional
		{"ss.SSS", "05.000"},
		{"ss.ffffff", "05.000000"},
		{"ss.nnnnnn", "05.000000000"},
		// AM/PM combinations
		{"hh A", "03 PM"},
		{"HH A", "15 PM"}, // syntactically odd but converts correctly
		{"h a", "3 pm"},
		// Minute alias cross
		{"Ii", "44"}, // I→4, i→4
		// Both hour types
		{"HH hh", "15 03"},
		// Year + minute (different cases)
		{"YYYYmm", "200604"},
		// Day + day-of-week
		{"DDddd", "02Mon"}, // DD→02, ddd→Mon
		// Complex interleaving
		{"mM:Mm", "41:14"}, // m→4, M→1, :, M→1, m→4
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := gdf.GoLayout(); got != tt.expected {
				t.Errorf("GoLayout(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Category 3: Escape mechanism edge cases
// Regex: '([^']+)' requires at least one character inside quotes.
// ---------------------------------------------------------------------------

func TestGoLayout_EscapeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"escape at start", "'prefix'YYYY", "prefix2006"},
		{"escape at end", "YYYY'suffix'", "2006suffix"},
		{"token chars inside escape", "'MM-DD'", "MM-DD"},
		{"all token chars escaped", "'YYYY-MM-DD HH:mm:ss'", "YYYY-MM-DD HH:mm:ss"},
		{"adjacent escapes", "'foo''bar'", "foobar"},
		{"interleaved tokens and escapes", "YYYY'year'MM'month'DD", "2006year01month02"},
		{"escape with separators", "'a/b:c'", "a/b:c"},
		{"empty quote is not escaped", "''", "''"}, // regex requires 1+ chars inside
		{"T without quote passes through", "YYYY-MM-DDTHH:mm:ss", "2006-01-02T15:04:05"},
		{"digits inside escape", "'2024'", "2024"},
		{"multiple T escapes interleaved", "YYYY'T'MM'T'DD", "2006T01T02"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := gdf.GoLayout(); got != tt.expected {
				t.Errorf("GoLayout(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Category 5: Pass-through / non-token characters
// Characters that look like tokens but are not registered.
// ---------------------------------------------------------------------------

func TestGoLayout_PassThrough(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Undefined single-char variants of longer tokens
		{"Y", "Y"},         // only YY / YYYY are defined
		{"d", "d"},         // only ddd / dddd are defined
		{"f", "f"},         // only ffffff is defined
		{"fffff", "fffff"}, // 5 chars: pass-through
		{"n", "n"},         // only nnnnnn is defined
		{"nnnnn", "nnnnn"}, // 5 chars: pass-through
		// Other uppercase letters
		{"T", "T"},
		{"WXQKLNOP", "WXQKLNOP"},
		// Digits
		{"0123456789", "0123456789"},
		// Symbols
		{"!@#$%^&*", "!@#$%^&*"},
		// Multibyte (Japanese)
		{"年月日", "年月日"},
		{"年MM月DD日", "年01月02日"}, // MM and DD convert; kanji pass-through
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := gdf.GoLayout(); got != tt.expected {
				t.Errorf("GoLayout(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Centisecond token: cc
// ---------------------------------------------------------------------------

func TestGoLayout_Centiseconds(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"cc", "00"}, // cc → centiseconds (Go layout "00")
		{"HH:mm:ss.cc", "15:04:05.00"},
		{"ss.cc", "05.00"},
		// cc next to other fractional tokens
		{"cc SSS", "00 000"},
		// cc repeated: cc+cc = 0000
		{"cccc", "0000"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := gdf.GoLayout(); got != tt.expected {
				t.Errorf("GoLayout(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRender_Centiseconds(t *testing.T) {
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
// Backslash escape
// ---------------------------------------------------------------------------

func TestGoLayout_BackslashEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// \\ → literal backslash
		{`\\ is literal backslash`, `YYYY\\MM`, `2006\01`},
		// \<token> → literal token character (no conversion)
		{`\Y suppresses year token`, `\YYYY`, `Y06Y`}, // \Y→Y, YY→06, Y→pass-through
		{`\M suppresses month token`, `\MM`, `M1`},    // \M→M, M→1
		// \' → literal single quote (outside quote block)
		{`\' is literal single quote`, `YYYY\'MM`, `2006'01`},
		// \' inside a quoted block
		{`\' inside quote block`, `'it\'s'`, `it's`},
		// \ at end of input — trailing backslash: the backslash itself has no next char
		// (this is an edge case: no character to consume — backslash passes through)
		// separator characters
		{`\n is literal n (not newline)`, `HH\nmm`, `15n04`},
		{`\t is literal t (not tab)`, `HH\tmm`, `15t04`},
		// backslash before non-token character
		{`\! is literal !`, `YYYY\!MM`, `2006!01`},
		// consecutive backslash escapes
		{`consecutive \\`, `YYYY\\\\MM`, `2006\\01`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := gdf.GoLayout(); got != tt.expected {
				t.Errorf("GoLayout(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
