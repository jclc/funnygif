package main

import (
	"flag"
	"fmt"
	"image/gif"
	"log"
	"os"

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
	position = flag.Int("position", 0, "Caption position")
	listFonts = flag.Bool("list-fonts", false, "List all fonts")
	flag.Parse()

	funnygif.LoadDefaultFonts(nil, 1)

	if *listFonts {
		fmt.Println(funnygif.ListFonts())
		return
	}

	if *font == "" {
		log.Fatalln("No font selected")
	}
	if *text == "" {
		log.Fatalln("No text specified")
	}
	inputPath := flag.Arg(0)
	if inputPath == "" {
		log.Fatalln("No input image selected")
	}
	outputPath := flag.Arg(1)
	if outputPath == "" {
		log.Fatalln("No output image specified")
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
	// log.Printf("Source image size: {%d, %d}\n", g.Config.Width, g.Config.Height)

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
		// TextColour:       color.RGBA{0xff, 0x00, 0x00, 0xff},
		// BackgroundColour: color.RGBA{0x00, 0x00, 0xff, 0xff},
	}

	out, err := funnygif.Make(g, opts)
	if err != nil {
		log.Fatalln(err)
	}

	f, err = os.Create(outputPath)
	if err != nil {
		log.Fatalln("Error creating output file:", err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, out)
	if err != nil {
		log.Fatalln("Error saving output:", err)
	}
}
