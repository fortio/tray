package main

import (
	"flag"
	"image"
	"math"
	"os"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/terminal/ansipixels"
	"fortio.org/tray/ray"
	"golang.org/x/image/draw"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	fSample := flag.Float64("s", 2, "Supersampling factor")
	cli.Main()
	supersample := *fSample
	if supersample <= 0 {
		supersample = 1
	}
	log.Infof("Starting TRay with supersampling x%f", supersample)
	ap := ansipixels.NewAnsiPixels(60)
	if err := ap.Open(); err != nil {
		return 1 // error already logged
	}
	defer ap.Restore()
	ap.SyncBackgroundColor()
	var resized *image.RGBA
	ap.OnResize = func() error {
		ap.StartSyncMode()
		ap.ClearScreen()
		imgWidth, imgHeight := int(math.Round(supersample*float64(ap.W))), int(math.Round(supersample*float64(ap.H*2)))
		rt := ray.New(imgWidth, imgHeight) // supersample x2
		img := rt.Render(nil)              // default scene
		// Downscale image:
		resized = img
		if supersample != 1 {
			origBounds := img.Bounds()
			resized = image.NewRGBA(image.Rect(0, 0, ap.W, ap.H*2))
			if supersample < 1 {
				draw.NearestNeighbor.Scale(resized, resized.Bounds(), img, origBounds, draw.Over, nil)
			} else {
				draw.CatmullRom.Scale(resized, resized.Bounds(), img, origBounds, draw.Over, nil)
			}
		}
		_ = ap.ShowScaledImage(resized)
		ap.WriteBoxed(ap.H/2-1, "TRay: Terminal Ray-tracing\nImage:%d x %d (Sample x%.1f)\nQ to quit.",
			imgWidth, imgHeight, supersample)
		ap.EndSyncMode()
		return nil
	}
	_ = ap.OnResize() // initial draw.
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
			log.Debugf("Input %q", c)
			ap.HideCursor()
			_ = ap.ShowScaledImage(resized)
		}
		return true
	})
	if err != nil {
		log.Infof("Exiting on %v", err)
		return 1
	}
	return 0
}
