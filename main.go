package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/big"
	"os"
	"sync"

	"github.com/ALTree/bigfloat"

	"gopkg.in/cheggaaa/pb.v1"
)

func iterate(x, y, bailout *big.Float, maxIter int, prec uint) (int, *big.Float) {
	// TODO: use object pool.
	x0 := big.NewFloat(0.0).SetPrec(prec)
	y0 := big.NewFloat(0.0).SetPrec(prec)

	x0sq := new(big.Float).SetPrec(prec)
	y0sq := new(big.Float).SetPrec(prec)
	x0y0 := new(big.Float).SetPrec(prec)
	norm := new(big.Float).SetPrec(prec)

	for i := 0; ; i++ {
		x0sq.Mul(x0, x0)
		y0sq.Mul(y0, y0)
		norm.Add(x0sq, y0sq)
		if norm.Cmp(bailout) == 1 || i >= maxIter {
			return i, norm
		}
		x0y0.Mul(x0, y0)

		x0.Sub(x0sq, y0sq)
		x0.Add(x0, x)

		y0.Add(x0y0, x0y0)
		y0.Add(y0, y)
	}
}

func main() {
	w := flag.Int("w", 1280, "width of the viewport")
	h := flag.Int("h", 720, "height of the viewport")
	centerx := flag.String("centerx", "0.0", "x coordinate of the center of the zoom area")
	centery := flag.String("centery", "0.0", "y coordinate of the center of the zoom area")
	bailout := flag.Float64("bailout", 64.0, "bailout")
	scale := flag.Float64("scale", 256.0, "palette scale")
	shift := flag.Float64("shift", 0.0, "palette shift")
	zoom := flag.String("zoom", "0.256", "zoom level")
	prec := flag.Uint("prec", 128, "computation precision")
	iters := flag.Int("iters", 1024, "maximum number of iterations per pixel")
	out := flag.String("out", "out.png", "file where output is stored")
	flag.Parse()

	// Naming things is hard :(

	cx, ok := new(big.Float).SetPrec(*prec).SetString(*centerx)
	if !ok {
		fmt.Println("could not parse centerx arg")
		os.Exit(1)
	}

	cy, ok := new(big.Float).SetPrec(*prec).SetString(*centery)
	if !ok {
		fmt.Println("could not parse centery arg")
		os.Exit(1)
	}

	z, ok := new(big.Float).SetPrec(*prec).SetString(*zoom)
	if !ok {
		fmt.Println("could not parse zoom arg")
		os.Exit(1)
	}

	// fw = w / (zoom*h)
	fw := big.NewFloat(float64(*w)).SetPrec(*prec)
	fw.Quo(fw, big.NewFloat(float64(*h)).SetPrec(*prec))
	fw.Quo(fw, z)

	// fh = 1 / zoom
	fh := big.NewFloat(1.0).SetPrec(*prec)
	fh.Quo(fh, z)

	// dx = fw / w
	dx := new(big.Float).Copy(fw)
	dx.Quo(dx, big.NewFloat(float64(*w)).SetPrec(*prec))

	// dy = fh / h
	dy := new(big.Float).Copy(fh)
	dy.Quo(dy, big.NewFloat(float64(*h)).SetPrec(*prec))

	// sx = cx - fw/2
	sx := new(big.Float).Copy(fw)
	sx.Mul(sx, big.NewFloat(0.5).SetPrec(*prec))
	sx.Sub(cx, sx)

	// sy = cy - fh/2
	sy := new(big.Float).Copy(fh)
	sy.Mul(sy, big.NewFloat(0.5).SetPrec(*prec))
	sy.Sub(cy, sy)

	bar := pb.StartNew((*w) * (*h))
	img := image.NewRGBA(image.Rect(0, 0, *w, *h))

	// Cache some constants so we don't have to compute them each time.
	b := big.NewFloat(*bailout).SetPrec(*prec)
	logBailout := bigfloat.Log(bigfloat.Log(b))
	log2 := bigfloat.Log(big.NewFloat(2).SetPrec(*prec))

	var wg sync.WaitGroup
	for y := 0; y < *h; y++ {
		wg.Add(1)
		go func(y int) {
			// fx = sx
			fx := new(big.Float).Copy(sx)

			// fy = sy + y * dy
			fy := big.NewFloat(float64(y)).SetPrec(*prec)
			fy.Mul(fy, dy)
			fy.Add(sy, fy)

			for x := 0; x < *w; x++ {
				it, norm := iterate(fx, fy, b, *iters, *prec)

				var c color.Color
				if it < *iters {
					// smoothed = it + 1 - (lnln(|z|^2) - lnln(bailout)) / ln2
					smoothed := bigfloat.Log(bigfloat.Log(norm))
					smoothed.Sub(smoothed, logBailout)
					smoothed.Quo(smoothed, log2)
					smoothed.Sub(big.NewFloat(float64(it+1)).SetPrec(*prec), smoothed)

					idx, _ := smoothed.Float64()
					idx = math.Sqrt(idx)*(*scale) + (*shift)
					c = palette[int(idx)%len(palette)]
				} else {
					c = backgroundColor
				}

				img.Set(x, *h-y-1, c)
				bar.Increment()
				fx.Add(fx, dx)
			}
			wg.Done()
		}(y)
	}
	wg.Wait()

	f, err := os.Create(*out)
	if err != nil {
		fmt.Printf("could not create %s: %v\n", *out, err)
		os.Exit(1)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		fmt.Printf("could not encode image: %v\n", err)
		os.Exit(1)
	}
}
