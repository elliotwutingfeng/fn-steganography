package function

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/auyer/steganography"
)

var urlRegEx, _ = regexp.Compile(`\b((http|https):\/\/?)[^\s()<>]+(?:\([\w\d]+\)|([^[:punct:]\s]|\/?))`)
var b64RegEx, _ = regexp.Compile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)

// requestData structure holds the Input Data.
/*
Message string : the message that will be enbedded in the image
Image string : Base64 encoded image
*/
type requestData struct {
	Message string `json:"message"`
	Image   string `json:"image"`
	Encode  bool   `json:"encode"`
}

func getImage(inputURL string) (string, error) {
	res, err := http.Get(inputURL)
	if err != nil {
		return fmt.Sprintf(`{"error" : "Unable to download image file from URI: %s"}`, inputURL), err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Sprintf(`{"error" : "Unable to read response body: %s"}`, err), err
	}
	// fmt.Println(string(data))
	return string(data), nil
}

// Handle a serverless request
func Handle(req []byte) string {
	// Reading the body
	data := requestData{}
	err := json.Unmarshal(req, &data)
	// If the body is not correct
	if err != nil {
		log.Println(fmt.Sprintf("error: bad body. %s ", err.Error()))
		return fmt.Sprintf(`{"error": "bad body. %s"}`, err.Error())
	}
	fmt.Println(strings.TrimSpace(data.Image))
	if urlRegEx.Match([]byte(strings.TrimSpace(data.Image))) {
		data.Image, err = getImage(strings.TrimSpace(data.Image))
		if err != nil {
			return data.Image
		}
		return encodeDecode(data)
	}
	if b64RegEx.Match([]byte(data.Image)) {
		decodedImg, _ := base64.StdEncoding.DecodeString(data.Image) // no need for error handling since passed regex check
		data.Image = string(decodedImg)
		return encodeDecode(data)
	}
	return `{"error": "image field didnt match a URL or base64 image"}`
}

func encodeDecode(data requestData) string {

	img, _, err := image.Decode(strings.NewReader(data.Image))
	if err != nil {
		log.Println(fmt.Sprintf("error: failed decoding image from base64. %s", err.Error()))
		return fmt.Sprintf(`{"error": "failed decoding image from base64. %s"}`, err.Error())
	}

	if data.Encode {

		buff := new(bytes.Buffer)

		err = steganography.Encode(buff, img, []byte(data.Message))

		if err != nil {
			log.Println(fmt.Sprintf("error: failed encoding message to image. %s ", err.Error()))
			return fmt.Sprintf(`{"error": "failed encoding message to image. %s"}`, err.Error())
		}
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buff.Bytes())
	}
	// DECODING MODE
	message := steganography.Decode(steganography.GetMessageSizeFromImage(img), img)
	return string(message)
}
