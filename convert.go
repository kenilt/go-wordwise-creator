package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	TempDir      = "tempData"
	TempBookName = "book_dump"
)

var ebookConvertCmd string

func convertBookToHtml(inputPath string) {
	log.Println("[+] Convert Book to HTML")

	done := make(chan bool)
	go showTimeProgress("    Converting book", done)

	tempBookPath := TempDir + "/" + TempBookName
	runCommand(ebookConvertCmd, inputPath, (tempBookPath + ".htmlz"))
	runCommand(ebookConvertCmd, (tempBookPath + ".htmlz"), tempBookPath)

	done <- true
	time.Sleep(50 * time.Millisecond)

	if _, err := os.Stat(tempBookPath + "/index1.html"); err != nil {
		logFatalln("Please check if you have installed Calibre. Can you run the command 'ebook-convert' in your shell? I cannot access the 'ebook-convert' command in your system's shell. This script requires Calibre to process ebook texts.")
	}
}

func createBookWithWordwised(inputPath string) {
	extension := filepath.Ext(inputPath)
	bookPath := filepath.Dir(inputPath)
	fileName := strings.TrimSuffix(filepath.Base(inputPath), extension)

	// handle output format type
	if len(formatType) > 0 {
		extension = formatType
	} else {
		extension = strings.Trim(extension, ".")
	}

	log.Println("[+] Create New Book with Wordwised")

	done := make(chan bool)
	go showTimeProgress("      Creating book", done)

	htmlPath := fmt.Sprintf("%s/%s/index1.html", TempDir, TempBookName)
	outputPath := fmt.Sprintf("%s/%s-wordwise.%s", bookPath, fileName, extension)
	metaDataPath := fmt.Sprintf("%s/%s/content.opf", TempDir, TempBookName)
	runCommand(ebookConvertCmd, htmlPath, outputPath, "-m", metaDataPath)

	done <- true
	time.Sleep(50 * time.Millisecond)

	log.Println("[+] The EPUB book with wordwise was generated at", outputPath)
}

func getEbookConvertCmd() string {
	cmd_name := "ebook-convert"
	if !isCmdToolExists(cmd_name) {
		// try mac version
		mac_cmd := "/Applications/calibre.app/Contents/MacOS/ebook-convert"
		if isCmdToolExists(mac_cmd) {
			cmd_name = mac_cmd
		}
	}
	return cmd_name
}

func runCommand(name string, arg ...string) {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		log.Println("Run command:", name, strings.Join(arg, " "))
		log.Print(string(out))
		log.Println("Error:", err)
	}
}

func isCmdToolExists(tool_name string) bool {
	out, _ := exec.Command("command", "-v", tool_name).Output()
	res := string(out)
	return len(res) > 0
}

func cleanTempData() {
	// remove temp folder
	os.RemoveAll(TempDir)
}

func createTempFolder() {
	if err := os.Mkdir(TempDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
}

func showTimeProgress(description string, done chan bool) {
	bar := createProgressBar(100, description)
	count := 0
	fCount := 0.0
	decelerateValue := 1.0
	for {
		select {
		case <-done:
			bar.Set(100)
			return

		default:
			if count <= 90 {
				count++
				fCount = float64(count)
			} else {
				decelerateValue /= 1.12
				fCount = fCount + decelerateValue
				count = int(fCount)
			}
			bar.Set(count)
			time.Sleep(50 * time.Millisecond)
		}
	}
}
