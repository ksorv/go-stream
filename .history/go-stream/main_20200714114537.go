package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/xfrr/goffmpeg/transcoder"
)

const uploadPath = "in"
const mediaRoot = "m3u8s"

func main() {
	http.Handle("/", handlers())
	http.ListenAndServe(":8000", nil)
}

func handlers() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", indexPage).Methods("GET")
	router.HandleFunc("/media/{folder}/stream/index.m3u8", streamHandler).Methods("GET")
	router.HandleFunc("/media/{folder}/stream/{segName:index[0-2]+.ts}", streamHandler).Methods("GET")
	router.HandleFunc("/upload", uploadHandler)
	return router
}

func uploadHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		template, _ := template.ParseFiles("upload.gtpl")
		template.Execute(response, nil)
		return
	}

	fileName := randToken(12)

	// Parse
	file, fileHeader, err := request.FormFile("uploadFile")
	if err != nil {
		renderError(response, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get & Print out file size
	fileSize := fileHeader.Size
	fmt.Printf("File size (bytes): %v\n", fileSize)

	// FileByte *!---!*
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		renderError(response, "INVALID_FILE", http.StatusBadRequest)
		return
	}

	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "video/mp4":
		break
	default:
		renderError(response, "INVALID_FILE_TYPE", http.StatusBadRequest)
		return
	}

	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	if err != nil {
		renderError(response, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
		return
	}
	newPath := filepath.Join(uploadPath, fileName+fileEndings[0])
	fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		renderError(response, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}

	defer newFile.Close() // idempotent, okay to call twice
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		renderError(response, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}

	_, err := os.Stat(fmt.Sprintf("m3u8s/%s", fileName))

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(fmt.Sprintf("m3u8s/%s", fileName), 0755)
		if errDir != nil {
			log.Fatal(err)
		}

	}

	trans := new(transcoder.Transcoder)
	trans.InitializeEmptyTranscoder()
	err = trans.Initialize(newPath, fmt.Sprintf("/m3u8s/%s/%s.m3u8", fileName, fileName))
	trans.MediaFile().SetAspect("640x360")
	trans.MediaFile().SetHlsListSize(0)
	trans.MediaFile().SetHlsSegmentDuration(5)

	if err != nil {
		renderError(response, "CANT_TRANSCODE_FILE", http.StatusInternalServerError)
		return
	}

	// Start transcoder process with progress checking
	done := trans.Run(true)

	// Returns a channel to get the transcoding progress
	progress := trans.Output()

	// Example of printing transcoding progress
	for msg := range progress {
		fmt.Println(msg)
	}

	// This channel is used to wait for the transcoding process to end
	err = <-done

	if err == nil {
		response.Write([]byte(fmt.Sprintf("SUCCESS, id: %s", fileName)))
	} else {
		response.Write([]byte(fmt.Sprintf("Failed: %s", err)))
		fmt.Print(err)
	}

}

func streamHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	folder := vars["folder"]

	segName, ok := vars["segName"]
	if !ok {
		mediaBase := getMediaBase(folder)
		m3u8Name := "index.m3u8"
		serveHlsM3u8(response, request, mediaBase, m3u8Name)
	} else {
		mediaBase := getMediaBase(folder)
		serveHlsTs(response, request, mediaBase, segName)
	}
}

func indexPage(response http.ResponseWriter, request *http.Request) {
	http.ServeFile(response, request, "index.html")
}

func getMediaBase(folder string) string {
	return fmt.Sprintf("%s/%s", mediaRoot, folder)
}

func serveHlsM3u8(response http.ResponseWriter, request *http.Request, mediaBase, m3u8Name string) {
	mediaFile := fmt.Sprintf("%s/%s", mediaBase, m3u8Name)
	http.ServeFile(response, request, mediaFile)
	response.Header().Set("Content-Type", "application/x-mpegURL")
}

func serveHlsTs(response http.ResponseWriter, request *http.Request, mediaBase, segName string) {
	mediaFile := fmt.Sprintf("%s/%s", mediaBase, segName)
	http.ServeFile(response, request, mediaFile)
	response.Header().Set("Content-Type", "video/MP2T")
}

func renderError(response http.ResponseWriter, message string, statusCode int) {
	response.WriteHeader(http.StatusBadRequest)
	response.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
