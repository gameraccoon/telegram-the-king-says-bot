package dialogFactories

import (
	"strconv"

	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-the-king-says-bot/staticFunctions"
	"github.com/nicksnyder/go-i18n/i18n"
)

type roleSelectVariantPrototype struct {
	id      string
	textId  string
	process func(int64, *processing.ProcessData) bool
	rowId   int
}

type roleSelectDialogFactory struct {
	variants []roleSelectVariantPrototype
}

func MakeRoleSelectDialogFactory() dialogFactory.DialogFactory {
	return &(roleSelectDialogFactory{
		variants: []roleSelectVariantPrototype{
			roleSelectVariantPrototype{
				id:      "fe",
				textId:  "role_female",
				process: selectRoleFemale,
				rowId:   1,
			},
			roleSelectVariantPrototype{
				id:      "ma",
				textId:  "role_male",
				process: selectRoleMale,
				rowId:   1,
			},
			roleSelectVariantPrototype{
				id:      "bo",
				textId:  "role_both",
				process: selectRoleBoth,
				rowId:   2,
			},
			roleSelectVariantPrototype{
				id:      "no",
				textId:  "role_none",
				process: selectRoleNone,
				rowId:   2,
			},
		},
	})
}

func selectRoleFemale(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetUserRole(data.UserId, 1)
	data.SubstituteMessage(data.Trans("role_changed"))
	return true
}

func selectRoleMale(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetUserRole(data.UserId, 2)
	data.SubstituteMessage(data.Trans("role_changed"))
	return true
}

func selectRoleBoth(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetUserRole(data.UserId, 3)
	data.SubstituteMessage(data.Trans("role_changed"))
	return true
}

func selectRoleNone(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.SetUserRole(data.UserId, 0)
	data.SubstituteMessage(data.Trans("role_changed"))
	return true
}

func (factory *roleSelectDialogFactory) createVariants(trans i18n.TranslateFunc) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		variants = append(variants, dialog.Variant{
			Id:    variant.id,
			Text:  trans(variant.textId),
			RowId: variant.rowId,
		})
	}
	return
}

func (factory *roleSelectDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	return &dialog.Dialog{
		Text:     trans("select_role"),
		Variants: factory.createVariants(trans),
	}
}

func (factory *roleSelectDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	sessionId, _ := strconv.ParseInt(additionalId, 10, 64)
	for _, variant := range factory.variants {
		if variant.id == variantId {
			isSucceeded := variant.process(sessionId, data)
			staticFunctions.FirstSetUpStep4(data)
			return isSucceeded
		}
	}
	return false
}
