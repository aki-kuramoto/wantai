# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

#### New timestamp types
- `UtcMicroTs` (`int64`): microsecond-precision UTC timestamp. Range: 1677-09-21 … 2262-04-11.
- `UtcMilliTs` (`int64`): millisecond-precision UTC timestamp. Range: 1677-09-21 … 2262-04-11.
- `UtcSecTsS32` (`int32`): second-precision UTC timestamp. Range: 1901-12-13 … 2038-01-19 (Year 2038 problem applies).
- `UtcSecTsU32` (`uint32`): second-precision UTC timestamp. Range: 1970-01-01 … 2106-02-07.
- `UtcSecTsS32Ep2k` (`int32`): second-precision UTC timestamp with **Ep2k epoch** (2000-01-01T00:00:00Z). Range: 1931-12-13 … 2068-01-19.
- `UtcSecTsU32Ep2k` (`uint32`): second-precision UTC timestamp with **Ep2k epoch** (2000-01-01T00:00:00Z). Range: 2000-01-01 … 2136-02-07.
- `FromTimeMicros`, `FromTimeMillis`, `FromTimeS32`, `FromTimeU32`, `FromTimeS32Ep2k`, `FromTimeU32Ep2k`: constructors from `time.Time` for each new type.
- `epoch2kUnixSec` internal constant (`946,684,800`): Unix timestamp of the Ep2k epoch origin.

#### Code structure
- `location.go`: shared timezone infrastructure (`locEntry`, `locationCache`, `detectDST`, `getLocEntryWithCache`, `ClearLocationCache`) extracted from `utc_nano_ts.go` into its own file.

#### Format tokens
- `fff` token: microsecond sub-part (digits 4–6 of the nanosecond value).
- `nnn` token: nanosecond sub-part (digits 7–9).
- `SSSfff` / `SSSfffnnn` composite patterns for composing fractional-second segments.
- Quiz-style masking: `SSS'***'nnn` renders milliseconds and nanosecond sub-part while hiding microseconds.

### Changed
- **`GeneralDateFormat`** now stores a pre-parsed `[]formatComponent` instead of a Go layout string.
  The internal engine renders each component directly from the raw integer timestamp value,
  without calling `time.Format`.
- **`RenderWithFormat`** no longer uses `time.Time.Format`. For fixed-offset timezones (no DST)
  all fields are computed via pure integer arithmetic (Gregorian calendar algorithm). For DST
  timezones, `time.Time.Zone()` is called once per render to obtain the correct UTC offset.
- **DST detection** (`detectDST`) now samples 12 points at monthly intervals relative to the
  current time, instead of two fixed points in 1970. This correctly identifies Southern
  Hemisphere DST zones (e.g. `Australia/Sydney`, `Pacific/Auckland`).
- **`locationCache`** now stores `locEntry` (containing `*time.Location`, `hasDST`, `tzName`,
  `offsetSec`) instead of a bare `*time.Location`.
- **`utc_nano_ts.go`** slimmed down to contain only the `UtcNanoTs` type and its methods.
  Shared timezone infrastructure has been moved to `location.go`.

### Removed
- `(gdf *GeneralDateFormat) GoLayout() string` — replaced by the pre-parsed component list.
- `(ts UtcNanoTs) RenderWithGoLayout(timezone, layout string) string` — superseded by `RenderWithFormat`.

---

## [0.1.0] - 2026-03-03

### Added

#### `UtcNanoTs`
- `UtcNanoTs` type: stores a UTC nanosecond Unix timestamp as `int64`.
- `Zero` constant: zero-value `UtcNanoTs(0)` representing the Unix epoch.
- `FromTime(t time.Time) UtcNanoTs`: converts a `time.Time` to `UtcNanoTs`.
- `(ts UtcNanoTs) ToTime() time.Time`: converts `UtcNanoTs` back to UTC `time.Time`.
- `(ts UtcNanoTs) Render(timezone string) string`: renders as RFC3339.
- `(ts UtcNanoTs) RenderWithGoLayout(timezone, layout string) string`: renders with a Go layout string.
- `(ts UtcNanoTs) RenderWithFormat(timezone string, format GeneralDateFormat) string`: renders with a `GeneralDateFormat`.
- `(ts UtcNanoTs) String() string`: implements `fmt.Stringer`; returns RFC3339 in UTC.
- `ClearLocationCache()`: clears the internal timezone location cache.

#### `GeneralDateFormat`
- `GeneralDateFormat` type: converts common format strings (e.g. `YYYY-MM-DD HH:mm:ss`) to Go layout.
- `NewGeneralDateFormat(format string) (*GeneralDateFormat, error)`: constructor.
- `(gdf *GeneralDateFormat) GoLayout() string`: returns the converted Go layout string.
- `(gdf *GeneralDateFormat) String() string`: implements `fmt.Stringer`; returns the original format string.
- Supported tokens: `YYYY`, `YY`, `MMMM`, `MMM`, `MM`, `M`, `DD`, `D`, `dddd`, `ddd`,
  `HH`, `hh`, `h`, `mm`/`ii`/`II`, `m`/`i`/`I`, `ss`/`SS`, `s`/`S`,
  `SSS` (ms), `cc` (cs), `ffffff` (µs), `nnnnnnnnn` (ns), `A`, `a`, `ZZ`, `Z`.
- Single-quote literal escape: `'text'` preserves text without token substitution.
- Backslash escape: `\<char>` outputs any character literally; `\\` outputs a backslash.

[Unreleased]: https://github.com/aki-kuramoto/wantai/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/aki-kuramoto/wantai/releases/tag/v0.1.0
