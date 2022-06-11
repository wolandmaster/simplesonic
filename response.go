package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

const coverJpegQuality = 90

type Exchange struct {
	Request        *http.Request
	Response       *Response
	requestTime    time.Time
	responseWriter http.ResponseWriter
}

func init() {
	ProcessError(mime.AddExtensionType(".json", "application/json"))

	ProcessError(mime.AddExtensionType(".mp3", "audio/mpeg"))
	ProcessError(mime.AddExtensionType(".m4a", "audio/mpeg"))
	ProcessError(mime.AddExtensionType(".flac", "audio/flac"))
	ProcessError(mime.AddExtensionType(".ogg", "audio/ogg"))
	ProcessError(mime.AddExtensionType(".opus", "audio/ogg"))
	ProcessError(mime.AddExtensionType(".oga", "audio/ogg"))
	ProcessError(mime.AddExtensionType(".aac", "audio/aac"))
	ProcessError(mime.AddExtensionType(".wav", "audio/x-wav"))
	ProcessError(mime.AddExtensionType(".wma", "audio/x-ms-wma"))

	ProcessError(mime.AddExtensionType(".mp4", "video/mp4"))
	ProcessError(mime.AddExtensionType(".m4v", "video/mp4"))
	ProcessError(mime.AddExtensionType(".mpg", "video/mpeg"))
	ProcessError(mime.AddExtensionType(".webm", "video/webm"))
	ProcessError(mime.AddExtensionType(".mkv", "video/x-matroska"))
	ProcessError(mime.AddExtensionType(".avi", "video/x-msvideo"))
	ProcessError(mime.AddExtensionType(".wmv", "video/x-ms-wmv"))
	ProcessError(mime.AddExtensionType(".flv", "video/x-flv"))
	ProcessError(mime.AddExtensionType(".mov", "video/quicktime"))
	ProcessError(mime.AddExtensionType(".3gp", "video/3gpp"))

	ProcessError(mime.AddExtensionType(".m3u", "audio/x-mpegurl"))
	ProcessError(mime.AddExtensionType(".m3u8", "application/x-mpegURL"))
}

func RegisterHandler(pattern string, handler func(Exchange)) {
	http.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		exchange := Exchange{Request: request, Response: NewResponse(), requestTime: time.Now(), responseWriter: writer}
		log.Printf("Request: %s\n", request.URL)
		if request.URL.Scheme = "http"; request.TLS != nil {
			request.URL.Scheme += "s"
		}
		defer func() {
			if p := recover(); p != nil {
				fmt.Printf("%v: %s\n", p, string(debug.Stack()))
				exchange.SendError(0, "An error happened; check the logs!")
			}
		}()
		if !verifyCredentials(exchange) {
			exchange.SendError(40, "Wrong username or password")
		} else {
			handler(exchange)
		}
	})
}

func (exchange Exchange) QueryGetInt(key string, defaultValue int) int {
	value := defaultValue
	if valueStr := exchange.Request.URL.Query().Get(key); valueStr != "" {
		value = int(ParseNumber(valueStr))
	}
	return value
}

func (exchange Exchange) SendResponse() {
	var response []byte
	if exchange.Request.URL.Query().Get("f") == "json" {
		exchange.responseWriter.Header().Set("Content-Type", mime.TypeByExtension(".json"))
		response = ProcessErrorArg(json.Marshal(exchange.Response)).([]byte)
	} else {
		exchange.responseWriter.Header().Set("Content-Type", mime.TypeByExtension(".xml"))
		response = ProcessErrorArg(xml.Marshal(exchange.Response)).([]byte)
	}
	n := ProcessErrorArg(exchange.responseWriter.Write(response)).(int)
	log.Printf("Response (%d bytes, %v): %s\n\n", n, time.Since(exchange.requestTime), response)
}

func (exchange Exchange) SendFile(filename string) {
	file := ProcessErrorArg(os.Open(filename)).(*os.File)
	exchange.responseWriter.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filename)))
	n := ProcessErrorArg(io.Copy(exchange.responseWriter, file)).(int64)
	log.Printf("Response (%d bytes, %v): file: %s", n, time.Since(exchange.requestTime), filename)
}

func (exchange Exchange) SendJpeg(img image.Image) {
	var responseJpeg bytes.Buffer
	ProcessError(jpeg.Encode(&responseJpeg, img, &jpeg.Options{Quality: coverJpegQuality}))
	exchange.responseWriter.Header().Set("Content-Type", mime.TypeByExtension(".jpg"))
	n := ProcessErrorArg(exchange.responseWriter.Write(responseJpeg.Bytes())).(int)
	log.Printf("Response (%d bytes, %v): jpeg: %d x %d px",
		n, time.Since(exchange.requestTime), img.Bounds().Size().X, img.Bounds().Size().Y)
}

func (exchange Exchange) SendPng(img image.Image) {
	var responsePng bytes.Buffer
	ProcessError(png.Encode(&responsePng, img))
	exchange.responseWriter.Header().Set("Content-Type", mime.TypeByExtension(".png"))
	n := ProcessErrorArg(exchange.responseWriter.Write(responsePng.Bytes())).(int)
	log.Printf("Response (%d bytes, %v): png: %d x %d px",
		n, time.Since(exchange.requestTime), img.Bounds().Size().X, img.Bounds().Size().Y)
}

func (exchange Exchange) SendError(code int, message string) {
	exchange.Response.Status = Failed
	exchange.Response.Error = &Error{Code: code, Message: message}
	exchange.SendResponse()
}

func verifyCredentials(exchange Exchange) bool {
	var (
		username = exchange.Request.URL.Query().Get("u")
		password = exchange.Request.URL.Query().Get("p")
		token    = exchange.Request.URL.Query().Get("t")
		salt     = exchange.Request.URL.Query().Get("s")
	)
	if len(password) >= 4 && password[:4] == "enc:" {
		decodedPassword := ProcessErrorArg(hex.DecodeString(password[4:])).([]byte)
		password = string(decodedPassword)
	}
	for _, user := range Config.Users {
		if user.Username == username {
			if password != "" && user.Password == password {
				return true
			} else if token != "" && salt != "" {
				hash := md5.Sum([]byte(user.Password + salt))
				return hex.EncodeToString(hash[:]) == token
			}
		}
	}
	return false
}
