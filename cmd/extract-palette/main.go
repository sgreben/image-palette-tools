package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"text/template"

	"github.com/sgreben/flagvar/template"
	"github.com/sgreben/image-palette-tools/pkg/palette"
)

var (
	k            int
	outPng       = flagvar.Template{Root: templateSettings}
	outTxt       = flagvar.Template{Root: templateSettings}
	outJSON      = flagvar.Template{Root: templateSettings}
	outColorSize int
	maxParallel  int

	templateSettings = template.New("").Funcs(map[string]interface{}{
		"abs":      func(s string) (string, error) { return filepath.Abs(s) },
		"basename": func(s string) string { return filepath.Base(s) },
		"dirname":  func(s string) string { return filepath.Dir(s) },
		"ext":      func(s string) string { return filepath.Ext(s) },
		"html":     func(c color.RGBA) string { return html(c) },
	})
	cache       = palette.NewColorCache(512)
	printBuffer = 1024
)

func init() {
	log.SetOutput(os.Stderr)
	flag.IntVar(&k, "k", 8, "number of colors to extract")
	flag.IntVar(&maxParallel, "p", runtime.GOMAXPROCS(0), "number of images to process in parallel")
	flag.Var(&outPng, "out-png", "path of output palette image (PNG) (go template)")
	flag.IntVar(&outColorSize, "out-png-height", 100, "size of each color square in the palette output image")
	flag.Var(&outTxt, "out-txt", "path of output text file (go template)")
	flag.Var(&outJSON, "out-json", "path of output JSON file (go template)")
	flag.Parse()
}

func writeOutPng(sourcePath string, p []color.RGBA) {
	b := bytes.NewBuffer(nil)
	outPng.Value.Execute(b, map[string]interface{}{
		"Path":    sourcePath,
		"K":       k,
		"Palette": p,
	})
	targetPath := b.String()
	fOut, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, 0600)
	log.Println("writing", targetPath)
	if err != nil {
		log.Println(err)
	}
	defer fOut.Close()
	png.Encode(fOut, palette.Render(p, outColorSize))
}

func html(c color.RGBA) string {
	if c.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}

func htmls(p []color.RGBA) (out []string) {
	out = make([]string, len(p))
	for i := range p {
		out[i] = html(p[i])
	}
	return
}

func writeOutTxt(sourcePath string, p []color.RGBA) {
	b := bytes.NewBuffer(nil)
	outTxt.Value.Execute(b, map[string]interface{}{
		"Path":    sourcePath,
		"K":       k,
		"Palette": p,
	})
	targetPath := b.String()
	fOut, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, 0600)
	log.Println("writing", targetPath)
	if err != nil {
		log.Println(err)
	}
	defer fOut.Close()
	for _, c := range p {
		io.WriteString(fOut, html(c))
		io.WriteString(fOut, "\n")
	}
}

func writeOutJSON(sourcePath string, p []color.RGBA, obj interface{}) {
	b := bytes.NewBuffer(nil)
	outJSON.Value.Execute(b, map[string]interface{}{
		"Path":    sourcePath,
		"K":       k,
		"Palette": p,
	})
	targetPath := b.String()
	fOut, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, 0600)
	log.Println("writing", targetPath)
	if err != nil {
		log.Println(err)
		return
	}
	defer fOut.Close()
	bytes, err := json.Marshal(obj)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = fOut.Write(bytes)
	if err != nil {
		log.Println(err)
		return
	}
}

func extractPalette(path string) ([]color.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Println(path, "error:", err)
		return nil, err
	}
	defer f.Close()
	i, typ, err := image.Decode(f)
	if err != nil {
		log.Println(path, "error:", err)
		return nil, err
	}
	log.Println(path, "loaded:", typ, i.Bounds().Size().String())
	return palette.Extract(cache, k, i)
}

func main() {
	print := make(chan interface{}, printBuffer)
	var printWg sync.WaitGroup

	printWg.Add(1)
	go func() {
		enc := json.NewEncoder(os.Stdout)
		defer printWg.Done()
		for obj := range print {
			enc.Encode(obj)
		}
	}()

	var workWg sync.WaitGroup
	work := make(chan string, maxParallel)
	workWg.Add(maxParallel)
	for i := 0; i < maxParallel; i++ {
		go func() {
			defer workWg.Done()
			for path := range work {
				p, err := extractPalette(path)
				if err != nil {
					log.Println(path, "error:", err)
					continue
				}
				if outPng.Value != nil {
					writeOutPng(path, p)
				}
				if outTxt.Value != nil {
					writeOutTxt(path, p)
				}
				jsonObj := map[string]interface{}{
					"path":    path,
					"palette": htmls(p),
				}
				if outJSON.Value != nil {
					writeOutJSON(path, p, jsonObj)
				}
				log.Println(path, htmls(p))
				print <- jsonObj
			}
		}()
	}
	for _, path := range flag.Args() {
		work <- path
	}
	close(work)
	workWg.Wait()
	close(print)
	printWg.Wait()
}
