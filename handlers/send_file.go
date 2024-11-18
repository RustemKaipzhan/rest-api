package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

var allowedMimeTypes1 = map[string]bool{
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
}

// SendFileToEmails handles file upload and sends the file to email addresses
func SendFileToEmails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (10MB max for files)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !isAllowedMimeType(header) {
		http.Error(w, "Unsupported file type", http.StatusBadRequest)
		return
	}

	// Get email addresses from form data
	emailsStr := r.FormValue("emails")
	emails := strings.Split(emailsStr, ",")
	if len(emails) == 0 {
		http.Error(w, "No emails provided", http.StatusBadRequest)
		return
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Send emails using SMTP
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	for _, email := range emails {
		err := sendEmailWithAttachment(smtpHost, smtpPort, auth, smtpUsername, email, header.Filename, fileContent, header.Header.Get("Content-Type"))
		if err != nil {
			// Log and continue to the next email
			fmt.Printf("Error sending email to %s: %v\n", email, err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Check if the file MIME type is allowed
func isAllowedMimeType(header *multipart.FileHeader) bool {
	return allowedMimeTypes1[header.Header.Get("Content-Type")]
}

// Send an email with an attachment
func sendEmailWithAttachment(smtpHost, smtpPort string, auth smtp.Auth, smtpUsername, toEmail, filename string, fileContent []byte, mimeType string) error {
	boundary := "boundary42"

	// Create email message with multipart/mixed encoding
	msg := []byte(
		"From: " + smtpUsername + "\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: File Attachment\r\n" +
			"Content-Type: multipart/mixed; boundary=" + boundary + "\r\n" +
			"\r\n" +
			"--" + boundary + "\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			"Please find the attached file.\r\n" +
			"--" + boundary + "\r\n" +
			"Content-Type: " + mimeType + "\r\n" +
			"Content-Disposition: attachment; filename=\"" + filename + "\"\r\n" +
			"Content-Transfer-Encoding: base64\r\n" +
			"\r\n" +

			base64.StdEncoding.EncodeToString(fileContent) + "\r\n" +
			"--" + boundary + "--\r\n")

	return smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		smtpUsername,
		[]string{toEmail},
		msg,
	)
}
