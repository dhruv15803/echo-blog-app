package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var (
	FILE_KEY string = "imageFile"
)

func (h *Handler) UploadFileHandler(w http.ResponseWriter, r *http.Request) {

	// get file from request (Content-Type:"multipart/form-data")
	// store file on server
	// try upload to cloudinary
	// send url response back to client

	file, fileHeader, err := r.FormFile(FILE_KEY)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("failed to read multipart/form-data file with key %s", FILE_KEY), http.StatusBadRequest)
		return
	}

	log.Printf("Filename :- %s", fileHeader.Filename)
	log.Printf("File Size :- %v", fileHeader.Size)

	// create file on server
	serverImageFilePath := fmt.Sprintf("./uploads/%s", fileHeader.Filename)

	dest, err := os.Create(serverImageFilePath)
	if err != nil {
		log.Printf("failed to create dest file on server :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(dest, file)
	if err != nil {
		log.Println("failed to copy contents from multipart file to destination file")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	file.Close()
	dest.Close()

	// dest file has the file contents
	// open dest file and upload to cloudinary

	uploadFile, err := os.Open(serverImageFilePath)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	maxRetries := 3
	retryCount := 0
	isImageFileUploaded := false
	var result *uploader.UploadResult

	for retryCount < maxRetries {
		result, err = h.cld.Upload.Upload(context.TODO(), uploadFile, uploader.UploadParams{})
		if err != nil {
			log.Printf("failed to upload file to cloudinary, retry count = %v", retryCount+1)
			retryCount++
		}

		// successfully uploaded
		isImageFileUploaded = true
		if isImageFileUploaded {
			break
		}
	}

	if !isImageFileUploaded {
		log.Printf("failed to upload file to cloudinary :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	uploadFile.Close()

	if err = os.Remove(serverImageFilePath); err != nil {
		log.Printf("failed to remove file :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool   `json:"success"`
		Url     string `json:"url"`
	}

	if err := writeJSON(w, Response{Success: true, Url: result.SecureURL}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
