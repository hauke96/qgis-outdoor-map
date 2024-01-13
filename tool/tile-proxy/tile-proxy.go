package tile_proxy

import (
	"bytes"
	"fmt"
	"github.com/hauke96/sigolo"
	"golang.org/x/image/webp"
	_ "golang.org/x/image/webp" // register webp format
	"image"
	"image/png"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	formatWebp = "webp"
	formatPng  = "png"
)

func StartProxy(port string, targetUrlString string) {
	targetUrl, err := url.Parse(targetUrlString)
	sigolo.FatalCheck(err)

	targetImageFormat := strings.Trim(path.Ext(targetUrl.Path), ".")
	if targetImageFormat != formatWebp {
		sigolo.Fatal("Unsupported image format %s", targetImageFormat)
	}

	sigolo.Info("Start tile proxy on port localhost:%s -> %s to convert images from %s", port, targetUrlString, targetImageFormat)

	client := http.Client{}

	globalRequestCounter := 0

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Use local variable here to ensure each logging has exactly the counter it belongs to. Otherwise, subsequent
		// requests have increased the counter and concurrent requests print wrong log counter.
		globalRequestCounter++
		requestCounter := globalRequestCounter

		sigolo.Debug("[%X] Request URL: %s", requestCounter, r.URL)

		pathSegmentsAndFormat := strings.Split(strings.Trim(r.URL.Path, "/"), ".")

		// Trim the ".png" ending off
		requestedPath := pathSegmentsAndFormat[0]
		requestedFormat := pathSegmentsAndFormat[1]

		sigolo.Debug("[%X] Requested path  : %s", requestCounter, requestedPath)
		sigolo.Debug("[%X] Requested format: %s", requestCounter, requestedFormat)

		if requestedFormat != formatPng {
			responseWithError(requestCounter, w, fmt.Sprintf("Unknown requested format %s", requestedFormat), err)
			return
		}

		segments := strings.Split(requestedPath, "/") // scheme: .../z/x/y.ext

		requestUrl := targetUrlString
		requestUrl = strings.Replace(requestUrl, "{z}", segments[0], 1)
		requestUrl = strings.Replace(requestUrl, "{x}", segments[1], 1)
		requestUrl = strings.Replace(requestUrl, "{y}", segments[2], 1)

		sigolo.Debug("[%X] Make GET request to %s", requestCounter, requestUrl)

		var resp *http.Response
		resp, err = client.Get(requestUrl)
		if err != nil {
			sigolo.Error("[%X] Error making GET request to %s: %s", requestCounter, requestUrl, err.Error())
		}

		var originalImage image.Image
		if targetImageFormat == formatWebp {
			originalImage, err = webp.Decode(resp.Body)
			if err != nil {
				responseWithError(requestCounter, w, fmt.Sprintf("Error decoding image: %s", err.Error()), err)
				return
			}
		}
		// No else: Target formats are checked above

		// Buffer to store newly encoded image into. This is then returned to the client and needed to determine the
		// size of the response content.
		buf := new(bytes.Buffer)

		err = resp.Header.WriteSubset(buf, map[string]bool{
			"Content-Length": true,
			"Content-Type":   true,
		})
		if err != nil {
			sigolo.Error("[%X] Error copying headers: %s", requestCounter, err.Error())
		}

		responseBuf := new(bytes.Buffer)
		if requestedFormat == formatPng {
			err = png.Encode(responseBuf, originalImage)
			if err != nil {
				sigolo.Error("[%X] Error returning PNG image: %s", requestCounter, err.Error())
			}
		}
		// No else: Target formats are checked above

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Length", strconv.Itoa(responseBuf.Len()))

		_, err = responseBuf.WriteTo(w)
		if err != nil {
			sigolo.Error("[%X] Error writing buffer to response: %s", requestCounter, err.Error())
		}
	})

	err = http.ListenAndServe(":"+port, nil)
	sigolo.FatalCheck(err)
}

func responseWithError(requestCounter int, w http.ResponseWriter, message string, err error) {
	sigolo.Errorb(1, "[%X] %s", requestCounter, message)
	w.WriteHeader(500)
	_, err = w.Write([]byte(message))
	if err != nil {
		sigolo.Error("[%X] Error returning error message: %s", requestCounter, err.Error())
	}
}
