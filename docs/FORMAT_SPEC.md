# GeneralDateFormat — Format String Specification

`GeneralDateFormat` pre-parses a common date format string into an internal component list.
`RenderWithFormat` walks the list and expands each component directly from the timestamp
value — **without** calling Go's `time.Format`.

---

## Parsing Algorithm

The format string is parsed left-to-right in a single pass after two preprocessing steps:

1. **Backslash escapes** — `\<char>` is replaced with a NUL-delimited placeholder.
2. **Single-quote blocks** — `'text'` is replaced with a NUL-delimited placeholder (inner `\'` is  first captured in step 1).
3. **Token matching** — at each position the **longest registered token** is matched greedily. Unrecognized characters become literal (`fcFixed`) components.

---

## Supported Tokens

### Year

| Token | Example |
|---|---|
| `YYYY` | `2024` |
| `YY` | `24` |

### Month

| Token | Example |
|---|---|
| `MMMM` | `January` |
| `MMM` | `Jan` |
| `MM` | `01` |
| `M` | `1` |

### Day of Month

| Token | Example |
|---|---|
| `DD` | `15` |
| `D` | `15` |

### Day of Week

| Token | Example |
|---|---|
| `dddd` | `Monday` |
| `ddd` | `Mon` |

> [!NOTE]
> `d` and `dd` are **not** defined tokens and will pass through unchanged.

### Hour

| Token | Clock | Example |
|---|---|---|
| `HH` | 24-hour | `14` |
| `hh` | 12-hour, zero-padded | `02` |
| `h` | 12-hour | `2` |

### Minute

All six tokens produce the same output:

| Token | Example |
|---|---|
| `mm` / `ii` / `II` | `05` |
| `m` / `i` / `I` | `5` |

### Second

| Token | Example |
|---|---|
| `ss` / `SS` | `09` |
| `s` / `S` | `9` |

### Fractional Second

Fractional-second tokens are **composable segments** of the nanosecond value.
Each token represents a fixed slice of digits within the 9-digit fractional part:

| Token | Digits | Precision | Example (`.123456789`) |
|---|---|---|---|
| `cc` | 1–2 | Centiseconds | `12` |
| `SSS` | 1–3 | Milliseconds | `123` |
| `fff` | 4–6 | Microsecond sub-part | `456` |
| `nnn` | 7–9 | Nanosecond sub-part | `789` |
| `ffffff` | 1–6 | Full microseconds (`SSS`+`fff`) | `123456` |
| `nnnnnnnnn` | 1–9 | Full nanoseconds (`SSS`+`fff`+`nnn`) | `123456789` |

> [!IMPORTANT]
> `fff` and `nnn` are **segment tokens**, not standalone fractional-second tokens.
> Used alone they output 3 digits from the corresponding position of the nanosecond value.
> Their primary purpose is composition with `SSS`:
>
> ```
> "SSS'***'nnn"   =>  "123***789"   (hide the microsecond segment)
> "SSSfff"        =>  "123456"      (same as ffffff)
> "SSSfffnnn"     =>  "123456789"   (same as nnnnnnnnn)
> ```

Shorter variants (`f`, `ff`, …, `fffff` and `n` through `nnnnn`, `nnnnnn` through `nnnnnnnn`) are **not** defined and pass through unchanged.

### AM/PM

| Token | Example |
|---|---|
| `A` | `PM` / `AM` |
| `a` | `pm` / `am` |

### Timezone

| Token | Example |
|---|---|
| `ZZ` | `+0900`, `-0500` |
| `Z` | `JST`, `UTC`, `EST` |

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

| Format | Example output |
|---|---|
| `YYYY-MM-DD'T'HH:mm:ss` | `2024-01-15T12:34:56` |
| `YYYY'year'MM'month'DD` | `2024year01month15` |
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
> Backslash escapes are processed **before** single-quote blocks. A trailing `\` at the end of the format string with no following character is passed through as-is.

---

## Pass-Through Characters

Characters that are not part of any defined token are copied to the output unchanged.

**Examples of characters that are NOT tokens:**

| Character(s) | Note |
|---|---|
| `Y` | Only `YY` and `YYYY` are defined |
| `d`, `dd` | Only `ddd` and `dddd` are defined |
| `f` through `fffff` | Only `fff` (3) and `ffffff` (6) are defined |
| `n` through `nn`, `nnnn` through `nnnnnnnn` | Only `nnn` (3) and `nnnnnnnnn` (9) are defined |
| `T`, `W`, `X`, etc. | Not defined |
| `0`–`9` | Digits are never tokens |
| `- : / . , [ ]` | Separators pass through unchanged |
| Unicode / multibyte | e.g. `年`, `月`, `日` pass through unchanged |

---

## Token Repetition Behavior

The tokenizer consumes greedily from left to right. Repeating a token character produces predictable, documented behavior:

| Input | Example output | Explanation |
|---|---|---|
| `YYY` | `24Y` | `YY`→year2, remaining `Y` passes through |
| `YYYYY` | `2024Y` | `YYYY`→year4, remaining `Y` passes through |
| `DDD` | `152` | `DD`→`15`, `D`→`15` |
| `SSSS` | `1235` | `SSS`→ms(3), `S`→sec |
| `SSSSSS` | `123123` | `SSS`→ms, `SSS`→ms |
| `ddddd` | `Mondayd` | `dddd`→weekday, remaining `d` passes through |
| `ddddddd` | `MondayMon` | `dddd`→weekday, `ddd`→abbr |
| `MMMMM` | `January1` | `MMMM`→full month, `M`→month num |
| `ZZZ` | `+0900JST` | `ZZ`→offset, `Z`→abbr |
| `ffff` | `456f` | `fff`→micro sub, `f` passes through |
| `nnnn` | `789n` | `nnn`→nano sub, `n` passes through |

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
| `HH:mm:ss.nnnnnnnnn` | `12:34:56.123456789` |
| `HH:mm:ss.SSSfff` | `12:34:56.123456` |
| `HH:mm:ss.SSSfffnnn` | `12:34:56.123456789` |
| `HH:mm:ss.SSS'***'nnn` | `12:34:56.123***789` |
| `[DD/MMM/YYYY:HH:mm:ss ZZ]` | `[15/Jan/2024:12:34:56 +0000]` |
| `年MM月DD日` | `年01月15日` |
