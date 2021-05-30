package utils

import (
	"regexp"
)

var matchRegex = []string{
	"\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*? \\(.*?\\))]\\s?(?P<name>.*?)\\s?\\((?P<theme>.*?)\\)\\s?\\[(?P<translator>.*?)]",
	"^\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*?)]\\s?(?P<name>.*?)\\s?\\[(?P<translator>.*?)]$",
	"^\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*?)]\\s?(?P<name>.*?)\\s?\\((?P<theme>.*?)\\)",
	"\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*?)]\\s?(?P<theme>.*?)$",
	"^[[［【](?P<translator>.*?)[]］】]\\s?\\[(?P<artist>.*?)]\\s?(?P<name>.*?)\\s?\\((?P<theme>.*?)\\)$",
}

type MatchTagResult struct {
	Name       string
	Artist     string
	Series     string
	Theme      string
	Translator string
}

func MatchName(name string) *MatchTagResult {
	for _, regexPattern := range matchRegex {
		rex := regexp.MustCompile(regexPattern)
		result := toMatchResult(rex, name)
		if result != nil {
			return result
		}
	}
	return nil
}

func toMatchResult(regex *regexp.Regexp, text string) *MatchTagResult {
	match := regex.FindStringSubmatch(text)
	if match == nil {
		return nil
	}
	result := &MatchTagResult{}
	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" {
			switch name {
			case "name":
				result.Name = match[i]
			case "artist":
				result.Artist = match[i]
			case "theme":
				result.Theme = match[i]
			case "series":
				result.Series = match[i]
			case "translator":
				result.Translator = match[i]
			}
		}
	}
	return result
}
