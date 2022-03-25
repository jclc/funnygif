package funnygif

import (
	"image"
	"image/color"
	"image/gif"
	"sync"

	_ "github.com/ericpauley/go-quantize/quantize"
	"golang.org/x/image/draw"
)

func GifToRgba(g *gif.GIF) []*image.RGBA {
	job := func(area image.Rectangle, tpindex uint8, in *image.Paletted, background, out *image.RGBA, wg *sync.WaitGroup) {
		for x := area.Min.X; x < area.Max.X; x++ {
			for y := area.Min.Y; y < area.Max.Y; y++ {
				ind := in.ColorIndexAt(x, y)
				if ind == tpindex {
					// pixel is transparent; copy value from background
					out.SetRGBA(x, y, background.RGBAAt(x, y))
				} else {
					// pixel is opaque; copy pixel value from in to out and background
					col := in.At(x, y)
					out.Set(x, y, col)
					background.Set(x, y, col)
				}
			}
		}
		if wg != nil {
			wg.Done()
		}
	}

	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	background := image.NewRGBA(bounds)
	outputs := make([]*image.RGBA, 0, len(g.Image))

	for i := range g.Image {

		// Find the transparent index if it exists
		tpindex := g.BackgroundIndex
		// if g.Disposal[i] == gif.DisposalPrevious {
		// 	tpindex = int(g.BackgroundIndex)
		// }
		// for i, col := range g.Image[i].Palette {
		// 	if _, _, _, a := col.RGBA(); a == 0 {
		// 		tpindex = i
		// 	}
		// }
		// fmt.Printf("tpindex@%d: %d\n", i, tpindex)
		// fmt.Printf("color@%d: %v\n", i, g.Image[i].At(0, 0))
		// fmt.Printf("index@%d: %v\n", i, g.Image[i].ColorIndexAt(0, 0))
		// fmt.Printf("disposal@%d, %v\n", i, g.Disposal[i])

		outputs = append(outputs, image.NewRGBA(bounds))
		inBounds := g.Image[i].Bounds()
		if inBounds != bounds {
			// copy parts of the background that are outside the current frame
			if topHalf := image.Rect(
				bounds.Min.X, bounds.Min.Y,
				bounds.Max.X, inBounds.Min.Y); !topHalf.Empty() {
				draw.Draw(outputs[i], topHalf, background, bounds.Min, draw.Src)
			}
			if leftSide := image.Rect(
				bounds.Min.X, inBounds.Min.Y,
				inBounds.Min.X, inBounds.Max.Y); !leftSide.Empty() {
				draw.Draw(outputs[i], leftSide, background, image.Pt(bounds.Min.X, inBounds.Min.Y), draw.Src)
			}
			if rightSide := image.Rect(
				inBounds.Max.X, inBounds.Min.Y,
				bounds.Max.X, inBounds.Max.Y); !rightSide.Empty() {
				draw.Draw(outputs[i], rightSide, background, image.Pt(inBounds.Max.X, inBounds.Min.Y), draw.Src)
			}
			if bottomHalf := image.Rect(
				bounds.Min.X, inBounds.Max.Y,
				bounds.Max.X, bounds.Max.Y); !bottomHalf.Empty() {
				draw.Draw(outputs[i], bottomHalf, background, image.Pt(bounds.Min.X, inBounds.Max.Y), draw.Src)
			}
			// draw.Copy(&outputs[i], image.Pt(0, 0), g.Image[i], g.Image[i].Bounds(), draw.Src, nil)
			// continue
			// draw.Draw(&outputs[i], bounds, background, image.Pt(0, 0), draw.Src)
		}

		// Draw the new frame on the new image and the background
		if inBounds.Dx() >= 8 && inBounds.Dy() >= 8 {
			// Split to jobs
			areas := []image.Rectangle{
				{
					Min: inBounds.Min,
					Max: image.Pt(inBounds.Min.X+inBounds.Dx()/2, inBounds.Min.Y+inBounds.Dy()/2),
				},
				{
					Min: image.Pt(inBounds.Min.X, inBounds.Min.Y+inBounds.Dy()/2),
					Max: image.Pt(inBounds.Min.X+inBounds.Dx()/2, inBounds.Min.Y+inBounds.Dy()),
				},
				{
					Min: image.Pt(inBounds.Min.X+inBounds.Dx()/2, inBounds.Min.Y),
					Max: image.Pt(inBounds.Min.X+inBounds.Dx(), inBounds.Min.Y+inBounds.Dy()/2),
				},
				{
					Min: image.Pt(inBounds.Min.X+inBounds.Dx()/2, inBounds.Min.Y+inBounds.Dy()/2),
					Max: inBounds.Max,
				},
			}

			var wg sync.WaitGroup
			wg.Add(len(areas))
			for _, area := range areas {
				go job(area, tpindex, g.Image[i], background, outputs[i], &wg)
			}
			wg.Wait()
		} else {
			job(inBounds, tpindex, g.Image[i], background, outputs[i], nil)
		}
	}

	return outputs
}

func RgbaToGif(imgs []*image.RGBA, delays []int) *gif.GIF {
	bounds := imgs[0].Rect
	// generate delta masks
	deltas := make([]*image.Alpha, len(imgs)-1)
	var wg sync.WaitGroup
	wg.Add(len(imgs) - 1)
	for i := 1; i < len(imgs); i++ {
		go func(i int) {
			delta := image.NewAlpha(bounds)
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					a := imgs[i-1].RGBAAt(x, y)
					b := imgs[i].RGBAAt(x, y)
					if a.R-b.R != 0 ||
						a.G-b.G != 0 ||
						a.B-b.B != 0 ||
						a.A-b.A != 0 {
						delta.SetAlpha(x, y, color.Alpha{255})
					}
				}
			}
			deltas[i-1] = delta
			wg.Done()
		}(i)
	}
	wg.Wait()
	return nil
}
