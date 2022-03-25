package main

import (
	"flag"
	"fmt"
	"image/gif"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/jclc/funnygif"
)

var (
	font      *string
	text      *string
	listFonts *bool
	speed     *float64
	position  *int
)

func main() {
	font = flag.String("font", "", "Name of the font to use")
	text = flag.String("text", "", "Text to insert")
	speed = flag.Float64("speed", 1.0, "Speed multiplier (negative values = reverse)")
	position = flag.Int("position", 0, "Caption/speech bubble position")
	listFonts = flag.Bool("list-fonts", false, "List all fonts")
	flag.Parse()

	if *listFonts {
		fmt.Println(funnygif.ListFonts())
		return
	}

	operation := flag.Arg(0)
	inputPath := flag.Arg(1)
	if inputPath == "" {
		log.Fatalln("No input image selected")
	}
	outputPath := flag.Arg(2)
	if outputPath == "" {
		log.Fatalln("No output specified")
	}

	f, err := os.Open(inputPath)
	if err != nil {
		log.Fatalln("Error opening input file:", err)
	}

	g, err := gif.DecodeAll(f)
	f.Close()
	if err != nil {
		log.Fatalln("Error decoding gif:", err)
	}

	var output *gif.GIF
	switch operation {
	case "caption":
		output, err = caption(g)
	case "convert":
		convert(g, outputPath)
		return
	default:
		log.Fatalln("Unknown operation", operation)
	}

	f, err = os.Create(outputPath)
	if err != nil {
		log.Fatalln("Error creating output file:", err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, output)
	if err != nil {
		log.Fatalln("Error saving output:", err)
	}
}

func caption(g *gif.GIF) (*gif.GIF, error) {
	funnygif.LoadDefaultFonts(nil, 1)

	if *font == "" {
		log.Fatalln("No font selected")
	}
	if *text == "" {
		log.Fatalln("No text specified")
	}

	opts := &funnygif.Options{
		Font:            *font,
		Caption:         *text,
		FontSize:        2.0,
		Speed:           *speed,
		CaptionPosition: funnygif.Position(*position),
		CropLeft:        0.1,
		CropBottom:      0.2,
		ScaleWidth:      1.5,
		ScaleHeight:     1.5,
	}

	return funnygif.MakeCaptionGIF(g, opts)
}

func convert(g *gif.GIF, outpath string) {
	outputs := funnygif.GifToRgba(g)
	if err := os.MkdirAll(outpath, 0755); err != nil {
		log.Fatalln(err)
	}
	for i := range outputs {
		f, err := os.Create(filepath.Join(outpath, fmt.Sprintf("%03d.png", i)))
		if err != nil {
			log.Fatalln(err)
		}
		err = png.Encode(f, &outputs[i])
		if err != nil {
			log.Fatalln(err)
		}
	}
}
