package tile_proxy

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hauke96/sigolo"
	"golang.org/x/image/webp"
	_ "golang.org/x/image/webp" // register webp format
	"image"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	formatWebp = "webp"
	formatPng  = "png"
	formatPbf  = "pbf"
)

func StartProxy(port string, remoteUrlString string, cacheBaseFolder string) {
	remoteUrl, err := url.Parse(remoteUrlString)
	sigolo.FatalCheck(err)

	remoteTileFormat := strings.Trim(path.Ext(remoteUrl.Path), ".")
	if remoteTileFormat != formatWebp && remoteTileFormat != formatPbf {
		sigolo.Fatal("Unsupported image format %s", remoteTileFormat)
	}

	sigolo.Info("Start tile proxy on port localhost:%s -> %s to convert images from %s", port, remoteUrlString, remoteTileFormat)

	client := http.Client{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Use local variable here to ensure each logging has exactly the counter it belongs to. Otherwise, subsequent
		// requests have increased the counter and concurrent requests print wrong log counter.
		log := newLogger(port)

		log.Debug("Request URL: %s", r.URL)

		pathSegmentsAndFormat := strings.Split(strings.Trim(r.URL.Path, "/"), ".")

		// Trim the ".png" ending off
		requestedPath := pathSegmentsAndFormat[0]
		requestedFormat := pathSegmentsAndFormat[1]

		log.Debug("Requested path  : %s", requestedPath)
		log.Debug("Requested format: %s", requestedFormat)

		if requestedFormat != formatPng && requestedFormat != formatPbf && requestedFormat != formatWebp {
			responseWithError(log, w, fmt.Sprintf("Unknown requested format %s", requestedFormat), err)
			return
		}

		segments := strings.Split(requestedPath, "/") // scheme: .../z/x/y.ext
		z := segments[0]
		x := segments[1]
		y := segments[2]

		cacheKey := toCacheKey(remoteUrl)
		tileBytes := getTile(z, x, y, cacheKey, remoteTileFormat, cacheBaseFolder, log)

		if tileBytes == nil {
			log.Debug("Tile not cached, load it from remote server")
			// Tile not in cache -> Request original tile and cache it
			tileBytes, err = requestOriginalTile(remoteUrlString, z, x, y, log, client)
			if err != nil {
				responseWithError(log, w, fmt.Sprintf("Error requesting original tile %s/%s/%s.%s: %s", z, x, y, remoteTileFormat, err.Error()), err)
				return
			}

			log.Debug("Cache new tile")
			err = cacheTile(z, x, y, cacheKey, remoteTileFormat, cacheBaseFolder, tileBytes)
			if err != nil {
				responseWithError(log, w, fmt.Sprintf("Error caching tile %s/%s/%s.%s: %s", z, x, y, remoteTileFormat, err.Error()), err)
				return
			}
		} else {
			log.Debug("Found tile in cache")
		}

		// Decode tile from remote format and encode it into the wanted request format.
		var remoteTile bytes.Buffer
		if remoteTileFormat == formatWebp {
			log.Debug("Decode tile as %s", formatWebp)
			var remoteTileImage image.Image
			remoteTileImage, err = webp.Decode(bytes.NewReader(tileBytes))
			if err != nil {
				responseWithError(log, w, fmt.Sprintf("Error decoding tile %s/%s/%s.%s as %s: %s", z, x, y, remoteTileFormat, requestedFormat, err.Error()), err)
				return
			}

			if requestedFormat == formatPng {
				err = png.Encode(&remoteTile, remoteTileImage)
				if err != nil {
					responseWithError(log, w, fmt.Sprintf("Error encoding tile %s/%s/%s.%s as %s: %s", z, x, y, remoteTileFormat, requestedFormat, err.Error()), err)
				}
			} else if requestedFormat == formatWebp {
				_, err = remoteTile.Write(tileBytes)
				if err != nil {
					responseWithError(log, w, fmt.Sprintf("Error writing raw bytes for tile %s/%s/%s.%s as %s: %s", z, x, y, remoteTileFormat, requestedFormat, err.Error()), err)
				}
			} else {
				responseWithError(log, w, fmt.Sprintf("Unsupported request format %s for tile %s/%s/%s.%s", requestedFormat, z, x, y, remoteTileFormat), nil)
			}
		} else if remoteTileFormat == formatPbf {
			log.Debug("Remote tile has format PBF, no format conversion performed")
			remoteTile.Write(tileBytes)
		}
		// No else: Remote formats are checked above

		err = writeTileToResponse(w, &remoteTile)
		if err != nil {
			log.Error("Error returning tile %s/%s/%s.%s: %s", z, x, y, remoteTileFormat, err.Error())
			return
		}
		log.Debug("Response written - Done")
	})

	err = http.ListenAndServe(":"+port, nil)
	sigolo.FatalCheck(err)
}

func requestOriginalTile(remoteUrlString string, z string, x string, y string, log *logger, client http.Client) ([]byte, error) {
	requestUrl := remoteUrlString
	requestUrl = strings.Replace(requestUrl, "{z}", z, 1)
	requestUrl = strings.Replace(requestUrl, "{x}", x, 1)
	requestUrl = strings.Replace(requestUrl, "{y}", y, 1)

	log.Debug("Make GET request to %s", requestUrl)

	resp, err := client.Get(requestUrl)
	if err != nil {
		log.Error("Error making GET request to %s: %s", requestUrl, err.Error())
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error reading response body: %s", err.Error()))
	}

	return content, nil
}

func writeTileToResponse(w http.ResponseWriter, responseBuf *bytes.Buffer) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Length", strconv.Itoa(responseBuf.Len()))

	_, err := responseBuf.WriteTo(w)
	if err != nil {
		return errors.New(fmt.Sprintf("Error writing buffer to response: %s", err.Error()))
	}

	return nil
}

func responseWithError(log *logger, w http.ResponseWriter, returnedMessage string, err error) {
	log.Errorb(1, "%s", returnedMessage)
	w.WriteHeader(500)
	w.Header().Set("Content-Type", "application/text")
	_, err = w.Write([]byte(returnedMessage))
	if err != nil {
		log.Error("Error returning error message: %s", log, err.Error())
	}
}
