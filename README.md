# go-wordwise-creator

Generate wordwise for ebook formats (EPUB, MOBI, PRC, AZW3, PDF...)

This project is a golang port version of this one: https://github.com/xnohat/wordwisecreator
And it based on this project: https://github.com/dungxmta/wordwisecreator

## Milestones
- [x] Port basic features from php project
- [x] Support config output type
- [ ] Support the meaning of pharse
- [ ] Support stemmer words
- [ ] Support Eng - Viet by input config
- [ ] Improve the performance by using multiple threads

## Installation
Before using this tool, you need to install these things:
- [Calibre](https://calibre-ebook.com/) 
- [Golang](https://go.dev/doc/install) with correct config (can run `go verion` in your terminal)


## Usage
Update all the needed dependencies  
`go mod download`

To create book with wordwise:  
`go run . input_path hint_level format_type`

```
Usage: go run . input_file hint_level format_type
input_file: A path to file need to generate wordwise
hint_level: From 1 to 5, where 5 shows all wordwise hints, and 1 shows hints only for hard words with definitions. The default is 5
format_type: The format type of output book, (ex: epub). The default is use the input format

The output book will be exported at the same location with the input book with `-wordwise` suffix.
```

Example: `go run . Sample_book_test.epub 3 mobi`