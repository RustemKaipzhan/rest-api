package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Allowed MIME types
var allowedMimeTypes = map[string]bool{
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/xml": true,
	"image/jpeg":      true,
	"image/png":       true,
}

func CreateArchive(w http.ResponseWriter, r *http.Request) {
	// 1. Check if it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Parse the multipart form (32MB max)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// 3. Get the files from form
	files := r.MultipartForm.File["files[]"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	// 4. Check MIME types
	for _, fileHeader := range files {
		if !allowedMimeTypes[fileHeader.Header.Get("Content-Type")] {
			errorMsg := fmt.Sprintf("File type not allowed: %s", fileHeader.Filename)
			http.Error(w, errorMsg, http.StatusBadRequest)
			return
		}
	}

	// 5. Create a buffer to write our archive to
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// 6. Add files to zip
	for _, fileHeader := range files {
		// Open uploaded file
		uploadedFile, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error processing file", http.StatusInternalServerError)
			return
		}
		defer uploadedFile.Close()

		// Create file in zip
		zipFile, err := zipWriter.Create(fileHeader.Filename)
		if err != nil {
			http.Error(w, "Error creating zip", http.StatusInternalServerError)
			return
		}

		// Copy file content to zip
		if _, err := io.Copy(zipFile, uploadedFile); err != nil {
			http.Error(w, "Error adding file to zip", http.StatusInternalServerError)
			return
		}
	}

	// 7. Close the zip writer
	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Error finalizing zip", http.StatusInternalServerError)
		return
	}

	// 8. Send the zip file
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=archive.zip")
	w.Write(buf.Bytes())
}
