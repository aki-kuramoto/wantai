# wantai

> UTC timestamp rendering and general date format conversion for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/aki-kuramoto/wantai.svg)](https://pkg.go.dev/github.com/aki-kuramoto/wantai)
[![CI](https://github.com/aki-kuramoto/wantai/actions/workflows/ci.yml/badge.svg)](https://github.com/aki-kuramoto/wantai/actions/workflows/ci.yml)

---

## Overview

**wantai** is a small Go library that offers a family of typed UTC timestamp wrappers — each backed by a primitive integer type — and a flexible date format engine. Every timestamp type provides the same interface: construct from `time.Time`, render as RFC 3339, and render with a `GeneralDateFormat` pattern.

Choose the type that matches your storage or wire format:

| Type | Underlying | Precision | Range (complete days) |
|---|---|---|---|
| `UtcNanoTs` | `int64` | Nanosecond | 1677-09-22 … 2262-04-10 |
| `UtcMicroTs` | `int64` | Microsecond | ~290,301 BCE … ~294,241 CE |
| `UtcMilliTs` | `int64` | Millisecond | ~292,269,000 BCE … ~292,273,000 CE |
| `UtcSecTsS32` | `int32` | Second | 1901-12-14 … **2038-01-18** |
| `UtcSecTsU32` | `uint32` | Second | 1970-01-01 … 2106-02-06 |
| `UtcSecTsS32Ep2k` | `int32` | Second (Ep2k) | 1931-12-14 … **2068-01-18** |
| `UtcSecTsU32Ep2k` | `uint32` | Second (Ep2k) | 2000-01-01 … 2136-02-06 |

> Ranges are expressed as **complete-day boundaries**: the first and last UTC calendar days
> for which the entire 24-hour period (00:00:00–23:59:59) falls within the type's integer range.
> Partial days at the extremes are excluded.

> **Ep2k** types use **2000-01-01T00:00:00Z** as epoch (instead of the Unix epoch 1970-01-01).  
> Common in embedded / IoT systems that define their own Y2K origin.

---

## Installation

```sh
go get github.com/aki-kuramoto/wantai
```

---

## Usage

### Constructing a timestamp

Each type has a dedicated constructor that accepts `time.Time`:

```go
import (
    "time"
    "github.com/aki-kuramoto/wantai"
)

now := time.Now()

nano  := wantai.FromTime(now)          // UtcNanoTs
micro := wantai.FromTimeMicros(now)    // UtcMicroTs
milli := wantai.FromTimeMillis(now)    // UtcMilliTs
s32   := wantai.FromTimeS32(now)       // UtcSecTsS32
u32   := wantai.FromTimeU32(now)       // UtcSecTsU32
s32k  := wantai.FromTimeS32Ep2k(now)   // UtcSecTsS32Ep2k
u32k  := wantai.FromTimeU32Ep2k(now)   // UtcSecTsU32Ep2k
```

You can also construct directly from the raw integer value:

```go
ts := wantai.UtcMilliTs(1705318496123)  // milliseconds since Unix epoch
```

### Converting back to time.Time

All types implement `ToTime() time.Time`, which always returns a UTC value:

```go
t := wantai.FromTimeMillis(time.Now()).ToTime()  // time.Time in UTC
```

### RFC 3339 rendering

`String()` (implements `fmt.Stringer`) and `Render(timezone)` are available on every type:

```go
ts := wantai.FromTime(time.Now())

fmt.Println(ts)                   // => "2024-01-15T12:34:56Z"    (UTC)
fmt.Println(ts.Render("Asia/Tokyo")) // => "2024-01-15T21:34:56+09:00"
```

An invalid timezone name silently falls back to UTC.

### Rendering with GeneralDateFormat

`RenderWithFormat(timezone, format)` renders using a pre-parsed `GeneralDateFormat` pattern:

```go
gdf, err := wantai.NewGeneralDateFormat("YYYY/MM/DD HH:mm:ss.SSSfffnnn")
if err != nil {
    log.Fatal(err)
}

nano := wantai.FromTime(time.Now())
fmt.Println(nano.RenderWithFormat("UTC", *gdf))
// => "2024/01/15 12:34:56.123456789"

milli := wantai.FromTimeMillis(time.Now())
fmt.Println(milli.RenderWithFormat("Asia/Tokyo", *gdf))
// => "2024/01/15 21:34:56.123000000"  (sub-millisecond digits are zero)
```

---

## GeneralDateFormat tokens

| Token | Meaning | Example |
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
| `ffffff` | Full microseconds (`SSS`+`fff`) | `123456` |
| `nnnnnnnnn` | Full nanoseconds (`SSS`+`fff`+`nnn`) | `123456789` |
| `A` | AM/PM | `PM` |
| `a` | am/pm | `pm` |
| `ZZ` | Numeric timezone offset | `+0900` |
| `Z` | Timezone abbreviation | `JST` |

> Sub-second tokens output `0` for types with lower precision than nanoseconds.  
> For example, `UtcMilliTs` with `HH:mm:ss.ffffff` → `12:34:56.123000`.

The fractional-second tokens `SSS`, `fff`, `nnn` are **composable**:

```
"HH:mm:ss.SSSfffnnn"   => "12:34:56.123456789"   (full nanoseconds)
"HH:mm:ss.SSSfff"      => "12:34:56.123456"       (full microseconds)
"HH:mm:ss.SSS'***'nnn" => "12:34:56.123***789"    (hide the micro part)
```

Wrap text in single quotes to prevent token substitution:

```
"YYYY-MM-DD'T'HH:mm:ss"  =>  "2024-01-15T12:34:56"
"DD MMM YYYY 'at' HH:mm" =>  "15 Jan 2024 at 12:34"
```

For the full specification, see [docs/FORMAT_SPEC.md](docs/FORMAT_SPEC.md).

---

## Timezone handling

`RenderWithFormat` uses different strategies depending on the timezone:

- **Fixed-offset zones** (e.g. `UTC`, `Asia/Tokyo`): all fields are computed via **pure integer arithmetic** — no `time.Time` is created at render time.
- **DST zones** (e.g. `America/New_York`, `Australia/Sydney`): `time.Time` is used only to obtain the correct UTC offset for the given instant; field rendering still uses pure arithmetic.

DST detection is performed once at cache population time via monthly sampling over a 12-month window. Both Northern and Southern Hemisphere DST zones are correctly detected.

## Cache management

Timezone data is cached globally. To reset the cache (e.g. after updating the timezone database):

```go
wantai.ClearLocationCache()
```

---

## License

[MIT](LICENSE.md)
