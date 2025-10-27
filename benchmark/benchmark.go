package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"runtime"
	"runtime/pprof"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/progressbar"
	"fortio.org/tray/ray"
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
	// default matches the book code.
	fRays := flag.Int("r", 10, "Number of rays per pixel")
	fMaxDepth := flag.Int("d", 20, "Maximum ray bounce depth")
	fWorkers := flag.Int("w", 1, "Number of parallel workers (0 = GOMAXPROCS)")
	fCPUProfile := flag.String("profile-cpu", "", "Write CPU profile to file")
	fSave := flag.String("save", "out.png", "Save the rendered image to the specified PNG file")
	// We get 486 objects like the c++ version with seed 7
	fSceneSeed := flag.Uint64("seed", 7, "Seed for the random scene generation (0 randomizes each time)")
	// Matches https://github.com/RayTracing/raytracing.github.io/blob/release/src/InOneWeekend/main.cc#L66-L67
	fWidth := flag.Int("width", 1200, "Image width in pixels")
	fHeight := flag.Int("height", 675, "Image height in pixels")
	cli.Main()
	fname := *fSave
	imgWidth := *fWidth
	imgHeight := *fHeight
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
	var rand ray.Rand
	if *fSceneSeed != 0 {
		rand = ray.NewRand(*fSceneSeed)
	} else {
		rand = ray.NewRandomSource()
	}
	scene := ray.RichScene(rand)
	if *fWorkers <= 0 {
		*fWorkers = runtime.GOMAXPROCS(0)
	}
	log.Infof("Rendering image %dx%d with %d rays/pixel, max depth %d, %d workers, scene seed %d: %d objects",
		imgWidth, imgHeight, *fRays, *fMaxDepth, *fWorkers, *fSceneSeed, len(scene.Objects))
	rt := ray.New(imgWidth, imgHeight)
	rt.MaxDepth = *fMaxDepth
	rt.NumRaysPerPixel = *fRays
	rt.NumWorkers = *fWorkers
	// Camera setup:
	rt.Camera = ray.RichSceneCamera()
	// Setup progress bar
	pb := progressbar.NewBar()
	pb.Prefix = "Rendering "
	total := imgWidth * imgHeight
	p := progressbar.NewAutoProgress(pb, int64(total))
	rt.ProgressFunc = func(n int) {
		p.Update(n)
	}
	img := rt.Render(scene)
	pb.End()
	if fname != "" {
		err := SaveImage(img, fname)
		if err != nil {
			return log.FErrf("could not save image to %q: %v", fname, err)
		}
		log.Infof("Saved rendered image to %q", fname)
	}
	return 0
}
