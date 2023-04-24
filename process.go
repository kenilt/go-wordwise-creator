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

func processHtmlBookData(stopWords *map[string]bool, wordwiseDict *map[string]DictRow, lemmaDict *map[string]string) {
	htmlBookPath := fmt.Sprintf("%s/%s/index1.html", TempDir, TempBookName)

	bbytes, err := os.ReadFile(htmlBookPath)
	if err != nil {
		log.Fatalln("Error when open ", htmlBookPath, "->", err)
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
				processed, count, total := processBlock(collected, stopWords, wordwiseDict, lemmaDict)
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

func processBlock(content string, stopWords *map[string]bool, wordwiseDict *map[string]DictRow, lemmaDict *map[string]string) (string, int, int) {
	count := 0
	words := strings.Split(content, " ")
	for i := 0; i < len(words); i++ {
		word := cleanWord(words[i])
		if _, ok := (*stopWords)[word]; ok {
			continue
		}

		// first, find the word in dict
		ws, ok := (*wordwiseDict)[word]
		if !ok {
			// not found, find it normal form
			lm, ok := (*lemmaDict)[word]
			if !ok {
				continue
			}
			// then, find the normal form in dict
			ws, ok = (*wordwiseDict)[lm]
			if !ok {
				continue
			}
		}

		if ws.HintLvl > hintLevel {
			continue
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

		words[i] = fmt.Sprintf("<ruby>%v<rt>%v</rt></ruby>", words[i], meaning)
		count++
	}
	return strings.Join(words, " "), count, len(words)
}

// Remove special characters from word
func cleanWord(word string) string {
	return strings.ToLower(strings.Trim(word, ".?!,:;()[]{}<>“”‘’\"'`…*•&#~"))
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
