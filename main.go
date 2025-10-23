package main

import (
	"os"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/terminal/ansipixels"
	"fortio.org/tray/ray"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	cli.Main()
	ap := ansipixels.NewAnsiPixels(60)
	if err := ap.Open(); err != nil {
		return 1 // error already logged
	}
	defer ap.Restore()
	ap.SyncBackgroundColor()
	ap.OnResize = func() error {
		ap.StartSyncMode()
		ap.ClearScreen()
		imgWidth, imgHeight := ap.W, ap.H*2
		rt := ray.New(imgWidth, imgHeight)
		scene := ray.Scene{}
		img := rt.Render(scene)
		_ = ap.ShowScaledImage(img)
		ap.WriteBoxed(ap.H/2-1, "TRay: Terminal Ray-tracing\nImage:%d x %d\nQ to quit.", imgWidth, imgHeight)
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
			log.Infof("Input %q", c)
		}
		return true
	})
	if err != nil {
		log.Infof("Exiting on %v", err)
		return 1
	}
	return 0
}
