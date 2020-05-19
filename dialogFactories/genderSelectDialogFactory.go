package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/gameraccoon/telegram-the-king-says-bot/staticFunctions"
	"strconv"
)

type genderSelectVariantPrototype struct {
	id string
	textId string
	process func(int64, *processing.ProcessData) bool
	rowId int
}

type genderSelectDialogFactory struct {
	variants []genderSelectVariantPrototype
}

func MakeGenderSelectDialogFactory() dialogFactory.DialogFactory {
	return &(genderSelectDialogFactory{
		variants: []genderSelectVariantPrototype{
			genderSelectVariantPrototype{
				id: "fe",
				textId: "gender_female",
				process: selectGenderFemale,
				rowId:1,
			},
			genderSelectVariantPrototype{
				id: "ma",
				textId: "gender_male",
				process: selectGenderMale,
				rowId:1,
			},
			genderSelectVariantPrototype{
				id: "bo",
				textId: "gender_both",
				process: selectGenderBoth,
				rowId:2,
			},
			genderSelectVariantPrototype{
				id: "no",
				textId: "gender_none",
				process: selectGenderNone,
				rowId:2,
			},
		},
	})
}

func selectGenderFemale(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetUserGender(data.UserId, 1)
	data.SubstitudeMessage(data.Trans("gender_changed"))
	return true
}

func selectGenderMale(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetUserGender(data.UserId, 2)
	data.SubstitudeMessage(data.Trans("gender_changed"))
	return true
}

func selectGenderBoth(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetUserGender(data.UserId, 3)
	data.SubstitudeMessage(data.Trans("gender_changed"))
	return true
}

func selectGenderNone(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetUserGender(data.UserId, 0)
	data.SubstitudeMessage(data.Trans("gender_changed"))
	return true
}

func (factory *genderSelectDialogFactory) createVariants(trans i18n.TranslateFunc) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		variants = append(variants, dialog.Variant{
			Id:   variant.id,
			Text: trans(variant.textId),
			RowId: variant.rowId,
		})
	}
	return
}

func (factory *genderSelectDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	return &dialog.Dialog{
		Text:     trans("select_language"),
		Variants: factory.createVariants(trans),
	}
}

func (factory *genderSelectDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	sessionId, _ := strconv.ParseInt(additionalId, 10, 64)
	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(sessionId, data)
		}
	}
	return false
}
