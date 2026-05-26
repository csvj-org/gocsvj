# Changelog

All notable changes to this project are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
aims to use [Semantic Versioning](https://semver.org/spec/v2.0.0.html) once
a first tagged release is cut.

## [Unreleased]

### Added
- GitHub Actions CI workflow (`.github/workflows/ci.yml`) running `go vet`
  and `go test` on a matrix of Go 1.23 and stable.
- Dependabot configuration (`.github/dependabot.yml`) covering both Go
  modules and GitHub Actions.
- `go.mod` declaring the module path `github.com/csvj-org/gocsvj` and
  Go 1.23 as the minimum.

### Changed
- Third-party GitHub Actions in the CI workflow are now SHA-pinned with
  the tag preserved as a trailing comment.

### Removed
- `.travis.yml` — Travis CI is no longer used.

## 2018-10-29 — Initial reference implementation

Unreleased / untagged. The history before 2026 is preserved here for
reference; the public module path and tagged versions begin with the
first entry under [Unreleased].

### Added
- `reader.go` and `reader_test.go` — streaming reader for CSVJ files.
- `writer.go` and `writer_test.go` — writer producing spec-compliant
  output.
- `lint/` package and `csvj-cmd/` command-line linter.
- MIT LICENSE and initial README.
