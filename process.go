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

	processedContent := processHtmlData(string(bbytes))

	fo, err := os.Create(htmlBookPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fo.Close()

	_, err2 := fo.WriteString(processedContent)

	if err2 != nil {
		log.Fatal(err2)
	}
}

func processHtmlData(htmlContent string) string {
	chars := []rune(htmlContent)
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
	return bookBuilder.String()
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
				phrase, converted := getWordwisePhrase(chars, word, i)
				if len(converted) > 0 {
					resBuilder.WriteString(converted)
					i = i + len(phrase) - len(word) - 1 // -1 because it has i++ end of each loop
					count++
					total += len(strings.Fields(phrase)) - 1 // -1 because it will be increase by total++ each loop
				} else {
					resBuilder.WriteString(moddedWord)
					resBuilder.WriteRune(char)
				}
			}
			if len(word) > 0 {
				total++
			}
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
	if len(lastWord) > 0 {
		total++
	}

	return resBuilder.String(), total, count
}

// process based on the original chars from a position, word is the last word before from
// Its process do by combine "word" + some next word then find it in the dictionary
// Return the modded phrase, and len of original phrase
func getWordwisePhrase(chars []rune, word string, from int) (string, string) {
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
					return phrase, resWord
				}
			}
			sb.WriteRune(char)
		} else {
			sb.WriteRune(char)
		}
	}

	wordCount++
	if wordCount > 1 && wordCount <= 5 {
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
			return phrase, resWord
		}
	}

	return "", ""
}

func convertThePhrase() {

}

// originalWord is raw and contains functual marks, ex: "\"Whosever,.."
func getWordwiseWord(originalWord string) (string, bool) {
	ws := findWordwiseInDictionary(cleanWord(originalWord))
	if ws == nil {
		return originalWord, false
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

	trimmed := trimWord(originalWord)
	modded := fmt.Sprintf("<ruby>%v<rt>%v</rt></ruby>", trimmed, meaning)
	resWord := strings.Replace(originalWord, trimmed, modded, 1)
	return resWord, true
}

// word has to be cleanned and lowercased
func findWordwiseInDictionary(word string) *DictRow {
	// first, find the word in wordwise dictionary
	dicRow, isFound := (*wordwiseDict)[word]
	if !isFound {
		// not found, find its normal form
		normalForm, isFound := (*lemmaDict)[word]
		if !isFound {
			return nil
		}
		// then, find the normal form word in wordwise dictionary
		dicRow, isFound = (*wordwiseDict)[normalForm]
		if !isFound {
			return nil
		}
	}

	// skip the word if hint level is not pass
	if dicRow.HintLvl > hintLevel {
		return nil
	}

	return &dicRow
}

// Trim special characters from word, then lowercase
func cleanWord(word string) string {
	return strings.ToLower(trimWord(word))
}

// Trim special characters from word
func trimWord(word string) string {
	return strings.Trim(word, ".?!,:;()[]{}<>“”‘’\"'`…*•-&#~")
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
