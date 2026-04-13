package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
)

func processImage(imagePath string) (*image.NRGBA, error) {
	img, err := imaging.Open(imagePath)
	if err != nil {
		fmt.Println("Error opening image:", err)
		return nil, err
	}

	dst := imaging.New(img.Bounds().Dx(), img.Bounds().Dy(), color.White)

	dst = imaging.Overlay(dst, img, image.Pt(0, 0), 1.0)

	return dst, nil

}

func getPathsFromDir(dir string) ([]string, error) {

	var validExts = map[string]struct{}{
		".png":  {},
		".jpg":  {},
		".jpeg": {},
		".avif": {},
		".webp": {},
		".gif":  {},
		".tiff": {},
		".bmp":  {},
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		// filter only images
		if _, ok := validExts[ext]; ok {
			fullPath := filepath.Join(dir, name)
			paths = append(paths, fullPath)
		}
	}
	return paths, nil
}

func processBatch(imagePaths []string, saveDir string, fileType string) error {

	fileType = strings.ToLower(fileType)

	var validExts = map[string]struct{}{
		".png":  {},
		".jpg":  {},
		".jpeg": {},
		".avif": {},
		".webp": {},
		".gif":  {},
		".tiff": {},
		".tif":  {},
		".bmp":  {},
	}

	if _, ok := validExts["."+fileType]; !ok {
		return fmt.Errorf("unsupported file type: %s", fileType)
	}

	errCh := make(chan error, len(imagePaths))

	err := os.MkdirAll(saveDir, 0755)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	const workers = 4
	sem := make(chan struct{}, workers)

	for _, path := range imagePaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			processedImage, err := processImage(p)
			if err != nil {
				errCh <- err
				return
			}

			sem <- struct{}{}
			defer func() { <-sem }()
			ext := filepath.Ext(p)

			filename := strings.TrimSuffix(filepath.Base(p), ext)

			outPath := fmt.Sprintf("%s/BG White %s.%s", saveDir, filename, fileType)

			if err := imaging.Save(processedImage, outPath); err != nil {
				errCh <- err
				return
			}

		}(path)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	// imagePaths := []string{"images/Haruka haruka11_sticker.png", "images/Luka luka5_sticker 2.png", "images/Luka luka5_sticker.png", "images/Miku miku1_sticker.png"}
	imagePaths, err := getPathsFromDir("images")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	if err := processBatch(imagePaths, "processed_images", "jpeg"); err != nil {
		fmt.Println("Error processing batch:", err)
	} else {
		fmt.Println("Batch processing completed successfully.")
	}
}
