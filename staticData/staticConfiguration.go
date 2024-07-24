package staticData

import (
	cedar "github.com/iohub/ahocorasick"
)

type LanguageData struct {
	Key  string
	Name string
}

type PlaceholderInfo struct {
	Values  []string
	Matcher *cedar.Matcher
}

type PlaceholderInfos struct {
	Male     PlaceholderInfo
	Female   PlaceholderInfo
	Common   PlaceholderInfo
	Opposite [2]PlaceholderInfo
}

type StaticConfiguration struct {
	AvailableLanguages []LanguageData
	DefaultLanguage    string
	ExtendedLog        bool
	Placeholders       PlaceholderInfos
	RunHttpServer      bool
	HttpServerPort     int
}

func compilePlaceholder(placeholder *PlaceholderInfo) {
	placeholder.Matcher = cedar.NewMatcher()
	for i, value := range placeholder.Values {
		placeholder.Matcher.Insert([]byte(value), i)
	}
	placeholder.Matcher.Compile()
}

func (placeholders *PlaceholderInfos) Compile() {
	compilePlaceholder(&placeholders.Female)
	compilePlaceholder(&placeholders.Male)
	compilePlaceholder(&placeholders.Common)
	compilePlaceholder(&placeholders.Opposite[0])
	compilePlaceholder(&placeholders.Opposite[1])
}
