package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kenilt/go-wordwise-creator/filepicker"
)

var inputFormats = []string{"azw", "azw3", "azw4", "cbz", "cbr", "cb7", "cbc", "chm", "djvu", "docx", "epub", "fb2", "fbz", "html", "htmlz", "lit", "lrf", "mobi", "odt", "pdf", "prc", "pdb", "pml", "rb", "rtf", "snb", "tcr", "txt", "txtz"}
var outputFormats = []string{"azw3", "epub", "docx", "fb2", "htmlz", "oeb", "lit", "lrf", "mobi", "pdb", "pmlz", "rb", "pdf", "rtf", "snb", "tcr", "txt", "txtz", "zip"}
var hintLevel int = 5
var formatType string
var inputPath string
var wLang string = "en"
var isVietnamese bool = false
var isDoubleClick bool = false

func main() {
	readInputParams(os.Args)

	log.Println("[+] Input path:", inputPath)
	log.Println(fmt.Sprintf("[+] Hint level: %d, Output format type: %s, Language: %s", hintLevel, formatType, wLang))

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
	processHtmlBookData(wordwiseDict, lemmaDict)
	modifyCalibreTitle()

	// create wordwise book
	createBookWithWordwised(inputPath)

	cleanTempData()
	if isDoubleClick {
		pauseConsole()
	}
}

func readInputParams(args []string) {
	if len(args) < 2 {
		isDoubleClick = true
		readInputFromConsole()
	} else {
		assignInputPath(args[1])

		if len(args) > 2 {
			assignHintLevel(args[2])
		}

		if len(args) > 3 {
			assignOutputFormat(args[3])
		} else {
			assignOutputFormat("")
		}

		if len(args) > 4 {
			assignLanguage(args[4])
		}
	}
}

func readInputFromConsole() {
	checkThenChangeWorkingDir()

	userInput := bufio.NewReader(os.Stdin)
	log.Println("Choose the book's path:")
	fp := filepicker.New()
	fp.Path, _ = os.UserHomeDir()
	fp.CurrentDirectory, _ = os.Getwd()
	fp.AllowedTypes = inputFormats
	fmodel := model{filepicker: fp}
	teaModel, _ := tea.NewProgram(&fmodel, tea.WithOutput(os.Stderr)).StartReturningModel()
	mm := teaModel.(model)
	assignInputPath(mm.selectedFile)
	log.Println("Selected file:", mm.selectedFile)

	log.Println("Enter hint level (1-5): ")
	fmt.Print("                    ")
	scanValue, _ := userInput.ReadString('\n')
	scanValue = strings.TrimSpace(scanValue)
	assignHintLevel(scanValue)

	log.Println("Enter output format (not support \"mobi\"): ")
	fmt.Print("                    ")
	scanValue, _ = userInput.ReadString('\n')
	scanValue = strings.TrimSuffix(scanValue, "\n")
	assignOutputFormat(scanValue)

	log.Println("Enter language (\"en\" or \"vi\" only): ")
	fmt.Print("                    ")
	scanValue, _ = userInput.ReadString('\n')
	scanValue = strings.TrimSuffix(scanValue, "\n")
	assignLanguage(scanValue)
}

func assignInputPath(scanValue string) {
	inputPath = scanValue
	if _, err := os.Stat(inputPath); err != nil {
		logFatalln(fmt.Sprintf("File at %s is not found!", inputPath))
	}
}

func assignHintLevel(scanValue string) {
	parseNum, err := strconv.Atoi(scanValue)
	if err == nil {
		hintLevel = parseNum
	}
}

func assignOutputFormat(scanValue string) {
	if contains(outputFormats, scanValue) {
		formatType = scanValue
	} else {
		extension := strings.Trim(filepath.Ext(inputPath), ".")
		if contains(outputFormats, extension) {
			formatType = extension
		} else {
			formatType = "epub"
		}
	}
}

func assignLanguage(scanValue string) {
	wLang = scanValue
	if wLang == "vi" {
		isVietnamese = true
	} else {
		wLang = "en"
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func logFatalln(v ...any) {
	log.Println(v...)
	if isDoubleClick {
		pauseConsole()
	}
	os.Exit(1)
}

func pauseConsole() {
	log.Println("Press the Enter Key to exit!")
	fmt.Scanln()
}

func checkThenChangeWorkingDir() {
	isFoundResources := true
	if _, err := os.Stat(LemmaDictionaryPath); err != nil {
		isFoundResources = false
	}
	if _, err := os.Stat(WordwiseDictionaryPath); err != nil {
		isFoundResources = false
	}
	if !isFoundResources {
		execPath, _ := os.Executable()
		workingDir := filepath.Dir(execPath)
		fmt.Println("workingDir", workingDir)
		os.Chdir(workingDir)
	}
}
