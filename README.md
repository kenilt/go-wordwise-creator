# go-wordwise-creator

Generate wordwise for ebook formats (EPUB, MOBI, PRC, AZW3, PDF...)

This project is a Golang port version of this one: https://github.com/xnohat/wordwisecreator  
And it is based on this project: https://github.com/dungxmta/wordwisecreator  
And it uses data from https://github.com/michmech/lemmatization-lists  
And the phonemes data from https://github.com/open-dict-data/ipa-dict, https://github.com/cmusphinx/cmudict, and https://github.com/menelik3/cmudict-ipa  
And use Google Translate to prepare resources for Vietnamese meaning

**WARNING: DO NOT, UNDER ANY CIRCUMSTANCES, DELETE THE FILES FOR THE SOURCE FORMAT. ALWAYS KEEP THE ORIGINAL FORMAT FOR YOUR BOOKS**

## Features
- [x] Support multiple book formats
- [x] Support stemming words
- [x] Support pronunciation symbols
- [x] Support meaning in English, and Vietnamese
- [x] Support the meaning of phrases
- [x] The speed is impressive, almost books were processed in seconds


## Usage
- You need to install [Calibre](https://calibre-ebook.com/)  
    + On MacOS, install Calibre at `/Applications`, if not please correct the config in your path  
- Download the latest version from [latest release](https://github.com/kenilt/go-wordwise-creator/releases/latest), unzip then double click on the `go-wordwise-creator` binary file to run.  
- OR you can run by the command  
    + On Windows: `go-wordwise-creator.exe input_path hint_level format_type language`  
    + On MacOS: `./go-wordwise-creator input_path hint_level format_type language`

### Screenshots
![Apr-24-2023 17-26-39](https://user-images.githubusercontent.com/3811063/233970925-f4a4c8a0-4065-4ccb-a2e8-404bad01462c.gif)

Output example  
<img width="672" alt="output-example" src="https://user-images.githubusercontent.com/3811063/233971197-1afe2086-43d0-4d53-a325-8b9817250cd1.png">

## To run from the source

### Installation
Before running this tool, you need to install these things:
- [Calibre](https://calibre-ebook.com/) 
    + On MacOS, install Calibre at `/Applications`, if not please correct the config in your path  
- [Golang](https://go.dev/doc/install) with correct config (can run `go version` in your terminal)

### Run from source

Update all the needed dependencies  
`go mod download`

To create a wordwise book:   
`go run . input_path hint_level format_type language`

```
Usage: go run . input_file hint_level format_type
- input_file: A path to the file needs to generate wordwise
- hint_level: From 1 to 5, where 5 shows all wordwise hints and 1 shows hints only for hard words with definitions. The default is 5
- format_type: The format type of the output book, (ex: "epub"). The default is to use the input format. Note: the "mobi" format is not compatible with this tool.
- language: The language output for wordwise meaning is only supported in "en" and "vi"

The output book will be exported at the same location as the input book with the "-wordwise" suffix.
```

Example: `go run . Sample_book_test.epub`  
OR `go run . Sample_book_test.epub 3 azw3 en`  

### To prepare another optimized dictionary
Use Excel or Google Sheets to edit the `wordwise-dict.csv` file, use Google Translate to translate words to Vietnamese, use the `lemmatization-en.csv` file as the lemmatizer dictionary, and use the `phoneme-dict.csv` then use VLOOKUP function to get the phoneme of each word in the original file.

## Release step

- For Windows  
`GOOS=windows GOARCH=amd64 go build`
- For MacOS  
`GOOS=darwin GOARCH=amd64 go build`  

Then zip the bin file with the resources folder

### To auto-build for MacOS and Windows

Give permission for sh script  
`chmod +x autobuild.sh`  
Run the autobuild  
`./autobuild.sh`  
