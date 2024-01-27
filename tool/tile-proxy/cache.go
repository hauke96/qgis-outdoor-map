package tile_proxy

import (
	"errors"
	"fmt"
	"github.com/hauke96/sigolo"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func getTile(z, x, y, cacheKey, remoteFormat, cacheBaseFolder string, log *logger) []byte {
	cachePath := filepath.Join(cacheBaseFolder, cacheKey)

	imageFolder := ensureFolderExists(z, x, cachePath)
	imageFilePath := filepath.Join(imageFolder, y+"."+remoteFormat)
	if _, err := os.Stat(imageFilePath); errors.Is(err, os.ErrNotExist) {
		// Image does not exist
		return nil
	}

	fileContent, err := os.ReadFile(imageFilePath)
	if err != nil {
		log.Error("Error reading cached image file %s from disk. Pretend it's not cached. Error: %s", imageFilePath, err.Error())
		return nil
	}

	return fileContent
}

func cacheTile(z, x, y, cacheKey, remoteFormat, cacheBaseFolder string, image []byte) error {
	cachePath := filepath.Join(cacheBaseFolder, cacheKey)

	imageFolder := ensureFolderExists(z, x, cachePath)
	imageFilePath := filepath.Join(imageFolder, y+"."+remoteFormat)

	err := os.WriteFile(imageFilePath, image, 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("Error writing image file to %s: %s", imageFilePath, err.Error()))
	}

	return nil
}

func ensureFolderExists(z string, x string, cachePath string) string {
	imageFolder := filepath.Join(cachePath, z, x)
	err := os.MkdirAll(imageFolder, os.ModePerm)
	sigolo.FatalCheck(err)
	return imageFolder
}

func toCacheKey(targetUrl *url.URL) string {
	cacheKey := targetUrl.Host + targetUrl.Path
	cacheKey = strings.ReplaceAll(cacheKey, "/", "_")
	cacheKey = strings.ReplaceAll(cacheKey, ".", "-")
	cacheKey = strings.ReplaceAll(cacheKey, "{", "")
	cacheKey = strings.ReplaceAll(cacheKey, "}", "")
	return cacheKey
}
