package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// These structs define what our JSON response will look like
type ArchiveInfo struct {
	Filename    string     `json:"filename"`     // Name of the ZIP file
	ArchiveSize float64    `json:"archive_size"` // Size of the ZIP file
	TotalSize   float64    `json:"total_size"`   // Size of all files when unzipped
	TotalFiles  float64    `json:"total_files"`  // Number of files in ZIP
	Files       []FileInfo `json:"files"`        // List of files inside
}

type FileInfo struct {
	FilePath string  `json:"file_path"` // Path to file inside ZIP
	Size     float64 `json:"size"`      // Size of the file
	MimeType string  `json:"mimetype"`  // Type of file (jpeg, pdf, etc)
}

// This is your main API handler function
func GetArchiveInformation(w http.ResponseWriter, r *http.Request) {
	// 1. Check if the request is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Cannot read file", http.StatusInternalServerError)
		return
	}

	// 4. Try to open it as a ZIP file
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		http.Error(w, "Not a valid ZIP file", http.StatusBadRequest)
		return
	}

	// 5. Prepare our response
	response := ArchiveInfo{
		Filename:    header.Filename,
		ArchiveSize: float64(len(content)),
		Files:       make([]FileInfo, 0),
	}

	// 6. Look at each file in the ZIP
	for _, f := range reader.File {
		// Skip folders
		if f.FileInfo().IsDir() {
			continue
		}

		// Add file information to our response
		fileInfo := FileInfo{
			FilePath: f.Name,
			Size:     float64(f.UncompressedSize64),
			MimeType: "application/octet-stream", // Default type
		}

		response.Files = append(response.Files, fileInfo)
		response.TotalSize += float64(f.UncompressedSize64)
	}

	// 7. Calculate total number of files
	response.TotalFiles = float64(len(response.Files))

	// 8. Send the JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
