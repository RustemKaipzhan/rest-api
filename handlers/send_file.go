package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

var allowedMimeTypes2 = map[string]bool{
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/pdf": true,
}

func SendFileToEmails(w http.ResponseWriter, r *http.Request) {
	// Check POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file MIME type
	if !allowedMimeTypes2[header.Header.Get("Content-Type")] {
		http.Error(w, "Unsupported file type", http.StatusBadRequest)
		return
	}

	// Get emails
	emailsStr := r.FormValue("emails")
	emails := strings.Split(emailsStr, ",")
	if len(emails) == 0 {
		http.Error(w, "No emails provided", http.StatusBadRequest)
		return
	}

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Send emails
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	for _, email := range emails {
		msg := []byte(
			"To: " + email + "\r\n" +
				"Subject: File Attachment\r\n" +
				"Content-Type: application/octet-stream\r\n" +
				"Content-Disposition: attachment; filename=\"" + header.Filename + "\"\r\n\r\n",
		)
		msg = append(msg, fileContent...)

		err = smtp.SendMail(
			smtpHost+":"+smtpPort,
			auth,
			smtpUsername,
			[]string{email},
			msg,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error sending email to %s: %v", email, err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
