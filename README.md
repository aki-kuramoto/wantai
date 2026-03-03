# wantai

> UTC nanosecond timestamp rendering and general date format conversion for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/aki-kuramoto/wantai.svg)](https://pkg.go.dev/github.com/aki-kuramoto/wantai)
[![CI](https://github.com/aki-kuramoto/wantai/actions/workflows/ci.yml/badge.svg)](https://github.com/aki-kuramoto/wantai/actions/workflows/ci.yml)

---

## Overview

**wantai** is a small Go library that provides:

- **`UtcNanoTs`** — A type that wraps a UTC nanosecond Unix timestamp (`int64`) and renders it as a human-readable string in any timezone.
- **`GeneralDateFormat`** — Pre-parses common date format strings (e.g. `YYYY-MM-DD HH:mm:ss`) into an efficient component list, then renders directly without using Go's `time.Format`.

---

## Installation

```sh
go get github.com/aki-kuramoto/wantai
```

---

## Usage

### UtcNanoTs

```go
package main

import (
    "fmt"
    "time"
    "github.com/aki-kuramoto/wantai"
)

func main() {
    // Create from the current time
    ts := wantai.FromTime(time.Now())

    // RFC3339 in a given timezone (falls back to UTC on invalid timezone)
    fmt.Println(ts.Render("Asia/Tokyo"))
    // => "2024-01-15T21:34:56+09:00"

    // fmt.Stringer — same as Render("UTC")
    fmt.Println(ts)
    // => "2024-01-15T12:34:56Z"

    // Render with a GeneralDateFormat
    gdf, _ := wantai.NewGeneralDateFormat("YYYY/MM/DD HH:mm:ss.SSSfffnnn")
    fmt.Println(ts.RenderWithFormat("UTC", *gdf))
    // => "2024/01/15 12:34:56.123456789"

    // Convert back to time.Time (always UTC)
    t := ts.ToTime()
    _ = t
}
```

---

### GeneralDateFormat

```go
gdf, err := wantai.NewGeneralDateFormat("YYYY-MM-DD'T'HH:mm:ss.SSSfffnnn")
if err != nil {
    log.Fatal(err)
}

fmt.Println(gdf.String()) // => "YYYY-MM-DD'T'HH:mm:ss.SSSfffnnn"

ts := wantai.FromTime(time.Now())
fmt.Println(ts.RenderWithFormat("Asia/Tokyo", *gdf))
// => "2024-01-15T21:34:56.123456789"
```

#### Supported Tokens

| Token | Meaning | Example output |
|---|---|---|
| `YYYY` | 4-digit year | `2024` |
| `YY` | 2-digit year | `24` |
| `MMMM` | Full month name | `January` |
| `MMM` | Short month name | `Jan` |
| `MM` | Zero-padded month | `01` |
| `M` | Month | `1` |
| `DD` | Zero-padded day | `15` |
| `D` | Day | `15` |
| `dddd` | Full weekday | `Monday` |
| `ddd` | Short weekday | `Mon` |
| `HH` | 24-hour | `14` |
| `hh` | 12-hour, zero-padded | `02` |
| `h` | 12-hour | `2` |
| `mm` / `ii` / `II` | Zero-padded minute | `05` |
| `m` / `i` / `I` | Minute | `5` |
| `ss` / `SS` | Zero-padded second | `09` |
| `s` / `S` | Second | `9` |
| `cc` | Centiseconds — digits 1–2 | `12` |
| `SSS` | Milliseconds — digits 1–3 | `123` |
| `fff` | Microsecond sub-part — digits 4–6 | `456` |
| `nnn` | Nanosecond sub-part — digits 7–9 | `789` |
| `ffffff` | Full microseconds (= `SSS`+`fff`) | `123456` |
| `nnnnnnnnn` | Full nanoseconds (= `SSS`+`fff`+`nnn`) | `123456789` |
| `A` | AM/PM | `PM` |
| `a` | am/pm | `pm` |
| `ZZ` | Numeric timezone offset | `+0900` |
| `Z` | Timezone abbreviation | `JST` |

The fractional-second tokens `SSS`, `fff`, `nnn` are **composable**:

```
"HH:mm:ss.SSSfffnnn"  => "12:34:56.123456789"   (full nanoseconds)
"HH:mm:ss.SSSfff"     => "12:34:56.123456"       (full microseconds)
"HH:mm:ss.SSS'***'nnn"=> "12:34:56.123***789"    (hide the micro part)
```

#### Literal escape — single quotes

Wrap text in single quotes to prevent token substitution:

```
"YYYY-MM-DD'T'HH:mm:ss"   =>  "2024-01-15T12:34:56"
"DD MMM YYYY 'at' HH:mm"  =>  "15 Jan 2024 at 12:34"
```

#### Literal escape — backslash

Prepend `\` to output any character literally:

```
"YYYY\\MM"    =>  "2024\01"    (\\ = literal backslash)
"YYYY\'MM"    =>  "2024'01"    (\' = literal single quote)
"'it\'s'"     =>  "it's"       (\' inside a quote block)
```

For full format specification and edge-case behavior, see [docs/FORMAT_SPEC.md](docs/FORMAT_SPEC.md).

---

## Timezone handling

`RenderWithFormat` uses different strategies depending on the timezone:

- **Fixed-offset zones** (e.g. `UTC`, `Asia/Tokyo`): all fields are computed via **pure integer arithmetic** — no `time.Time` is created at render time.
- **DST zones** (e.g. `America/New_York`, `Australia/Sydney`): `time.Time` is used only to obtain the correct UTC offset for the given instant; field rendering still uses pure arithmetic.

DST detection is performed once at cache population time via monthly sampling over a 12-month window.

## Cache management

`UtcNanoTs` caches loaded timezone data for performance. To reset the cache (e.g. after updating timezone data):

```go
wantai.ClearLocationCache()
```

---

## License

[MIT](LICENSE.md)
