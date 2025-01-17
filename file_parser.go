package main

import (
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/aymerick/douceur/parser"
	"github.com/mazznoer/csscolorparser"
)

// SectionsParse returns LongScripts and ColorScripts
func SectionsParse(filePath string) ([]Script, []Script) {
	dat, err := os.ReadFile(filePath)
	if err != nil {
		return []Script{}, []Script{}
	}
	stylesheet, err := parser.Parse(string(dat))
	if err != nil {
		return []Script{}, []Script{}
	}
	styleString := strings.Replace(stylesheet.String(), "\r", "", -1)

	styleSection := StyleSection{name: "", value: []string{}, filePath: ""}
	longScriptList := []Script{}
	colorScriptList := []Script{}
	colorReg := regexp.MustCompile(`#[A-Fa-f0-9]{3,6}|rgba\([0-9,%/ ]*\)|hsla\([0-9,%/ ]*\)|rgb\([0-9,%/ ]*\)|hsl\([0-9,%/ ]*\)`)
	for _, sub := range strings.Split(styleString, "\n") {
		if strings.HasSuffix(sub, "{") {
			styleSection.name = strings.Replace(sub, "{", "", -1)
			styleSection.filePath = filePath
		} else if strings.Contains(sub, "}") {
			if len(styleSection.value) > 0 {
				sort.Strings(styleSection.value)
				styleSection.valueHash = hash(strings.Join(styleSection.value, ""))
				styleList = append(styleList, styleSection)
			}
			// Generate hashes for each line in class, for similarity compare
			if len(styleSection.value) >= 5 {
				for _, value := range styleSection.value {
					hashValue := hash(value)
					if counter, found := hashCounters[hashValue]; found {
						hashCounters[hashValue] = append(counter, StyleHashRecorder{sectionIndex: len(styleList) - 1, originString: value})
					} else {
						hashCounters[hashValue] = []StyleHashRecorder{{sectionIndex: len(styleList) - 1, originString: value}}
					}
				}
			}
			styleSection = StyleSection{name: "", value: []string{}, filePath: ""}
		} else {
			partials := strings.Split(sub, ": ")
			if len(partials) == 2 {
				key := strings.TrimSpace(partials[0])
				value := strings.TrimSpace(partials[1])
				// Check is Long Script
				if len(value) > params.longScriptLength && !strings.Contains(value, "var") {
					longScriptList = append(longScriptList, Script{filePath: filePath,
						sectionName: styleSection.name,
						hashValue:   hash(value),
						value:       value,
						key:         key,
					})
				}
				// Colors Checking
				colors := ColorsProcessor(value, colorReg)
				for _, color := range colors {
					colorScriptList = append(colorScriptList, Script{
						filePath:    filePath,
						sectionName: styleSection.name,
						hashValue:   hash(color.rgb),
						value:       color.original,
						key:         key,
					})
				}
			}
			if len(strings.TrimSpace(sub)) > 0 {
				styleSection.value = append(styleSection.value, strings.TrimSpace(sub))
			}
		}
	}
	return longScriptList, colorScriptList
}

type ColorConversionStruct struct {
	original string
	rgb      string
}

// ColorsProcessor receive string with 0, 1 or more hex, rgb, rgba, hsl, hsla colors and returns their rgb string values
func ColorsProcessor(value string, colorReg *regexp.Regexp) []ColorConversionStruct {
	colors := []ColorConversionStruct{}
	matchList := colorReg.FindAllStringSubmatch(strings.ToLower(value), -1)
	for _, match := range matchList {
		if color, err := csscolorparser.Parse(match[0]); err == nil {
			colors = append(colors, ColorConversionStruct{original: match[0], rgb: color.RGBString()})
		}
	}
	return colors
}
