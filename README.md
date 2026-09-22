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

`ToTimeIn(timezone)` returns the same instant read in another zone:

```go
t, err := ts.ToTimeIn("Asia/Tokyo")  // time.Time with a +09:00 location
```

Use it rather than `ts.ToTime().In(loc)` with your own `time.LoadLocation`: the
standard library does not cache `LoadLocation`, so that costs roughly 12µs on
**every** call, while `ToTimeIn` goes through the same zone cache as the
renderers and settles at a few tens of nanoseconds.

Unlike `Render`, an unresolvable name is reported instead of falling back to
UTC, and the returned `time.Time` is the zero value — ignoring the error gives
you something obviously broken rather than a plausible instant in the wrong
zone.

### RFC 3339 rendering

`String()` (implements `fmt.Stringer`) and `Render(timezone)` are available on every type:

```go
ts := wantai.FromTime(time.Now())

fmt.Println(ts)                   // => "2024-01-15T12:34:56Z"    (UTC)
fmt.Println(ts.Render("Asia/Tokyo")) // => "2024-01-15T21:34:56+09:00"
```

See [Timezone names](#timezone-names) for what `Render` accepts — a name that
cannot be resolved renders as UTC, with nothing in the result to say so.

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

### Showing an instant in someone's local time

Store UTC; work out the local reading when you display it. Two ways, and the
difference is who the zone belongs to:

```go
// The machine this process runs on. "Local" is Go's own name for that zone,
// so no TZ environment variable and no lookup of the system's name is needed.
fmt.Println(ts.Render("Local"))        // => "2024-01-15T21:34:56+09:00"

// A zone you chose — from a config file, a user profile, a database row.
fmt.Println(ts.Render("Asia/Tokyo"))   // => "2024-01-15T21:34:56+09:00"
```

Reach for the explicit name when the zone belongs to the **data** (a person's
profile, a scheduled job, a record of where something happened), and for
`"Local"` when it belongs to the **machine** (a log read on the box that wrote
it).

---

## Timezone names

The `timezone` argument of `Render` and `RenderWithFormat` goes to
[`time.LoadLocation`](https://pkg.go.dev/time#LoadLocation), so it is an IANA
name from the tz database, plus the three names Go defines itself:

| Argument | Zone |
|---|---|
| `"Asia/Tokyo"` | that zone |
| `""` | UTC |
| `"UTC"` | UTC |
| `"Local"` | the zone this process is running in |

### A name is not an abbreviation and not an offset

This is the sharp edge. The two things a person is most likely to write are the
two that do not work:

```go
ts.Render("JST")      // => "2024-01-15T12:34:56Z"   UTC — nine hours off
ts.Render("+09:00")   // => "2024-01-15T12:34:56Z"   the same
```

An unresolvable name falls back to UTC, and what you get back is a well-formed
timestamp in the wrong zone. `Render` returns only a string, so the result
cannot tell you: `Render("UTC")` and `Render("Nonsense/Zone")` are byte for byte
identical.

Spelling is part of the name, too. `"asia/tokyo"` resolves on macOS, whose
filesystem does not distinguish case, and falls back to UTC on Linux — so a
lowercase name can pass every test on a laptop and render UTC in production.

### Naming the machine's own zone

`"Local"` renders in the machine's zone without needing its name. When the name
itself is wanted — to store next to an instant, to show which zone a reading
used, or simply because a concrete name is preferred — `SystemZoneName` reports
it:

```go
name, err := wantai.SystemZoneName()   // => "Asia/Tokyo"
```

There is no standard-library equivalent: `time.Local` knows its offset but not
its name, and `time.Local.String()` answers `"Local"` unless `TZ` is set. So it
reads what the platform recorded — `TZ` or `/etc/localtime` on Unix, the
registry's `TimeZoneKeyName` translated through CLDR on Windows — and the
answer agrees with `time.Local` on each.

The cost is uneven: a few hundred nanoseconds when `TZ` is set, microseconds for
a symlinked `/etc/localtime`, and **milliseconds** on a system where that file
is a copy rather than a symlink (common in container images), because the whole
zoneinfo tree then has to be read and compared. Which of those a machine is
cannot be told from the code, so the answer is cached. Drop it with
`ClearSystemZoneNameCache()` if the machine's zone may have changed — that is
separate from `ClearLocationCache()` on purpose, because the two go stale for
different reasons.

An error means the zone could not be *named*, not that there is none. A caller
that only needs to render can fall back to `"Local"`.

### Check a name before it gets to a renderer

Where the name comes from configuration, a user, or anywhere else it could be
mistyped, resolve it once at startup and fail loudly:

```go
if err := wantai.TryToLoadZoneForErrorCheck(cfg.Timezone); err != nil {
    return fmt.Errorf("timezone %q: %w", cfg.Timezone, err)
}
```

`TryToLoadZoneForErrorCheck` reports what the renderers cannot, and leaves the zone in the cache
so the first render does not pay for the lookup.

### The timezone database on Windows

Windows has no system timezone database. Go reads the current zone from system
calls there and has no file source for `time.LoadLocation`, so an IANA name
resolves only from an embedded copy. Without one every name falls back to UTC,
and `SystemZoneName` would hand you a name that `Render` then refuses.

So on Windows this package embeds the database (the standard library's
`time/tzdata`). It costs about **400KB** of binary size, and nothing on any
other platform — the file carrying it is compiled only for Windows.

A program that renders only in `"UTC"` and `"Local"` needs none of it, since
`time.LoadLocation` answers both without touching the database. Leave it out
with:

```sh
go build -tags wantai_no_tzdata_on_windows
```

The tag does nothing on any other platform. An unknown build tag is silently
ignored by Go, so a misspelling leaves the data in — a binary larger than
intended rather than one that reports the wrong time.

Building with the standard library's own `-tags timetzdata` embeds the same
data regardless of this tag.

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

Timezone data is cached globally, keyed by the name it was asked for. To reset
the cache:

```go
wantai.ClearLocationCache()
```

Reset it after updating the timezone database, and whenever the zone behind a
name changes under a running process. `"Local"` is cached like any other name,
so a machine that changes zone — and a test that swaps `time.Local` — keeps
rendering in the old zone until the cache is dropped.

`SystemZoneName` has its own cache and its own `ClearSystemZoneNameCache()`.
They are separate because they go stale for different reasons: a resolved zone
goes stale when the tz database is updated, a system zone name when the machine
is set to a different zone.

---

## License

[MIT](LICENSE.md)
