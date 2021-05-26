package funnygif

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	dpi              = 72.0
	textToWidthRatio = 0.1
	paddingRatio     = 0.035
	lineSpacing      = 0.0
	maxScaling       = 3 // maximum scaling per dimension; max size is this^2
)

type Position int

const (
	Above Position = iota
	Top
	Middle
	Bottom // Bottom Text
	Below
)

type loadedFont struct {
	name string
	font *opentype.Font
}

var fontFiles map[string]string // maps font names to .ttf filepaths
var defaultFont string
var maxLoadedFonts int
var loadedFonts []loadedFont
var extensionRegexp = regexp.MustCompile(`$\.[ot]tf^`)

// var loadedFonts []struct{Name string, }

func LoadFonts(searchPaths []string, maxLoaded int) error {
	maxLoadedFonts = maxLoaded

	if fontFiles == nil {
		fontFiles = make(map[string]string)
	}

	for _, p := range searchPaths {
		err := filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() || extensionRegexp.MatchString(filepath.Ext(info.Name())) {
				return nil
			}
			fontFiles[info.Name()[:len(info.Name())-4]] = path
			return nil
		})

		if err != nil {
			return fmt.Errorf("error loading fonts: %w", err)
		}
	}
	return nil
}

func LoadDefaultFonts(additionalPaths []string, maxLoaded int) error {
	return LoadFonts(append(getFontPaths(), additionalPaths...), maxLoaded)
}

func loadFont(font string) (*opentype.Font, error) {
	if font == "" {
		font = defaultFont
	}

	p, found := fontFiles[font]
	if !found {
		return nil, fmt.Errorf("font '%s' not found", font)
	}

	for _, f := range loadedFonts {
		if f.name == font {
			return f.font, nil
		}
	}

	if len(loadedFonts) >= maxLoadedFonts {
		loadedFonts = loadedFonts[1:]
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("error reading font file: %w", err)
	}
	f, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing font: %w", err)
	}
	loadedFonts = append(loadedFonts, loadedFont{
		name: font,
		font: f,
	})
	return f, nil
}

func SetDefaultFont(font string) error {
	if _, found := fontFiles[font]; !found {
		return fmt.Errorf("no font '%s' found", font)
	}
	defaultFont = font
	return nil
}

func ListFonts() []string {
	fonts := make([]string, 0, len(fontFiles))
	for k := range fontFiles {
		fonts = append(fonts, k)
	}
	return fonts
}

// Options is a struct containing the options to use when making the edited gif.
type Options struct {
	// Speed is a speed multiplier for the final gif. Negative values reverse
	// the animation.
	Speed float64
	// CropTop, CropBottom, CropLeft and CropRight are values ranging from
	// 0.0 to 1.0, indicating how much should be cropped from the sides.
	CropTop, CropBottom, CropLeft, CropRight float64
	// ScaleWidth and ScaleHeight are scaling multipliers.
	ScaleWidth, ScaleHeight float64
	// Start and End are values ranging from 0.0 to 1.0 indicating how much
	// of the start and end of the gif should be cut off.
	Start, End float64
	// Caption is the (un)funny text you'll want to add to the gif.
	Caption string
	// Position of the caption
	CaptionPosition Position
	// Name of the font to use.
	Font string
	// Relative size of the font.
	FontSize         float64
	TextColour       color.RGBA
	BackgroundColour color.RGBA
}

func Make(g *gif.GIF, opts *Options) (*gif.GIF, error) {
	if opts.Speed < 0.001 && opts.Speed > -0.001 {
		opts.Speed = 1.0
	}
	if opts.CropTop+opts.CropBottom >= 0.95 {
		return nil, errors.New("too much vertical crop")
	}
	if opts.CropLeft+opts.CropRight >= 0.95 {
		return nil, errors.New("too much horizontal crop")
	}
	if opts.Start+opts.End > 1.0 {
		return nil, errors.New("start and end can't overlap")
	}

	if opts.TextColour.A == 0 {
		opts.TextColour = color.RGBA{0xff, 0xff, 0xff, 0xff}
	}
	if opts.CaptionPosition > Above && opts.CaptionPosition < Below {
		opts.BackgroundColour = color.RGBA{0x00, 0x00, 0x00, 0x00}
	} else if opts.BackgroundColour.A == 0 {
		opts.BackgroundColour = color.RGBA{0x00, 0x00, 0x00, 0xff}
	}

	if opts.ScaleWidth < 0.1 {
		opts.ScaleWidth = 1.0
	}
	if opts.ScaleHeight < 0.1 {
		opts.ScaleHeight = 1.0
	}

	if opts.ScaleWidth*opts.ScaleHeight > maxScaling*maxScaling {
		scaleBack := math.Sqrt((maxScaling * maxScaling) / (opts.ScaleHeight * opts.ScaleHeight))
		opts.ScaleWidth /= scaleBack
		opts.ScaleHeight /= scaleBack
	}

	srcRect := image.Rect(
		int(float64(g.Config.Width)*(opts.CropLeft)),
		int(float64(g.Config.Height)*(opts.CropTop)),
		int(float64(g.Config.Width)*(1.0-opts.CropRight)),
		int(float64(g.Config.Height)*(1.0-opts.CropBottom)),
	)
	dstRect := image.Rect(0, 0,
		int(float64(srcRect.Dx()*int(opts.ScaleWidth))),
		int(float64(srcRect.Dy()*int(opts.ScaleHeight))),
	)

	fontSize := textToWidthRatio * float64(dstRect.Dx())

	fnt, err := loadFont(opts.Font)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating font face: %w", err)
	}

	padding := int(((float64(dstRect.Dx()) + float64(dstRect.Dy())) / 2) * paddingRatio)
	lines, widths := fitText(opts.Caption, dstRect.Dx()-(2*padding), face)

	textHeight := 2*padding +
		len(lines)*int(fontSize) +
		int(float64((len(lines)-1))*lineSpacing*float64(fontSize))

	if textHeight > 2*dstRect.Dy() {
		return nil, errors.New("text spans too many lines, try a smaller font size")
	}

	textImg := image.NewRGBA(image.Rect(0, 0, dstRect.Dx(), textHeight))

	// Fill background
	draw.Draw(textImg, textImg.Bounds(),
		&image.Uniform{C: opts.BackgroundColour}, image.Point{}, draw.Src)

	d := font.Drawer{
		Dst:  textImg,
		Src:  &image.Uniform{C: opts.TextColour},
		Face: face,
	}
	for i := range lines {
		d.Dot = fixed.P(
			(dstRect.Dx()-widths[i])/2,
			padding+(i+1)*int(fontSize)+i*int(fontSize*lineSpacing),
		)
		d.DrawString(lines[i])
	}

	return process(dstRect, srcRect, g, textImg, opts), nil
}

func fitText(text string, width int, face font.Face) (lines []string, widths []int) {
	type word struct {
		s   string
		len int
	}
	words := make([]word, 0)
	var w word
	var l fixed.Int26_6
	for _, r := range text {
		if unicode.IsSpace(r) {
			if w.s != "" {
				w.len = l.Round()
				words = append(words, w)
				w = word{}
				l = 0
			}
			continue
		}
		l1, _ := face.GlyphAdvance(r)
		l += l1
		w.s += string(r)
	}
	if w.s != "" {
		w.len = l.Round()
		words = append(words, w)
	}

	wsF, _ := face.GlyphAdvance(' ')
	ws := wsF.Round()

	var line strings.Builder
	var lineLen int
	for _, w := range words {
		if lineLen+ws+w.len > width && lineLen != 0 {
			// move to the next line
			lines = append(lines, line.String())
			line.Reset()
			widths = append(widths, lineLen)
			lineLen = 0
		}
		if lineLen != 0 {
			line.WriteByte(' ')
			lineLen += ws
		}
		line.WriteString(w.s)
		lineLen += w.len
	}

	if lineLen != 0 {
		lines = append(lines, line.String())
		widths = append(widths, lineLen)
	}
	return
}

func process(dstRect, srcRect image.Rectangle, src *gif.GIF, text *image.RGBA, opts *Options) *gif.GIF {
	reverse := opts.Speed < 0
	if reverse {
		opts.Speed *= -1
	}

	imageBounds := dstRect
	if opts.CaptionPosition == Above || opts.CaptionPosition == Below {
		imageBounds = imageBounds.Union(dstRect.Add(image.Point{0, text.Rect.Dy()}))
	}

	if opts.CaptionPosition == Above {
		dstRect = dstRect.Add(image.Point{0, text.Rect.Dy()})
	}

	var textOffset image.Point
	switch opts.CaptionPosition {
	case Middle:
		textOffset.Y = (imageBounds.Dy() - text.Rect.Dy()) / 2
	case Bottom, Below:
		textOffset.Y = imageBounds.Dy() - text.Rect.Dy()
	}

	// var d draw.Drawer
	output := &gif.GIF{}
	output.Config = image.Config{
		Width:      imageBounds.Dx(),
		Height:     imageBounds.Dy(),
		ColorModel: src.Config.ColorModel,
	}

	for i := range src.Image {
		var srcI int
		if reverse {
			srcI = len(src.Image) - i - 1
		} else {
			srcI = i
		}

		img := image.NewPaletted(imageBounds, src.Image[srcI].Palette)

		xdraw.ApproxBiLinear.Scale(img, dstRect, src.Image[srcI], srcRect, xdraw.Src, nil)
		draw.Draw(img, text.Bounds().Add(textOffset), text, image.Point{}, draw.Over)
		output.Image = append(output.Image, img)
		output.Delay = append(output.Delay, int(float64(src.Delay[srcI])/opts.Speed))
	}

	return output
}
