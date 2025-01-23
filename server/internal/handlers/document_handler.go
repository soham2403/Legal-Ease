package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
	"github.com/otiai10/gosseract/v2"
)

type DocumentTextExtractor struct {
	FilePath string
}

func UploadDocument(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // Limit upload size to 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Print file info (Optional: For debugging purposes)
	fmt.Printf("Uploaded File: %s\n", header.Filename)
	fmt.Printf("File Size: %d bytes\n", header.Size)
	fmt.Printf("MIME Header: %v\n", header.Header)

	// Save the file temporarily
	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, header.Filename)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		http.Error(w, "Unable to create temp file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// Write the file content to the temp file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	// Create a new DocumentTextExtractor
	extractor := NewDocumentTextExtractor(tempFilePath)

	// Extract text using the method on the extractor
	text, err := extractor.Extract()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error extracting text: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the extracted text
	w.Write([]byte(text))
}

func NewDocumentTextExtractor(filePath string) *DocumentTextExtractor {
	return &DocumentTextExtractor{
		FilePath: filePath,
	}
}

// Extract determines the file type and calls the appropriate extraction method
func (d *DocumentTextExtractor) Extract() (string, error) {
	ext := strings.ToLower(filepath.Ext(d.FilePath))

	switch ext {
	case ".pdf":
		return d.ExtractPDF()
	case ".doc", ".docx":
		return d.ExtractDOC()
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}
}

// extractPDF handles PDF text extraction
func (d *DocumentTextExtractor) ExtractPDF() (string, error) {
	// If the PDF is a scanned copy, we need to convert it to images and run OCR
	if isScannedPDF(d.FilePath) {
		return d.ExtractTextFromScannedPDF()
	}
	// Process as regular text-based PDF
	f, r, err := pdf.Open(d.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var text strings.Builder
	totalPages := r.NumPage()
	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("failed to extract text from page %d: %w", pageNum, err)
		}
		text.WriteString(pageText)
	}
	return text.String(), nil
}

// Check if the PDF is a scanned copy (image-based content detection)
func isScannedPDF(filePath string) bool {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	// Check if the PDF contains only images or has a text layer
	for pageNum := 1; pageNum <= r.NumPage(); pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		// Use GetPlainText to extract text from the page
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			continue // If there's an error extracting text, skip this page
		}

		// If any text is found on the page, it's likely not a scanned PDF
		if len(pageText) > 0 {
			return false
		}
	}

	// If no text is found, consider it a scanned PDF
	return true
}

// extractTextFromScannedPDF extracts text from a scanned PDF using OCR
func (d *DocumentTextExtractor) ExtractTextFromScannedPDF() (string, error) {
	// Convert the PDF to images
	err := ConvertPDFToImages(d.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to convert PDF to images: %w", err)
	}
	// Perform OCR on the images
	return d.PerformOCROnImages()
}

// Convert PDF to images using poppler's pdftoppm
func ConvertPDFToImages(pdfFilePath string) error {
	// Using pdftoppm to convert PDF to images
	cmd := exec.Command("pdftoppm", pdfFilePath, "output_image", "-png")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run pdftoppm: %w", err)
	}
	return nil
}

// Perform OCR on converted images
func (d *DocumentTextExtractor) PerformOCROnImages() (string, error) {
	client := gosseract.NewClient()
	defer client.Close()
	// Example: OCR for images output_image-1.png, output_image-2.png, etc.
	var text strings.Builder
	for i := 1; ; i++ {
		imagePath := fmt.Sprintf("output_image-%d.png", i)
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			break // No more images
		}
		client.SetImage(imagePath)
		ocrText, err := client.Text()
		if err != nil {
			return "", fmt.Errorf("failed to extract text from image %s: %w", imagePath, err)
		}
		text.WriteString(ocrText)
		text.WriteString("\n")
	}
	return text.String(), nil
}

// extractDOC handles DOC/DOCX text extraction using the docx library
func (d *DocumentTextExtractor) ExtractDOC() (string, error) {
	// Open the DOCX file using the docx library
	r, err := docx.ReadDocxFile(d.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open DOCX document: %w", err)
	}
	defer r.Close()

	// Extract the content from the DOCX file
	docxContent := r.Editable()
	return docxContent.GetContent(), nil
}
