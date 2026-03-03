# GeneralDateFormat — Format String Specification

`GeneralDateFormat` converts a common date format string into a Go `time.Format`-compatible layout string.

---

## Conversion Algorithm

Conversion is performed left-to-right using `strings.NewReplacer`. At each position, the **longest registered token** among those starting at that position (in registration order) is matched and replaced. After a match, the cursor advances past the matched text (non-overlapping). Unrecognized characters are passed through unchanged.

---

## Supported Tokens

### Year

| Token | Output (Go layout) | Example |
|---|---|---|
| `YYYY` | `2006` | `2024` |
| `YY` | `06` | `24` |

### Month

| Token | Output | Example |
|---|---|---|
| `MMMM` | `January` | `January` |
| `MMM` | `Jan` | `Jan` |
| `MM` | `01` | `01` |
| `M` | `1` | `1` |

### Day of Month

| Token | Output | Example |
|---|---|---|
| `DD` | `02` | `05` |
| `D` | `2` | `5` |

### Day of Week

| Token | Output | Example |
|---|---|---|
| `dddd` | `Monday` | `Monday` |
| `ddd` | `Mon` | `Mon` |

> [!NOTE]
> `d` and `dd` are **not** defined tokens and will pass through unchanged.

### Hour

| Token | Output | Clock | Example |
|---|---|---|---|
| `HH` | `15` | 24-hour | `14` |
| `hh` | `03` | 12-hour, zero-padded | `02` |
| `h` | `3` | 12-hour | `2` |

### Minute

All four tokens produce the same Go layout (`04`):

| Token | Output | Example |
|---|---|---|
| `mm` | `04` | `05` |
| `m` | `4` | `5` |
| `ii` | `04` | `05` |
| `II` | `04` | `05` |
| `i` | `4` | `5` |
| `I` | `4` | `5` |

### Second

| Token | Output | Example |
|---|---|---|
| `ss` | `05` | `09` |
| `SS` | `05` | `09` |
| `s` | `5` | `9` |
| `S` | `5` | `9` |

### Fractional Second

> [!IMPORTANT]
> Fractional second tokens must be registered **before** `SS` and `S` in the replacer to take priority.

| Token | Output | Precision | Example |
|---|---|---|---|
| `SSS` | `000` | Milliseconds | `123` |
| `cc` | `00` | Centiseconds (2 digits) | `12` |
| `ffffff` | `000000` | Microseconds | `123456` |
| `nnnnnn` | `000000000` | Nanoseconds | `123456789` |

Shorter variants (`f`, `ff`, ..., `fffff` and `n` through `nnnnn`) are **not** defined and pass through unchanged.

### AM/PM

| Token | Output | Example |
|---|---|---|
| `A` | `PM` | `PM` or `AM` |
| `a` | `pm` | `pm` or `am` |

### Timezone

| Token | Output | Example |
|---|---|---|
| `ZZ` | `-0700` | `+0900`, `-0500` |
| `Z` | `MST` | `JST`, `UTC` |

---

## Literal Escape

There are two mechanisms for including literal text in a format string.

### Single-quote blocks

Wrap text in **single quotes** to prevent token substitution:

```
'<literal text>'
```

**Rules:**
- At least one character must be inside the quotes. An empty pair `''` is not treated as an escape and passes through as-is.
- Any characters (including token letters, digits, separators) inside quotes are preserved literally.
- Multiple escape blocks can appear consecutively or interleaved with tokens.
- To include a literal `'` inside a quoted block, use `\'` (backslash escape, see below).

**Examples:**

| Format | Result (Go layout) |
|---|---|
| `YYYY-MM-DD'T'HH:mm:ss` | `2006-01-02T15:04:05` |
| `YYYY'year'MM'month'DD` | `2006year01month02` |
| `'foo''bar'` | `foobar` |
| `'YYYY-MM-DD'` | `YYYY-MM-DD` (no conversion) |
| `''` | `''` (not an escape — passes through) |
| `'it\'s'` | `it's` (backslash escapes the inner quote) |

> [!TIP]
> `T` is not a defined token, so `YYYY-MM-DDTHH:mm:ss` (without quotes) works correctly. Quotes are recommended for clarity in ISO 8601 formats.

### Backslash escape

Prepend a **backslash** to output any single character literally:

```
\<char>
```

| Sequence | Output | Notes |
|---|---|---|
| `\\` | `\` | Literal backslash |
| `\'` | `'` | Literal single quote (works inside or outside quote blocks) |
| `\Y` | `Y` | Suppresses `Y` from being consumed as part of `YY`/`YYYY` |
| `\n` | `n` | Literal letter n — **not** a newline |
| `\t` | `t` | Literal letter t — **not** a tab |
| `\<any>` | `<any>` | Any other character is passed through as-is |

> [!CAUTION]
> Backslash escapes are interpreted **before** single-quote blocks. A trailing `\` at the end of the format string with no following character is passed through as-is.

---

## Pass-Through Characters

Characters that are not part of any defined token are copied to the output unchanged.

**Examples of characters that are NOT tokens:**

| Character(s) | Note |
|---|---|
| `Y` | Only `YY` and `YYYY` are defined |
| `d`, `dd` | Only `ddd` and `dddd` are defined |
| `f` through `fffff` | Only `ffffff` (6 chars) is defined |
| `n` through `nnnnn` | Only `nnnnnn` (6 chars) is defined |
| `T`, `W`, `X`, etc. | Not defined |
| `0`–`9` | Digits are never tokens |
| `- : / . , [ ]` | Separators pass through unchanged |
| Unicode / multibyte | e.g. `年`, `月`, `日` pass through unchanged |

---

## Token Repetition Behavior

The replacer consumes tokens greedily from left to right. Repeating a token character produces predictable, documented behavior:

| Input | Result | Explanation |
|---|---|---|
| `YYY` | `06Y` | `YY`→`06`, remaining `Y` passes through |
| `YYYYY` | `2006Y` | `YYYY`→`2006`, remaining `Y` passes through |
| `DDD` | `022` | `DD`→`02`, `D`→`2` |
| `DDDD` | `0202` | `DD`→`02`, `DD`→`02` |
| `SSSS` | `0005` | `SSS`→`000`, `S`→`5` |
| `SSSSS` | `00005` | `SSS`→`000`, remaining `SS` has only 2 chars → `S`→`5` |
| `SSSSSS` | `000000` | `SSS`→`000`, `SSS`→`000` |
| `ddddd` | `Mondayd` | `dddd`→`Monday`, remaining `d` passes through |
| `ddddddd` | `MondayMon` | `dddd`→`Monday`, `ddd`→`Mon` |
| `MMMMM` | `January1` | `MMMM`→`January`, `M`→`1` |
| `ZZZ` | `-0700MST` | `ZZ`→`-0700`, `Z`→`MST` |

---

## Common Format Examples

| Format String | Example Output |
|---|---|
| `YYYY-MM-DD'T'HH:mm:ss` | `2024-01-15T12:34:56` |
| `YYYY-MM-DD'T'HH:mm:ssZZ` | `2024-01-15T12:34:56+0000` |
| `MM/DD/YYYY` | `01/15/2024` |
| `M/D/YY` | `1/15/24` |
| `DD.MM.YYYY` | `15.01.2024` |
| `dddd, D MMMM YYYY` | `Monday, 15 January 2024` |
| `hh:mm:ss A` | `12:34:56 PM` |
| `HH:mm:ss.SSS` | `12:34:56.123` |
| `HH:mm:ss.cc` | `12:34:56.12` |
| `HH:mm:ss.ffffff` | `12:34:56.123456` |
| `HH:mm:ss.nnnnnn` | `12:34:56.123456789` |
| `[DD/MMM/YYYY:HH:mm:ss ZZ]` | `[15/Jan/2024:12:34:56 +0000]` |
| `年MM月DD日` | `年01月15日` |
