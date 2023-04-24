# go-wordwise-creator

Generate wordwise for ebook formats (EPUB, MOBI, PRC, AZW3, PDF...)

This project is a golang port version of this one: https://github.com/xnohat/wordwisecreator  
And it based on this project: https://github.com/dungxmta/wordwisecreator  
And it also based on the data from this project https://github.com/michmech/lemmatization-lists  
And the phonemes data from this project https://github.com/open-dict-data/ipa-dict  
And use Google stranslate to prepare resource for vietnamese meaning

**WARNING: DO NOT, UNDER ANY CIRCUMSTANCES, DELETE THE FILES FOR THE SOURCE FORMAT. ALWAYS KEEP THE ORIGINAL FORMAT FOR YOUR BOOKS**

## Milestones
- [x] Port basic features from php project
- [x] Support config output type
- [x] Support stemming words
- [x] Support pronunciation symbols
- [x] Support Eng - Viet by input config
- [ ] Run app by double click on binary file
- [ ] Support the meaning of pharse
- [ ] Improve the performance by using multiple threads


## Usage
- You need to install [Calibre](https://calibre-ebook.com/)  
    + On MacOS, install Calibre at `/Applications`, if not please correct config in your path  
- Download the latest version from [latest release](https://github.com/kenilt/go-wordwise-creator/releases/latest)  
- Unzip the downloaded file  
- Run the command  
    + On Windows: `go-wordwise-creator.exe input_path hint_level format_type language`  
    + On MacOS: `./go-wordwise-creator input_path hint_level format_type language`

### Screenshots
![Apr-24-2023 17-26-39](https://user-images.githubusercontent.com/3811063/233970925-f4a4c8a0-4065-4ccb-a2e8-404bad01462c.gif)

Output example  
<img width="672" alt="output-example" src="https://user-images.githubusercontent.com/3811063/233971197-1afe2086-43d0-4d53-a325-8b9817250cd1.png">

## To run from source

### Installation
Before run this tool, you need to install these things:
- [Calibre](https://calibre-ebook.com/) 
    + On MacOS, install Calibre at `/Applications`, if not please correct config in your path  
- [Golang](https://go.dev/doc/install) with correct config (can run `go verion` in your terminal)

### Run from source

Update all the needed dependencies  
`go mod download`

To create book with wordwise:  
`go run . input_path hint_level format_type language`

```
Usage: go run . input_file hint_level format_type
- input_file: A path to file need to generate wordwise
- hint_level: From 1 to 5, where 5 shows all wordwise hints, and 1 shows hints only for hard words with definitions. The default is 5
- format_type: The format type of output book, (ex: "epub"). The default is use the input format. Note: the "mobi" format is not compatiable with this tool.
- language: The language output for wordwise meaning is only supported in "en" and "vi"

The output book will be exported at the same location with the input book with "-wordwise" suffix.
```

Example: `go run . Sample_book_test.epub`  
OR `go run . Sample_book_test.epub 3 azw3 en`  

### To prepare another optimized dictionary
Use Excel or Google Sheet to edit the `wordwise-dict.csv` file, use Google Translate to translate words to vietnamese, use `lemmatization-en.csv` file as the lemmatizer dictionary, and use the `phoneme-dict.csv` then use VLOOKUP function to get the phoneme of each word in the original file.

## Release step

- For windows  
`GOOS=windows GOARCH=amd64 go build`
- For MacOS  
`GOOS=darwin GOARCH=amd64 go build`  

Then zip the bin file with resources folder
