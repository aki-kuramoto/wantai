# wantai

> UTC nanosecond timestamp rendering and general date format conversion for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/aki-kuramoto/wantai.svg)](https://pkg.go.dev/github.com/aki-kuramoto/wantai)
[![CI](https://github.com/aki-kuramoto/wantai/actions/workflows/ci.yml/badge.svg)](https://github.com/aki-kuramoto/wantai/actions/workflows/ci.yml)

---

## Overview

**wantai** is a small Go library that provides:

- **`UtcNanoTs`** — A type that wraps a UTC nanosecond Unix timestamp (`int64`) and renders it as a human-readable string in any timezone.
- **`GeneralDateFormat`** — Converts common date format strings (e.g. `YYYY-MM-DD HH:mm:ss`) to Go's reference-time layout, with support for literal escapes and backslash escaping.

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

    // Render with a Go layout string
    fmt.Println(ts.RenderWithGoLayout("America/New_York", "2006-01-02 15:04:05"))
    // => "2024-01-15 07:34:56"

    // Render with a GeneralDateFormat
    gdf, _ := wantai.NewGeneralDateFormat("YYYY/MM/DD HH:mm:ss")
    fmt.Println(ts.RenderWithFormat("UTC", *gdf))
    // => "2024/01/15 12:34:56"

    // Convert back to time.Time (always UTC)
    t := ts.ToTime()
    _ = t
}
```

---

### GeneralDateFormat

```go
gdf, err := wantai.NewGeneralDateFormat("YYYY-MM-DD'T'HH:mm:ss.SSS")
if err != nil {
    log.Fatal(err)
}

fmt.Println(gdf.GoLayout()) // => "2006-01-02T15:04:05.000"
fmt.Println(gdf.String())   // => "YYYY-MM-DD'T'HH:mm:ss.SSS"
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
| `DD` | Zero-padded day | `05` |
| `D` | Day | `5` |
| `dddd` | Full weekday | `Monday` |
| `ddd` | Short weekday | `Mon` |
| `HH` | 24-hour | `14` |
| `hh` | 12-hour, zero-padded | `02` |
| `h` | 12-hour | `2` |
| `mm` / `ii` / `II` | Zero-padded minute | `05` |
| `m` / `i` / `I` | Minute | `5` |
| `ss` / `SS` | Zero-padded second | `09` |
| `s` / `S` | Second | `9` |
| `SSS` | Milliseconds (3 digits) | `123` |
| `cc` | Centiseconds (2 digits) | `12` |
| `ffffff` | Microseconds (6 digits) | `123456` |
| `nnnnnn` | Nanoseconds (9 digits) | `123456789` |
| `A` | AM/PM | `PM` |
| `a` | am/pm | `pm` |
| `ZZ` | Numeric timezone offset | `+0900` |
| `Z` | Timezone abbreviation | `JST` |

#### Literal escape — single quotes

Wrap text in single quotes to prevent token substitution:

```
"YYYY-MM-DD'T'HH:mm:ss"   =>  "2006-01-02T15:04:05"
"DD MMM YYYY 'at' HH:mm"  =>  "02 Jan 2006 at 15:04"
```

#### Literal escape — backslash

Prepend `\` to output any character literally:

```
"YYYY\\MM"    =>  "2006\01"    (\\ = literal backslash)
"YYYY\'MM"    =>  "2006'01"    (\' = literal single quote)
"'it\'s'"     =>  "it's"       (\' inside a quote block)
```

For full format specification and edge-case behavior, see [docs/FORMAT_SPEC.md](docs/FORMAT_SPEC.md).

---

## Cache management

`UtcNanoTs` caches loaded `*time.Location` values for performance. To reset the cache (e.g. after updating timezone data):

```go
wantai.ClearLocationCache()
```

---

## License

[MIT](LICENSE.md)
