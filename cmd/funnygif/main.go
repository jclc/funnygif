package main

import (
	"flag"
	"image/gif"
	"image/jpeg"
	"log"
	"os"

	"github.com/jclc/funnygif"
)

var (
	fontPath *string
	font     *string
	image    *string
)

func main() {
	fontPath = flag.String("font-path", "/usr/share/fonts", "Path to the font files")
	font = flag.String("font", "", "Name of the font to use")
	// image = flag.String("image", "", "Path to the image to edit")
	flag.Parse()

	funnygif.LoadFonts([]string{*fontPath}, 1)
	if *font == "" {
		log.Fatalln("No font selected")
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

	opts := &funnygif.Options{}

	out, err := funnygif.Make(g, opts)
	if err != nil {
		log.Fatalln(err)
	}

	f, err = os.Create(outputPath)
	if err != nil {
		log.Fatalln("Error creating output file:", err)
	}
	defer f.Close()

	err = jpeg.Encode(f, out, nil)
	if err != nil {
		log.Fatalln("Error saving output:", err)
	}
}
