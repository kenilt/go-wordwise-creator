package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	StopwordsPath          = "resources/stopwords.txt"
	WordwiseDictionaryPath = "resources/wordwise-dict.csv"
	TempDir                = "tempData"
	IntermediateBookName   = "book_dump"
)

var hintLevel int = 5
var inputPath string
var ebookConvertCmd string

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Println("usage: go run . input_file hint_level")
		log.Println("input_file : path to file need to generate wordwise")
		log.Println("hint_level : from 1 to 5 default is 5, 1 is less wordwise hint show - only hard word will have definition, 5 is all wordwise hints show")
		os.Exit(0)
	}

	inputPath = args[1]
	if len(args) > 2 {
		parseNum, err := strconv.Atoi(args[2])
		if err == nil {
			hintLevel = parseNum
		}
	}
	log.Printf("[+] Hint level: %d\n", hintLevel)

	log.Println("[+] Load stopwords...")
	// stopWords := loadStopWords()
	loadStopWords()

	log.Println("[+] Load wordwise dict...")
	// dict := loadWordwiseDict()
	loadWordwiseDict()

	// clean old temp
	log.Println("[+] Cleaning old temp files...")
	cleanTempData()

	// get ebook convert cmd
	ebookConvertCmd = getEbookConvertCmd()

	// convert book to html
	convertBookToHtml(inputPath)

	// create wordwise book
	// createBookWithWordwised()

	// log.Println("[+] Cleaning temp files...")
	// cleanTempData()

	log.Println("--> Finished!")
}

func cleanTempData() {
	// remove htmlz file
	os.Remove(fmt.Sprintf("%s.htmlz", IntermediateBookName))
	// remove temp folder
	os.RemoveAll(IntermediateBookName)
}

func convertBookToHtml(inputPath string) {
	log.Println("[+] Convert Book to HTML")
	log.Println(ebookConvertCmd)
	exec.Command(ebookConvertCmd, inputPath, fmt.Sprintf("%s.htmlz", IntermediateBookName)).Run()
	exec.Command(ebookConvertCmd, fmt.Sprintf("%s.htmlz", IntermediateBookName), IntermediateBookName).Run()

	if _, err := os.Stat(fmt.Sprintf("%s/index1.html", IntermediateBookName)); err != nil {
		log.Fatalln("Please check did you installed Calibre ? Can you run command ebook-convert in shell ? I cannot access command ebook-convert in your system shell, This script need Calibre to process ebook texts")
	}
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

func isCmdToolExists(tool_name string) bool {
	out, _ := exec.Command("command", "-v", tool_name).Output()
	res := string(out)
	return len(res) > 0
}

// Load Stop Words from txt
func loadStopWords() *map[string]bool {
	dict := make(map[string]bool)

	file, err := os.Open(StopwordsPath)
	if err != nil {
		log.Fatalln("Error when open ", StopwordsPath, "->", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	// scanner.Split(bufio.ScanWords)

	count := 0
	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "#") {
			continue
		}

		if word != "" {
			dict[word] = true
			count++
			// log.Println(word)
		}
	}

	if scanner.Err() != nil {
		log.Fatalln("Error when scan word ", "->", err)
	}

	log.Println("--> Stop words:", count)
	return &dict
}

type DictRow struct {
	Word     string
	FullDef  string
	ShortDef string
	Example  string
	HintLv   int
}

// Load Dict from CSV
func loadWordwiseDict() *map[string]DictRow {

	file, err := os.Open(WordwiseDictionaryPath)
	if err != nil {
		log.Fatalln("Error when open ", StopwordsPath, "->", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	reader := csv.NewReader(file)

	dict := make(map[string]DictRow)
	row := DictRow{}

	// Read each record from csv
	// skip header
	record, err := reader.Read()
	if err == io.EOF {
		log.Fatalln("Empty csv file")
	}

	count := 0

	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("Error when scan word ", count, "->", err)
		}

		if len(record) < 6 {
			log.Println("Invalid word: ", record)
			continue
		}

		hintLv, err := strconv.Atoi(record[5])
		if err != nil {
			log.Println("Can't get hint_level: ", record, "->", err)
			continue
		}

		row = DictRow{
			Word:     record[1],
			FullDef:  record[2],
			ShortDef: record[3],
			Example:  record[4],
			HintLv:   hintLv,
		}

		dict[row.Word] = row
		count++
	}

	log.Println("--> csv words:", count)
	return &dict
}
