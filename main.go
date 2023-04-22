package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	StopwordsPath          = "resources/stopwords.txt"
	WordwiseDictionaryPath = "resources/wordwise-dict.csv"
	TempDir                = "tempData"
	TempBookName           = "book_dump"
)

var specialChars = []string{",", "<", ">", ";", "&", "*", "~", "/", "\"", "[", "]", "#", "?", "`", "–", ".", "'", "!", "“", "”", ":", "."}
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

	log.Println("[+] Load stopwords")
	stopWords := loadStopWords()

	log.Println("[+] Load wordwise dict")
	dict := loadWordwiseDict()

	// clean old temp
	log.Println("[+] Cleaning old temp files")
	cleanTempData()

	// get ebook convert cmd
	ebookConvertCmd = getEbookConvertCmd()

	// convert book to html
	createTempFolder()
	convertBookToHtml(inputPath)

	// process book
	processHtmlBookData(stopWords, dict)

	// create wordwise book
	createBookWithWordwised(inputPath)

	// log.Println("[+] Cleaning temp files")
	cleanTempData()

	log.Println("--> Finished!")
}

type ProcessState int

const (
	OpenTag ProcessState = iota
	Collecting
)

func processHtmlBookData(stopWords *map[string]bool, dict *map[string]DictRow) {
	htmlBookPath := fmt.Sprintf("%s/%s/index1.html", TempDir, TempBookName)

	bbytes, err := os.ReadFile(htmlBookPath)
	if err != nil {
		log.Fatalln("Error when open ", htmlBookPath, "->", err)
	}

	chars := []rune(string(bbytes))
	var bookBuilder strings.Builder

	state := Collecting
	var collectBuilder strings.Builder
	var tagBuilder strings.Builder
	isSawBody := false
	for i := 0; i < len(chars); i++ {
		char := chars[i]
		if char == '<' { // see the open tag mean collecting finish, process what was collected
			state = OpenTag
			collected := collectBuilder.String()
			trimmed := strings.TrimSpace(collected)
			if isSawBody && len(trimmed) > 0 {
				processed := processBlock(trimmed, stopWords, dict)
				bookBuilder.WriteString(processed)
			} else {
				bookBuilder.WriteString(collected)
			}
			collectBuilder.Reset()
			bookBuilder.WriteRune(char)
		} else if char == '>' { // see the close tag mean the tag content finish
			state = Collecting
			collectedTag := tagBuilder.String()
			if !isSawBody && strings.HasPrefix(collectedTag, "body") {
				isSawBody = true
			}
			tagBuilder.Reset()
			bookBuilder.WriteString(collectedTag)
			bookBuilder.WriteRune(char)
		} else {
			if state == Collecting {
				collectBuilder.WriteRune(char)
			} else if state == OpenTag {
				tagBuilder.WriteRune(char)
			}
		}
	}

	fo, err := os.Create(htmlBookPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fo.Close()

	_, err2 := fo.WriteString(bookBuilder.String())

	if err2 != nil {
		log.Fatal(err2)
	}
}

func processBlock(content string, stopWords *map[string]bool, dict *map[string]DictRow) string {
	words := strings.Fields(content)
	for i := 0; i < len(words); i++ {
		word := words[i]
		if _, ok := (*stopWords)[word]; ok {
			continue
		}

		ws, ok := (*dict)[word]
		if !ok {
			continue
		}

		if ws.HintLv > hintLevel {
			continue
		}

		words[i] = fmt.Sprintf("<ruby>%v<rt>%v</rt></ruby>", word, ws.ShortDef)
	}
	return strings.Join(words, " ")
}

// Remove special characters from word
func cleanWord(word string) string {
	replacer := strings.NewReplacer(specialChars...)
	cleanWord := strings.ToLower(replacer.Replace(word))

	return cleanWord
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

func convertBookToHtml(inputPath string) {
	log.Println("[+] Convert Book to HTML...")
	tempBookPath := TempDir + "/" + TempBookName
	runCommand(ebookConvertCmd, inputPath, (tempBookPath + ".htmlz"))
	runCommand(ebookConvertCmd, (tempBookPath + ".htmlz"), tempBookPath)

	if _, err := os.Stat(tempBookPath + "/index1.html"); err != nil {
		log.Fatalln("Please check if you have installed Calibre. Can you run the command 'ebook-convert' in your shell? I cannot access the 'ebook-convert' command in your system's shell. This script requires Calibre to process ebook texts.")
	}
}

func createBookWithWordwised(inputPath string) {
	extension := filepath.Ext(inputPath)
	bookPath := filepath.Dir(inputPath)
	fileName := strings.TrimSuffix(filepath.Base(inputPath), extension)
	extension = strings.Trim(extension, ".")

	log.Println("[+] Create New Book with Wordwised...")
	htmlPath := fmt.Sprintf("%s/%s/index1.html", TempDir, TempBookName)
	outputPath := fmt.Sprintf("%s/%s-wordwise.%s", bookPath, fileName, extension)
	metaDataPath := fmt.Sprintf("%s/%s/content.opf", TempDir, TempBookName)
	runCommand(ebookConvertCmd, htmlPath, outputPath, "-m", metaDataPath)

	log.Println("[+] The EPUB book with wordwise was generated at", outputPath)
}

func runCommand(name string, arg ...string) {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		log.Println("Run command:", name, strings.Join(arg, " "))
		log.Print(string(out))
		log.Println("Error:", err)
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

	log.Println("--> Csv words:", count)
	return &dict
}
