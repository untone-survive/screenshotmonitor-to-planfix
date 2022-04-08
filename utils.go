package main

import (
	"strings"
)

var badCharacters = []string{
	"../",
	"./",
	"<!--",
	"-->",
	"<",
	">",
	"'",
	"\"",
	"&",
	"$",
	"#",
	"{", "}", "[", "]", "=",
	";", "?", "%20", "%22",
	"%3c",   // <
	"%253c", // <
	"%3e",   // >
	"%0e",   // >
	"%28",   // (
	"%29",   // )
	"%2528", // (
	"%26",   // &
	"%24",   // $
	"%3f",   // ?
	"%3b",   // ;
	"%3d",   // =
}

func RemoveBadCharacters(input string, dictionary []string) string {

	temp := input

	for _, badChar := range dictionary {
		temp = strings.Replace(temp, badChar, "", -1)
	}
	return temp
}

func SanitizeFilename(name string) string {

	// default settings
	var badDictionary []string = badCharacters

	if name == "" {
		return name
	}

	// trim(remove)white space
	trimmed := strings.TrimSpace(name)
	// remove bad characters from filename
	trimmed = RemoveBadCharacters(trimmed, badDictionary)

	stripped := strings.Replace(trimmed, ":", "-", -1)
	stripped = strings.Replace(stripped, "\\", "", -1)

	return stripped
}
