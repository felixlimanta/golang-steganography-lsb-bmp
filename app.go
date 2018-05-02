package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

var mask = [...]byte{128, 64, 32, 16, 8, 4, 2, 1}

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})

	r.POST("/encode", func(c *gin.Context) {
		file, _, err := c.Request.FormFile("image")
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusNotFound)
			panic(err)
		}
		defer file.Close()

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			panic(err)
		}

		img, err := encodeMessage(buf.Bytes()[:], c.PostForm("message"))
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusBadRequest)
			panic(err)
		}

		c.Writer.WriteString(base64.StdEncoding.EncodeToString(img))
	})

	r.POST("/decode", func(c *gin.Context) {
		file, _, err := c.Request.FormFile("image")
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusNotFound)
			panic(err)
		}
		defer file.Close()

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			panic(err)
		}

		msg := decodeMessage(buf.Bytes()[:])

		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.WriteString(msg)
	})

	r.Run(":5000")
}

func encodeMessage(img []byte, message string) ([]byte, error) {
	m := []byte(message)

	if len(message)*8 > len(img)-54 {
		return nil, errors.New("Error: image is not large enough to hold this message")
	}

	for i := 0; i < len(m); i++ {
		index := 55 + i*8
		b := m[i]

		for j := 0; j < 8; j++ {
			if b&mask[j] == 0 {
				img[index+j] = setLSB(0, img[index+j])
			} else {
				img[index+j] = setLSB(1, img[index+j])
			}
		}

		if i == len(m)-1 {
			for j := 8; j < 16; j++ {
				img[index+j] = setLSB(0, img[index+j])
			}
		}
	}

	return img, nil
}

func decodeMessage(img []byte) string {
	message := ""

	for i := 55; i < len(img)-9; i += 8 {
		var letter byte
		for j := 0; j < 8; j++ {
			b := img[i+j]
			if b%2 == 0 {
				letter &^= 1
			} else {
				letter |= 1
			}

			if j != 7 {
				letter = letter << 1
			}
		}

		if letter == 0 {
			break
		}

		message += string(letter)
	}
	return message
}

func setLSB(b byte, val byte) byte {
	if b != 0 {
		val |= 1
	} else {
		val &^= 1
	}
	return val
}
