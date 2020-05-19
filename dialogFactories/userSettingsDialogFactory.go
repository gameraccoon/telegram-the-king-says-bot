package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/gameraccoon/telegram-the-king-says-bot/staticFunctions"
	static "github.com/gameraccoon/telegram-the-king-says-bot/staticData"
)

type userSettingsData struct {
	userId int64
	staticData *processing.StaticProccessStructs
}

type userSettingsVariantPrototype struct {
	id string
	textId string
	process func(int64, *processing.ProcessData) bool
	rowId int
	isActiveFn func(*userSettingsData) bool
}

type userSettingsDialogFactory struct {
	variants []userSettingsVariantPrototype
}

func MakeUserSettingsDialogFactory() dialogFactory.DialogFactory {
	return &(userSettingsDialogFactory{
		variants: []userSettingsVariantPrototype{
			userSettingsVariantPrototype{
				id: "name",
				textId: "change_name",
				process: changeName,
				rowId:1,
			},
			userSettingsVariantPrototype{
				id: "lang",
				textId: "change_language",
				process: changeLanguage,
				rowId:2,
			},
			userSettingsVariantPrototype{
				id: "gend",
				textId: "change_gender",
				process: changeGender,
				rowId:3,
			},
		},
	})
}

func changeName(userId int64, data *processing.ProcessData) bool {
	data.SubstitudeMessage(data.Trans("enter_name"))
	data.Static.SetUserStateTextProcessor(userId, &processing.AwaitingTextProcessorData{
		ProcessorId: "changeName",
		AdditionalId: userId,
	})
	return true
}

func changeLanguage(userId int64, data *processing.ProcessData) bool {
	data.SubstitudeDialog(data.Static.MakeDialogFn("lc", data.UserId, data.Trans, data.Static, nil))
	return true
}

func changeGender(userId int64, data *processing.ProcessData) bool {
	data.SubstitudeDialog(data.Static.MakeDialogFn("gc", data.UserId, data.Trans, data.Static, nil))
	return true
}

func (factory *userSettingsDialogFactory) createVariants(settingsData *userSettingsData, trans i18n.TranslateFunc) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		if variant.isActiveFn == nil || variant.isActiveFn(settingsData) {
			variants = append(variants, dialog.Variant{
				Id:   variant.id,
				Text: trans(variant.textId),
				RowId: variant.rowId,
			})
		}
	}
	return
}

func (factory *userSettingsDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	settingsData := userSettingsData {
		userId: userId,
		staticData: staticData,
	}

	db := staticFunctions.GetDb(staticData)

	language := db.GetUserLanguage(userId)

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		config = static.StaticConfiguration{}
	}

	langName := language

	for _, lang := range config.AvailableLanguages {
		if lang.Key == language {
			langName = lang.Name
			break
		}
	}

	translationMap := map[string]interface{} {
		"Name":   db.GetUserName(userId),
		"Lang":   langName,
		"Gender": staticFunctions.GetGenderNameFromId(db.GetUserGender(userId), trans),
	}

	return &dialog.Dialog{
		Text:     trans("user_settings_title", translationMap),
		Variants: factory.createVariants(&settingsData, trans),
	}
}

func (factory *userSettingsDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(data.UserId, data)
		}
	}
	return false
}
