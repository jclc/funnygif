package funnygif

import (
	"fmt"
	"image"
	"image/gif"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/image/font/opentype"
)

const (
	dpi = 72.0
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

type Options struct {
	Speed                                    float64
	CropTop, CropBottom, CropLeft, CropRight float64
	ScaleWidth, ScaleHeight                  float64
	Start, End                               float64
	Caption                                  string
	CaptionBottom                            bool
	Font                                     string
}

func Make(g *gif.GIF, opts *Options) (*image.RGBA, error) {
	return image.NewRGBA(image.Rect(0, 0, 100, 100)), nil
}
