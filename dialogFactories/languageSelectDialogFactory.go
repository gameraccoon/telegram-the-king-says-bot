package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	static "github.com/gameraccoon/telegram-the-king-says-bot/staticData"
	"github.com/gameraccoon/telegram-the-king-says-bot/staticFunctions"
	"github.com/nicksnyder/go-i18n/i18n"
)

type languageSelectVariantPrototype struct {
	id      string
	text    string
	process func(*processing.ProcessData) bool
}

type languageSelectDialogFactory struct {
}

func MakeLanguageSelectDialogFactory() dialogFactory.DialogFactory {
	return &(languageSelectDialogFactory{})
}

func applyNewLanguage(data *processing.ProcessData, newLang string) bool {
	staticFunctions.GetDb(data.Static).SetUserLanguage(data.UserId, newLang)
	data.Trans = staticFunctions.FindTransFunction(data.UserId, data.Static)
	data.SubstituteMessage(data.Trans("language_changed"))
	return true
}

func (factory *languageSelectDialogFactory) createVariants(staticData *processing.StaticProccessStructs) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	itemId := 0
	itemsInRow := 2

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		config = static.StaticConfiguration{}
	}

	for _, lang := range config.AvailableLanguages {
		variants = append(variants, dialog.Variant{
			Id:    lang.Key,
			Text:  lang.Name,
			RowId: itemId/itemsInRow + 1,
		})
		itemId = itemId + 1
	}
	return
}

func (factory *languageSelectDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	return &dialog.Dialog{
		Text:     trans("select_language"),
		Variants: factory.createVariants(staticData),
	}
}

func (factory *languageSelectDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	isSucceeded := applyNewLanguage(data, variantId)
	staticFunctions.FirstSetUpStep2(data)
	return isSucceeded
}
