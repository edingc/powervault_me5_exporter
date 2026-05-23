# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.1] - 2026-05-23

- Packaging fixes.

## [2.0.0] - 2026-05-23

### Added

- Concurrent sub-collector dispatch via `--me5.scrape-concurrency` (default `6`) —
  sub-collectors run in parallel within a semaphore-bounded goroutine pool, reducing
  scrape latency on large arrays.
- Configurable scrape timeout via `--me5.scrape-timeout` (default `2m`), replacing
  the previously hardcoded 1-minute deadline.
- `build-static` Makefile target (`CGO_ENABLED=0`) for portable binaries without a
  glibc dependency — improves compatibility across Linux distributions.
- `enclosure_id` label on `me5_sensor_status` and `me5_sensor_value` — supports
  deployments with primary and expansion enclosures (JBOD shelves). Single-enclosure
  deployments will see `enclosure_id="0"` on all sensor metrics.

### Fixed

- `--no-collect.<name>` flags now correctly disable sub-collectors — previously
  `disabled=false` was never written to the enabled map, so the flag had no effect.

### Changed

- **BREAKING:** `me5_sensor_status` and `me5_sensor_value` gain a new `enclosure_id`
  label. Existing PromQL queries and Grafana panels targeting these metrics must be
  updated. Single-enclosure deployments will see `enclosure_id="0"`.

## [1.0.4] - 2026-05-23

- Dependency updates.

## [1.0.3] - 2026-03-19

- Deployment fixes.

## [1.0.2] - 2026-03-19

- GoReleaser fixes.

## [1.0.1] - 2026-03-19

- Various workflow and README fixes.

## [1.0.0] - 2026-03-19

- Initial release.