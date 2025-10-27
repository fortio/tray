[![GoDoc](https://godoc.org/fortio.org/tray?status.svg)](https://pkg.go.dev/fortio.org/tray)
[![Go Report Card](https://goreportcard.com/badge/fortio.org/tray)](https://goreportcard.com/report/fortio.org/tray)
[![GitHub Release](https://img.shields.io/github/release/fortio/tray.svg?style=flat)](https://github.com/fortio/tray/releases/)
[![CI Checks](https://github.com/fortio/tray/actions/workflows/include.yml/badge.svg)](https://github.com/fortio/tray/actions/workflows/include.yml)
[![codecov](https://codecov.io/github/fortio/tray/graph/badge.svg?token=Yx6QaeQr1b)](https://codecov.io/github/fortio/tray)

# TRay

Ray tracing in the terminal. Extending ansipixels.

Inspired by the [Ray Tracing In One Weekend](https://raytracing.github.io/books/RayTracingInOneWeekend.html) book.

Compared to the book:
- This version is in Go (golang)
  - Code is (imo) a lot easier to read
  - ~~With generics to share code between colors and vectors/points yet different types~~ sadly go generics on [3]float64 has a huge negative performance impact. so... not anymore.
- It uses goroutines to render faster (yeah go)
- It can render to any ANSI terminal (truecolor support being better)
- While also saving the full resolution as regular PNG (instead of PPM)
- Lots of (generated, mostly) tests
- You can specify a specific seed for the scene, for reproducible results
- High performance (4x improvements from initial version), almost matches the single threaded C++ (and thus beats it with multiple go routines/cpu cores)
- WIP: navigation in the world

Current demo scene

![Example](example.png)

(Created using `tray -save example.png -r 64 -s 8 -d 50 -seed 2 > /tmp/example.ansi` on a 160x45 terminal.
It does take a while, 3 minutes on 11 core M3 pro)

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

Hit a key to hide the splash info. After which any key causes a re-render, 'Q' to quit.

Save the full resolution image using `-save file.png`.

More options (number of workers, rays per pixel, image super sampling, etc...)
```
tray help

flags:
  -d int
    	Maximum ray bounce depth (default 12)
  -exit
    	Not interactive (no raw), and exit immediately after rendering the image once (for timing purposes)
  -profile-cpu string
    	Write CPU profile to file
  -r int
    	Number of rays per pixel (default 64)
  -s float
    	Image supersampling factor (default 4)
  -save string
    	Save the rendered image to the specified PNG file
  -w int
    	Number of parallel workers (0 = GOMAXPROCS)
```
