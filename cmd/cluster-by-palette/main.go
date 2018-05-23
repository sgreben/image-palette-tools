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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"text/template"

	"github.com/kballard/go-shellquote"
	flagvarGlob "github.com/sgreben/flagvar/glob"
	"github.com/sgreben/flagvar/template"
	"github.com/sgreben/image-palette-tools/pkg/palette"
)

type paletteJSON struct {
	Path    string   `json:"path"`
	Palette []string `json:"palette"`
}

type pathPalette struct {
	Path    string
	Palette []color.RGBA
}

var (
	kImage         int
	kPalette       int
	outPngCluster  = flagvar.Template{Root: templateSettings}
	outPngSingle   = flagvar.Template{Root: templateSettings}
	inJSON         = flagvar.Template{Root: templateSettings}
	outClusterJSON = flagvar.Template{Root: templateSettings}
	outJSON        = flagvar.Template{Root: templateSettings}
	outShell       = flagvar.Template{Root: templateSettings}
	globSelect     flagvarGlob.Glob
	outColorSize   int
	maxParallel    int
	colorSortOrder = palette.LessLHS

	templateSettings = template.New("").Funcs(map[string]interface{}{
		"abs":      func(s string) (string, error) { return filepath.Abs(s) },
		"basename": func(s string) string { return filepath.Base(s) },
		"dirname":  func(s string) string { return filepath.Dir(s) },
		"ext":      func(s string) string { return filepath.Ext(s) },
		"html":     func(c color.RGBA) string { return html(c) },
	})
	colorCache    = palette.NewColorCache(512)
	paletteCache  = palette.NewPaletteCache(512)
	printBuffer   = 1024
	clusterBuffer = 16
)

const defaultInJSON = "{{.Path}}.palette-{{.K}}.json"
const defaultOutJSON = defaultInJSON

func init() {
	log.SetOutput(os.Stderr)
	flag.IntVar(&kImage, "n", 5, "number of image clusters to make")
	flag.IntVar(&kPalette, "k", 4, "palette size")
	flag.IntVar(&maxParallel, "p", runtime.GOMAXPROCS(0), "number of images to process in parallel")
	flag.Var(&globSelect, "glob", "glob expression matching image files to cluster")
	flag.Var(&inJSON, "in-json", "path to read palette JSON from (go template)")
	flag.Var(&outJSON, "out-json", "path to write palette JSON to (go template)")
	flag.Var(&outPngSingle, "out-png", "path of output palette image (PNG) (go template)")
	flag.Var(&outPngCluster, "out-cluster-png", "path of output cluster palette image (PNG) (go template)")
	flag.IntVar(&outColorSize, "out-cluster-png-height", 100, "size of each color square in the palette output image")
	flag.Var(&outClusterJSON, "out-summary-json", "path of output JSON containing the clustering (go template)")
	flag.Var(&outShell, "out-shell", "shell command to run for each image (go template)")
	flag.Parse()

	if inJSON.Value == nil && inJSON.Text != "" {
		inJSON.Set(defaultInJSON)
	}
	if outJSON.Value == nil && outJSON.Text != "" {
		outJSON.Set(defaultOutJSON)
	}
}

func writeOutPngSingle(sourcePath string, label int, p []color.RGBA) {
	b := bytes.NewBuffer(nil)
	outPngSingle.Value.Execute(b, map[string]interface{}{
		"Path":    sourcePath,
		"N":       kImage,
		"K":       kPalette,
		"Label":   label,
		"I":       label,
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

func writeOutJSON(sourcePath string, pp pathPalette) {
	b := bytes.NewBuffer(nil)
	outJSON.Value.Execute(b, map[string]interface{}{
		"Path":    sourcePath,
		"K":       kPalette,
		"Palette": pp.Palette,
	})
	targetPath := b.String()
	fOut, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, 0600)
	log.Println("writing", targetPath)
	if err != nil {
		log.Println(err)
		return
	}
	defer fOut.Close()
	var obj paletteJSON
	obj.Path = pp.Path
	obj.Palette = htmls(pp.Palette)
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

func writeOutClusterJSON(obj interface{}) {
	b := bytes.NewBuffer(nil)
	outClusterJSON.Value.Execute(b, map[string]interface{}{
		"N": kImage,
		"K": kPalette,
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

func runOutShell(path string, label int, p []color.RGBA) {
	b := bytes.NewBuffer(nil)
	outShell.Value.Execute(b, map[string]interface{}{
		"Path":    path,
		"N":       kImage,
		"K":       kPalette,
		"Label":   label,
		"I":       label,
		"Palette": p,
	})
	shellCmd := b.String()

	var cmd *exec.Cmd
	if shell, ok := os.LookupEnv("SHELL"); ok {
		cmd = exec.Command(shell, "-c", shellCmd)
	} else {
		parts, err := shellquote.Split(shellCmd)
		if err != nil {
			log.Println(path, "error:", err)
			return
		}
		if len(parts) < 1 {
			log.Println(path, "error:", "empty shell command")
			return
		}
		cmd = exec.Command(parts[0], parts[1:]...)
	}
	log.Println(path, "running", cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Println(path, "shell error:", err)
	}
}

func writeOutPng(label int, p []color.RGBA) {
	b := bytes.NewBuffer(nil)
	outPngCluster.Value.Execute(b, map[string]interface{}{
		"N":       kImage,
		"K":       kPalette,
		"Label":   label,
		"I":       label,
		"Palette": p,
	})
	targetPath := b.String()
	fOut, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, 0600)
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

func htmlss(ps [][]color.RGBA) (out [][]string) {
	out = make([][]string, len(ps))
	for i := range ps {
		out[i] = htmls(ps[i])
	}
	return
}

func extractPalette(path string) ([]color.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Println(path, "error:", err)
		return nil, err
	}
	defer f.Close()
	i, fmt, err := image.Decode(f)
	if err != nil {
		log.Println(path, "error:", err)
		return nil, err
	}
	log.Println(path, "loaded:", fmt, i.Bounds().Size().String())
	return palette.Extract(colorCache, kPalette, i)
}

func loadPalette(path string) ([]color.RGBA, error) {
	b := bytes.NewBuffer(nil)
	inJSON.Value.Execute(b, map[string]interface{}{
		"Path": path,
		"N":    kImage,
		"K":    kPalette,
	})
	targetPath := b.String()
	f, err := os.Open(targetPath)
	if err != nil {
		return nil, err
	}
	log.Println("loading", targetPath)
	dec := json.NewDecoder(f)
	var pj paletteJSON
	err = dec.Decode(&pj)
	if err != nil {
		return nil, err
	}
	out := make([]color.RGBA, len(pj.Palette))
	for i := range out {
		var c color.RGBA
		fmt.Sscanf(pj.Palette[i], "#%02x%02x%02x", &c.R, &c.G, &c.B)
		c.A = 0xFF
		out[i] = c
	}
	return out, nil
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

	cluster := make(chan pathPalette, clusterBuffer)
	var clusterWg sync.WaitGroup

	clusterWg.Add(1)
	go func() {
		var paths []string
		var palettes [][]color.RGBA
		defer clusterWg.Done()
		for pathPalette := range cluster {
			paths = append(paths, pathPalette.Path)
			sort.Slice(pathPalette.Palette, func(i int, j int) bool {
				return colorSortOrder(pathPalette.Palette, i, j)
			})
			palettes = append(palettes, pathPalette.Palette)
		}
		labels, centroids, err := palette.Cluster(paletteCache, kImage, palettes)
		if err != nil {
			log.Fatal(err)
		}
		if outPngCluster.Value != nil {
			for i, p := range centroids {
				writeOutPng(i, p)
			}
		}
		m := make(map[string]int, len(labels))
		for i, l := range labels {
			m[paths[i]] = l
			writeOutPngSingle(paths[i], l, palettes[i])
		}
		if outShell.Value != nil {
			for i, l := range labels {
				runOutShell(paths[i], l, palettes[i])
			}
		}
		obj := map[string]interface{}{
			"centroids": htmlss(centroids),
			"mapping":   m,
		}
		if outClusterJSON.Value != nil {
			writeOutClusterJSON(obj)
		}
		print <- obj
	}()

	var extractWg sync.WaitGroup
	work := make(chan string, maxParallel)
	extractWg.Add(maxParallel)
	for i := 0; i < maxParallel; i++ {
		go func() {
			defer extractWg.Done()
			var p []color.RGBA
			var err error
			for path := range work {
				shouldWriteOutJSON := outJSON.Value != nil
				if inJSON.Value != nil {
					p, err = loadPalette(path)
					if len(p) < kPalette || err != nil {
						p, err = extractPalette(path)
						if err != nil {
							log.Println(path, "error:", err)
							continue
						}
					} else {
						shouldWriteOutJSON = false
					}
				} else {
					p, err = extractPalette(path)
					if err != nil {
						log.Println(path, "error:", err)
						continue
					}
				}
				pp := pathPalette{Path: path, Palette: p}
				if shouldWriteOutJSON {
					writeOutJSON(path, pp)
				}
				cluster <- pp
			}
		}()
	}

	paths := flag.Args()
	if globSelect.Value != nil {
		globPaths, err := filepath.Glob(globSelect.Text)
		if err != nil {
			log.Println(err)
		}
		if len(globPaths) == 0 {
			log.Println("no paths matched glob", globSelect.Text)
		}
		paths = append(paths, globPaths...)
	}
	log.Println("processing", len(paths), "files")
	for i, path := range paths {
		log.Printf("image %d/%d (%2.2f%%)", i, len(paths), float64(i)/float64(len(paths))*100)
		work <- path
	}
	close(work)
	extractWg.Wait()
	close(cluster)
	clusterWg.Wait()
	close(print)
	printWg.Wait()
}
