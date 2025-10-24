[![GoDoc](https://godoc.org/fortio.org/tray?status.svg)](https://pkg.go.dev/fortio.org/tray)
[![Go Report Card](https://goreportcard.com/badge/fortio.org/tray)](https://goreportcard.com/report/fortio.org/tray)
[![GitHub Release](https://img.shields.io/github/release/fortio/tray.svg?style=flat)](https://github.com/fortio/tray/releases/)
[![CI Checks](https://github.com/fortio/tray/actions/workflows/include.yml/badge.svg)](https://github.com/fortio/tray/actions/workflows/include.yml)
[![codecov](https://codecov.io/github/fortio/tray/graph/badge.svg?token=Yx6QaeQr1b)](https://codecov.io/github/fortio/tray)

# TRay

Ray tracing in the terminal. Extending ansipixels.

Inspired by the [Ray Tracing In One Weekend](https://raytracing.github.io/books/RayTracingInOneWeekend.html) book.

## Install
You can get the binary from [releases](https://github.com/fortio/tray/releases)

Or just run
```
CGO_ENABLED=0 go install fortio.org/tray@latest  # to install (in ~/go/bin typically) or just
CGO_ENABLED=0 go run fortio.org/tray@latest  # to run without install
```

or
```
brew install fortio/tap/tray
```

or even - but multicast that we need doesn't seem to work at least on docker for mac.
```
docker run --network host -v ~/.tray:/home/user/.tray -ti fortio/tray
```


## Usage

Hit a key to hide the splash info. After which any key cause a re-render, 'Q' to quit.

```
tray help
```
