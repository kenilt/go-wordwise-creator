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
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	StopwordsPath          = "resources/stopwords.txt"
	WordwiseDictionaryPath = "resources/wordwise-dict.csv"
	TempDir                = "tempData"
	TempBookName           = "book_dump"
)

var specialChars = []string{",", "<", ">", ";", "&", "*", "~", "/", "\"", "[", "]", "#", "?", "`", "–", ".", "'", "!", "“", "”", ":", "."}
var hintLevel int = 5
var formatType string
var inputPath string
var ebookConvertCmd string

type ProcessState int

const (
	OpenTag ProcessState = iota
	Collecting
)

type DictRow struct {
	Word     string
	FullDef  string
	ShortDef string
	Example  string
	HintLv   int
}

func main() {
	readInputParams(os.Args)

	log.Println("[+] Hint level:", hintLevel)
	log.Println("[+] Format type:", formatType)

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
	log.Println("[+] Process book with wordwise")
	processHtmlBookData(stopWords, dict)

	// create wordwise book
	createBookWithWordwised(inputPath)

	// log.Println("[+] Cleaning temp files")
	cleanTempData()

	log.Println("--> Finished!")
}

func readInputParams(args []string) {
	if len(args) < 2 {
		log.Println("Usage: go run . input_file hint_level format_type")
		log.Println("input_file: A path to file need to generate wordwise")
		log.Println("hint_level: From 1 to 5, where 5 shows all wordwise hints, and 1 shows hints only for hard words with definitions. The default is 5")
		log.Println("format_type: The format type of output book, (ex: epub). The default is use the input format")
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

	if _, err := os.Stat(inputPath); err != nil {
		log.Fatalln(fmt.Sprintf("File at %s is not found!", inputPath))
	}
}

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
		log.Fatalln("Please check if you have installed Calibre. Can you run the command 'ebook-convert' in your shell? I cannot access the 'ebook-convert' command in your system's shell. This script requires Calibre to process ebook texts.")
	}
}

func processHtmlBookData(stopWords *map[string]bool, dict *map[string]DictRow) {
	htmlBookPath := fmt.Sprintf("%s/%s/index1.html", TempDir, TempBookName)

	bbytes, err := os.ReadFile(htmlBookPath)
	if err != nil {
		log.Fatalln("Error when open ", htmlBookPath, "->", err)
	}

	chars := []rune(string(bbytes))
	charLength := len(chars)
	var bookBuilder strings.Builder
	processCount := 0
	bar := createProgressBar(100, "    Processing book")

	state := Collecting
	var collectBuilder strings.Builder
	var tagBuilder strings.Builder
	isSawBody := false
	for i := 0; i < charLength; i++ {
		char := chars[i]
		if char == '<' { // see the open tag mean collecting finish, process what was collected
			state = OpenTag
			collected := collectBuilder.String()
			trimmed := strings.TrimSpace(collected)
			if isSawBody && len(trimmed) > 0 {
				processed, count := processBlock(trimmed, stopWords, dict)
				bookBuilder.WriteString(processed)
				processCount += count

				// update progress
				bar.Set(i * 100 / charLength)
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
	bar.Set(100)

	log.Println(fmt.Sprintf("--> Added wordwise for %d words", processCount))

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

func processBlock(content string, stopWords *map[string]bool, dict *map[string]DictRow) (string, int) {
	count := 0
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
		count++
	}
	return strings.Join(words, " "), count
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

func createProgressBar(max int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionSetWidth(15),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
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
