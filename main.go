package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"runtime/pprof"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/progressbar"
	"fortio.org/terminal/ansipixels"
	"fortio.org/tray/ray"
	"golang.org/x/image/draw"
	"golang.org/x/term"
)

func main() {
	os.Exit(Main())
}

func SaveImage(img image.Image, fname string) error {
	pngFile, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("could not create PNG file %q: %w", fname, err)
	}
	defer pngFile.Close()
	if err := png.Encode(pngFile, img); err != nil {
		return fmt.Errorf("could not encode PNG to file %q: %w", fname, err)
	}
	return nil
}

// NonRawTerminalSize: gets the terminal size from any of the 3 standard file descriptors (stdout, stderr, stdin).
// TODO: move to ansipixels package as it's quite generally useful.
func NonRawTerminalSize() (width, height int, err error) {
	for _, attempt := range []*os.File{os.Stdout, os.Stderr, os.Stdin} {
		width, height, err = term.GetSize(int(attempt.Fd()))
		if err == nil {
			return width, height, nil
		}
	}
	log.Warnf("Unable to get terminal size from any of stdout, stderr, stdin: %v", err)
	return 80, 24, err
}

func Main() int { //nolint:funlen // yes but fairly linear.
	fSample := flag.Float64("s", 4, "Image supersampling factor")
	fRays := flag.Int("r", 64, "Number of rays per pixel")
	fMaxDepth := flag.Int("d", 12, "Maximum ray bounce depth")
	fWorkers := flag.Int("w", 0, "Number of parallel workers (0 = GOMAXPROCS)")
	fCPUProfile := flag.String("profile-cpu", "", "Write CPU profile to file")
	fExit := flag.Bool("exit", false,
		"Not interactive (no raw), and exit immediately after rendering the image once (for timing purposes)")
	fSave := flag.String("save", "", "Save the rendered image to the specified PNG file")
	fSceneSeed := flag.Uint64("seed", 0, "Seed for the random scene generation (0 randomizes each time)")
	cli.Main()
	if *fCPUProfile != "" {
		f, err := os.Create(*fCPUProfile)
		if err != nil {
			return log.FErrf("Could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return log.FErrf("Could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	supersample := *fSample
	if supersample <= 0 {
		supersample = 1
	}
	var ap *ansipixels.AnsiPixels
	exitAfterRender := *fExit
	normalRawMode := !exitAfterRender
	if normalRawMode && !term.IsTerminal(int(os.Stdout.Fd())) {
		log.Warnf("Stdout is not a terminal, switching to non-raw mode")
		normalRawMode = false
		exitAfterRender = true
	}
	if normalRawMode {
		ap = ansipixels.NewAnsiPixels(60)
		if err := ap.Open(); err != nil {
			return 1 // error already logged
		}
		defer ap.Restore()
		ap.SyncBackgroundColor()
	} else {
		ap = ansipixels.NewAnsiPixels(0) // 0 fps == for blocking non raw mode
		ap.W, ap.H, _ = NonRawTerminalSize()
		defer fmt.Println()
	}
	var resized *image.RGBA
	showSplash := normalRawMode
	fname := *fSave
	var rand ray.Rand
	if *fSceneSeed != 0 {
		rand = ray.NewRand(*fSceneSeed)
	} else {
		rand = ray.NewRandomSource()
	}
	scene := ray.RichScene(rand)
	ap.OnResize = func() error {
		ap.ClearScreen()
		// render at supersampled resolution
		imgWidth, imgHeight := int(math.Round(supersample*float64(ap.W))), int(math.Round(supersample*float64(ap.H*2)))
		rt := ray.New(imgWidth, imgHeight)
		rt.MaxDepth = *fMaxDepth
		rt.NumRaysPerPixel = *fRays
		rt.NumWorkers = *fWorkers
		// Camera setup:
		rt.Camera = ray.RichSceneCamera()
		// Setup progress bar
		pb := progressbar.NewBar()
		pb.Prefix = "Rendering "
		pb.ScreenWriter = ap.Logger
		total := imgWidth * imgHeight
		p := progressbar.NewAutoProgress(pb, int64(total))
		rt.ProgressFunc = func(n int) {
			p.Update(n)
		}
		img := rt.Render(scene)
		pb.End()
		if fname != "" && (showSplash || exitAfterRender) {
			// only save once, not after keypresses
			err := SaveImage(img, fname)
			if err != nil {
				return fmt.Errorf("could not save image to %q: %w", fname, err)
			}
			log.Infof("Saved rendered image to %q", fname)
		}
		// Downscale image:
		resized = img
		if supersample != 1 {
			origBounds := img.Bounds()
			resized = image.NewRGBA(image.Rect(0, 0, ap.W, ap.H*2))
			if supersample < 1 {
				draw.NearestNeighbor.Scale(resized, resized.Bounds(), img, origBounds, draw.Over, nil)
			} else {
				draw.BiLinear.Scale(resized, resized.Bounds(), img, origBounds, draw.Over, nil)
			}
		}
		_ = ap.ShowScaledImage(resized)
		if showSplash {
			ap.WriteBoxed(ap.H/2-2, "TRay: Terminal Ray-tracing\n%d x %d image (%.1fx)\nRays %d, Depth %d\nQ to quit.",
				imgWidth, imgHeight, supersample, rt.NumRaysPerPixel, rt.MaxDepth)
		}
		ap.EndSyncMode()
		return nil
	}
	_ = ap.OnResize() // initial draw.
	if exitAfterRender {
		return 0
	}
	ap.AutoSync = false
	err := ap.FPSTicks(func() bool {
		if len(ap.Data) == 0 {
			return true
		}
		c := ap.Data[0]
		switch c {
		case 'q', 'Q', 3: // Ctrl-C
			log.Infof("Exiting on %q", c)
			return false
		default:
			log.Debugf("Input %q, rerendering...", c)
			if showSplash {
				ap.HideCursor()
				showSplash = false
				_ = ap.ShowScaledImage(resized)
			} else {
				_ = ap.OnResize()
			}
		}
		return true
	})
	if err != nil {
		log.Infof("Exiting on %v", err)
		return 1
	}
	return 0
}
