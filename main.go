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
	"fortio.org/terminal/ansipixels"
	"fortio.org/tray/ray"
	"golang.org/x/image/draw"
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

func Main() int {
	fSample := flag.Float64("s", 2, "Image supersampling factor")
	fRays := flag.Int("r", 32, "Number of rays per pixel")
	fMaxDepth := flag.Int("d", 8, "Maximum ray bounce depth")
	fWorkers := flag.Int("w", 0, "Number of parallel workers (0 = GOMAXPROCS)")
	fCPUProfile := flag.String("profile-cpu", "", "Write CPU profile to file")
	fExit := flag.Bool("exit", false, "Exit immediately after rendering the image once (for timing purposes)")
	fSave := flag.String("save", "", "Save the rendered image to the specified PNG file")
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
	ap := ansipixels.NewAnsiPixels(60)
	if err := ap.Open(); err != nil {
		return 1 // error already logged
	}
	defer ap.Restore()
	ap.SyncBackgroundColor()
	var resized *image.RGBA
	showSplash := !*fExit
	fname := *fSave
	ap.OnResize = func() error {
		ap.StartSyncMode()
		ap.ClearScreen()
		// render at supersampled resolution
		imgWidth, imgHeight := int(math.Round(supersample*float64(ap.W))), int(math.Round(supersample*float64(ap.H*2)))
		rt := ray.New(imgWidth, imgHeight)
		rt.MaxDepth = *fMaxDepth
		rt.NumRaysPerPixel = *fRays
		rt.NumWorkers = *fWorkers
		img := rt.Render(nil) // default scene
		if fname != "" && showSplash {
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
	if *fExit {
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
