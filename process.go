package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

type ProcessState int

const (
	OpenTag ProcessState = iota
	Collecting
)

func processHtmlBookData() {
	htmlBookPath := fmt.Sprintf("%s/%s/index1.html", TempDir, TempBookName)

	bbytes, err := os.ReadFile(htmlBookPath)
	if err != nil {
		logFatalln("Error when open ", htmlBookPath, "->", err)
	}

	chars := []rune(string(bbytes))
	charLength := len(chars)
	var bookBuilder strings.Builder
	wordwiseCount := 0
	totalCount := 0
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
				processed, total, count := processBlock(collected)
				bookBuilder.WriteString(processed)
				wordwiseCount += count
				totalCount += total

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

	log.Println(fmt.Sprintf("--> Processed %d words, Added wordwise for %d words", totalCount, wordwiseCount))

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

func processBlock(content string) (string, int, int) {
	chars := []rune(string(content))
	charLength := len(chars)
	total, count := 0, 0
	var resBuilder strings.Builder
	var wordBuilder strings.Builder

	for i := 0; i < charLength; i++ {
		char := chars[i]
		if char == ' ' || char == '–' || char == '—' || // space, en dash, em dash
			(char == '-' && i < charLength-1 && chars[i+1] == '-') { // double hyphens
			word := wordBuilder.String()
			wordBuilder.Reset()

			moddedWord, isProcess := getWordwiseWord(word)
			if isProcess {
				resBuilder.WriteString(moddedWord)
				resBuilder.WriteRune(char)
				count++
			} else {
				phrase, pLen := processPharse(chars, word, i)
				if pLen > 0 {
					resBuilder.WriteString(phrase)
					i = i + pLen - len(word) - 1
					count++
				} else {
					resBuilder.WriteString(moddedWord)
					resBuilder.WriteRune(char)
				}
			}
			total++
		} else {
			wordBuilder.WriteRune(char)
		}
	}

	lastWord := wordBuilder.String()
	moddedWord, isProcess := getWordwiseWord(lastWord)
	resBuilder.WriteString(moddedWord)
	if isProcess {
		count++
	}
	total++

	return resBuilder.String(), total, count
}

func processPharse(chars []rune, word string, from int) (string, int) {
	var sb strings.Builder
	sb.WriteString(word)
	wordCount := 0
	for i := from; i < len(chars); i++ {
		char := chars[i]
		if char == ' ' {
			wordCount++
			if wordCount > 5 {
				break
			}

			if wordCount > 1 {
				phrase := sb.String()
				ws := findWordwiseInDictionary(cleanWord(phrase))
				if ws != nil {
					var meaning string
					if isVietnamese {
						if len(ws.Phoneme) > 0 {
							meaning = ws.Phoneme + " " + ws.Vi
						} else {
							meaning = ws.Vi
						}
					} else {
						meaning = ws.En
					}
					trimmed := trimWord(phrase)
					modded := fmt.Sprintf("<ruby>%v<rt>%v</rt></ruby>", trimmed, meaning)
					resWord := strings.Replace(phrase, trimmed, modded, 1)
					return resWord, len(phrase)
				}
			}
			sb.WriteRune(char)
		} else {
			sb.WriteRune(char)
		}
	}
	return "", 0
}

func getWordwiseWord(orgWord string) (string, bool) {
	ws := findWordwiseInDictionary(cleanWord(orgWord))
	if ws == nil {
		return orgWord, false
	}

	var meaning string
	if isVietnamese {
		if len(ws.Phoneme) > 0 {
			meaning = ws.Phoneme + " " + ws.Vi
		} else {
			meaning = ws.Vi
		}
	} else {
		meaning = ws.En
	}

	trimmed := trimWord(orgWord)
	modded := fmt.Sprintf("<ruby>%v<rt>%v</rt></ruby>", trimmed, meaning)
	resWord := strings.Replace(orgWord, trimmed, modded, 1)
	return resWord, true
}

func findWordwiseInDictionary(word string) *DictRow {
	// first, find the word in dict
	ws, isFound := (*wordwiseDict)[word]
	if !isFound {
		// not found, find its normal form
		lm, isFound := (*lemmaDict)[word]
		if !isFound {
			return nil
		}
		// then, find the normal form in dict
		ws, isFound = (*wordwiseDict)[lm]
		if !isFound {
			return nil
		}
	}

	if ws.HintLvl > hintLevel {
		return nil
	}

	return &ws
}

// Remove special characters from word
func cleanWord(word string) string {
	return strings.ToLower(trimWord(word))
}

func trimWord(word string) string {
	return strings.Trim(word, ".?!,:;()[]{}<>“”‘’\"'`…*•&#~")
}

func modifyCalibreTitle() {
	metaPath := fmt.Sprintf("%s/%s/content.opf", TempDir, TempBookName)

	bbytes, err := os.ReadFile(metaPath)
	if err != nil {
		return
	}

	chars := []rune(string(bbytes))
	charLength := len(chars)
	var metaBuilder strings.Builder

	state := Collecting
	isTitle := false
	var collectBuilder strings.Builder
	var tagBuilder strings.Builder
	for i := 0; i < charLength; i++ {
		char := chars[i]
		if char == '<' { // see the open tag mean collecting finish, process what was collected
			state = OpenTag
			collected := collectBuilder.String()
			collectBuilder.Reset()
			if isTitle {
				collected += " - Wordwise"
			}
			metaBuilder.WriteString(collected)
			metaBuilder.WriteRune(char)
		} else if char == '>' { // see the close tag mean the tag content finish
			state = Collecting
			collectedTag := tagBuilder.String()
			isTitle = (collectedTag == "dc:title")
			tagBuilder.Reset()
			metaBuilder.WriteString(collectedTag)
			metaBuilder.WriteRune(char)
		} else {
			if state == Collecting {
				collectBuilder.WriteRune(char)
			} else if state == OpenTag {
				tagBuilder.WriteRune(char)
			}
		}
	}

	fo, err := os.Create(metaPath)
	if err != nil {
		return
	}
	defer fo.Close()

	fo.WriteString(metaBuilder.String())
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
