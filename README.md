# go-wordwise-creator

Generate wordwise for ebook formats (EPUB, MOBI, PRC, AZW3, PDF...)

This project is a golang port version of this one: https://github.com/xnohat/wordwisecreator
And it based on this project: https://github.com/dungxmta/wordwisecreator

## Milestones
- [ ] Port basic feature from php project - WIP
- [ ] Support input flags and config output type
- [ ] Support the meaning of pharse
- [ ] Support stemmer words
- [ ] Support Eng - Viet by input config
- [ ] Improve the performance by using multiple threads

## Installation
Before using this tool, you need to install these things:
- [Calibre](https://calibre-ebook.com/) 
- [Golang](https://go.dev/doc/install) with correct config (can run `go verion` in your terminal)


## Usage
You need to have [calibre](https://calibre-ebook.com/) on your device.

Open your terminal then type:
`go run . input_file hint_level`
- input_file : path to file need to generate wordwise.
- hint_level : from 1 to 5 default is 5, 1 is less wordwise hint show - only hard word will have definition, 5 is all wordwise hints show.
- The output book will be exported at the same location with the input book with `-wordwise` suffix.
