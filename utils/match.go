package utils

import (
	"regexp"
)

var matchRegex = []string{
	"\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*? \\(.*?\\))]\\s?(?P<name>.*?)\\s?\\((?P<theme>.*?)\\)\\s?\\[(?P<translator>.*?)]",
	"^\\[(?P<translator>.*?)]\\s?\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*?)]\\s?(?P<name>.*?)\\s?\\((?P<theme>.*?)\\)$",
	"^\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*?)]\\s?(?P<name>.*?)\\s?\\((?P<theme>.*?)\\)\\s?\\[(?P<translator>.*?)]$",
	"^\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*?)]\\s?(?P<name>.*?)\\s?\\[(?P<translator>.*?)]$",
	"^\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*?)]\\s?(?P<name>.*?)\\s?\\((?P<theme>.*?)\\)$",
	"^\\[(?P<translator>.*?)]\\s?\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*?)]\\s?(?P<name>.*?)$",
	"\\((?P<series>.*?)\\)\\s?\\[(?P<artist>.*?)]\\s?(?P<theme>.*?)$",
	"^[[［【](?P<translator>.*?)[]］】]\\s?\\[(?P<artist>.*?)]\\s?(?P<name>.*?)\\s?\\((?P<theme>.*?)\\)$",
}
var tagMatchRegex = []string{
	"\\((.*?)\\)",
	"\\[(.*?)]",
	"{(.*?)}",
	"《(.*?)》",
	"（(.*?)）",
	"【(.*?)】",
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

func MatchTagTextFromText(text string) []string {
	result := make([]string, 0)
	for _, regexPattern := range tagMatchRegex {
		regex := regexp.MustCompile(regexPattern)
		groups := regex.FindAllStringSubmatch(text, -1)
		for _, group := range groups {
			if len(group) > 1 {
				result = append(result, group[1:]...)
			}
		}
	}
	return result
}
