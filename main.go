package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/generated", generatedHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/pdf/", http.StripPrefix("/pdf/", http.FileServer(http.Dir("static/pdf"))))

	fmt.Println("Starting server at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("from: %s, url: %s", r.RemoteAddr, r.URL)
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	chapter := r.FormValue("chapter")
	if chapter == "" {
		http.Error(w, "You need to provide a chapter number", http.StatusBadRequest)
		return
	}

	success := false
	for i := 1; i <= 30; i++ {
		page := fmt.Sprintf("%02d", i)
		url := fmt.Sprintf("https://lelscans.net/mangas/one-piece/%s/%s.jpg", chapter, page)
		fileName := fmt.Sprintf("onepiece_%s_%s.jpg", chapter, page)

		if err := downloadFile(url, fileName); err != nil {
			if i == 1 {
				break
			}
			continue
		}
		success = true
	}

	if !success {
		http.ServeFile(w, r, "static/error.html")
		return
	}

	if err := generatePDF(chapter); err != nil {
		http.Error(
			w,
			fmt.Sprintf("Failed to generate PDF: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	if err := cleanupImages(chapter); err != nil {
		http.Error(
			w,
			fmt.Sprintf("Failed to clean up images: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	http.ServeFile(w, r, fmt.Sprintf("static/pdf/%s.pdf", chapter))
}

func downloadFile(url, fileName string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func generatePDF(chapter string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	for i := 1; i <= 30; i++ {
		page := fmt.Sprintf("%02d", i)
		fileName := fmt.Sprintf("onepiece_%s_%s.jpg", chapter, page)
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			continue
		}
		pdf.AddPage()
		pdf.Image(fileName, 10, 10, 190, 0, false, "", 0, "")
	}
	return pdf.OutputFileAndClose(fmt.Sprintf("static/pdf/%s.pdf", chapter))
}

func generatedHandler(w http.ResponseWriter, r *http.Request) {
	files, err := filepath.Glob("static/pdf/*.pdf")
	if err != nil {
		http.Error(w, "Failed to list generated PDFs", http.StatusInternalServerError)
		return
	}

	var pdfs []string
	for _, file := range files {
		pdfs = append(pdfs, filepath.Base(file))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pdfs)
}

func cleanupImages(chapter string) error {
	for i := 1; i <= 30; i++ {
		page := fmt.Sprintf("%02d", i)
		fileName := fmt.Sprintf("onepiece_%s_%s.jpg", chapter, page)
		if err := os.Remove(fileName); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
