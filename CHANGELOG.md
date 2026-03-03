# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [0.1.0] - 2026-03-03

### Added

#### `UtcNanoTs`
- `UtcNanoTs` type: stores a UTC nanosecond Unix timestamp as `int64`.
- `ZERO` constant: zero-value `UtcNanoTs(0)` representing the Unix epoch.
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
  `SSS` (ms), `cc` (cs), `ffffff` (µs), `nnnnnn` (ns), `A`, `a`, `ZZ`, `Z`.
- Single-quote literal escape: `'text'` preserves text without token substitution.
- Backslash escape: `\<char>` outputs any character literally; `\\` outputs a backslash.

[Unreleased]: https://github.com/aki-kuramoto/wantai/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/aki-kuramoto/wantai/releases/tag/v0.1.0
