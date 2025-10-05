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
	Male      PlaceholderInfo
	Female    PlaceholderInfo
	Common    PlaceholderInfo
	Positions [10]PlaceholderInfo
	Opposite  [2]PlaceholderInfo
}

type StaticConfiguration struct {
	AvailableLanguages []LanguageData
	DefaultLanguage    string
	ExtendedLog        bool
	Placeholders       PlaceholderInfos
	RunHttpServer      bool
	HttpServerPort     int
	ShareWebAddress    string
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
	compilePlaceholder(&placeholders.Positions[0])
	compilePlaceholder(&placeholders.Positions[1])
	compilePlaceholder(&placeholders.Positions[2])
	compilePlaceholder(&placeholders.Positions[3])
	compilePlaceholder(&placeholders.Positions[4])
	compilePlaceholder(&placeholders.Positions[5])
	compilePlaceholder(&placeholders.Positions[6])
	compilePlaceholder(&placeholders.Positions[7])
	compilePlaceholder(&placeholders.Positions[8])
	compilePlaceholder(&placeholders.Positions[9])
}
