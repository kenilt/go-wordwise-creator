package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

var hintLevel int = 5
var formatType string
var inputPath string
var isVietnamese bool = false

func main() {
	readInputParams(os.Args)

	log.Println("[+] Hint level:", hintLevel)
	log.Println("[+] Format type:", formatType)

	log.Println("[+] Load stopwords")
	stopWords := loadStopWords()

	log.Println("[+] Load wordwise dict")
	wordwiseDict := loadWordwiseDict()

	log.Println("[+] Load lemma dict")
	lemmaDict := loadLemmatizerDict()

	// clean old temp
	log.Println("[+] Cleaning old temp files")
	cleanTempData()

	// get ebook convert cmd
	ebookConvertCmd = getEbookConvertCmd()

	// convert book to html
	createTempFolder()
	convertBookToHtml(inputPath)

	// process book
	log.Println("[+] Process book with wordwise")
	processHtmlBookData(stopWords, wordwiseDict, lemmaDict)

	// create wordwise book
	createBookWithWordwised(inputPath)

	// log.Println("[+] Cleaning temp files")
	cleanTempData()

	log.Println("--> Finished!")
}

func readInputParams(args []string) {
	if len(args) < 2 {
		log.Println(`
Usage: go run . input_file hint_level format_type
- input_file: A path to file need to generate wordwise
- hint_level: From 1 to 5, where 5 shows all wordwise hints, and 1 shows hints only for hard words with definitions. The default is 5
- format_type: The format type of output book, (ex: "epub"). The default is use the input format. Note: the "mobi" format is not compatiable with this tool.
- language: The language output for wordwise meaning is only supported in "en" and "vi"

The output book will be exported at the same location with the input book with "-wordwise" suffix.
		`)
		os.Exit(0)
	}

	inputPath = args[1]
	if len(args) > 2 {
		parseNum, err := strconv.Atoi(args[2])
		if err == nil {
			hintLevel = parseNum
		}
	}
	if len(args) > 3 {
		formatType = args[3]
	}
	if len(args) > 4 {
		wLang := args[4]
		if wLang == "vi" {
			isVietnamese = true
		}
	}

	if _, err := os.Stat(inputPath); err != nil {
		log.Fatalln(fmt.Sprintf("File at %s is not found!", inputPath))
	}
}
