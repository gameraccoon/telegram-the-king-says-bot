package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/gameraccoon/telegram-the-king-says-bot/staticFunctions"
	"strconv"
)

type suggestedConfirmedVariantPrototype struct {
	id string
	textId string
	process func(int64, *processing.ProcessData) bool
	rowId int
	isActiveFn func() bool
}

type suggestedConfirmedDialogFactory struct {
	variants []suggestedConfirmedVariantPrototype
}

func MakeSuggestedConfirmedDialogFactory() dialogFactory.DialogFactory {
	return &(suggestedConfirmedDialogFactory{
		variants: []suggestedConfirmedVariantPrototype{
			suggestedConfirmedVariantPrototype{
				id: "discsess",
				textId: "suggest_another",
				process: suggestAnother,
				rowId:1,
			},
		},
	})
}

func suggestAnother(sessionId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	currentSessionId, isInSession := db.GetUserSession(data.UserId)

	if !isInSession || sessionId != currentSessionId {
		data.SendMessage(data.Trans("session_is_too_old"))
		return true
	}

	data.SendMessage(data.Trans("suggest_command_msg"))
	data.Static.SetUserStateTextProcessor(data.UserId, &processing.AwaitingTextProcessorData{
		ProcessorId: "suggestCommand",
		AdditionalId: sessionId,
	})
	return true
}

func (factory *suggestedConfirmedDialogFactory) createVariants(trans i18n.TranslateFunc, sessionId int64) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		if variant.isActiveFn == nil || variant.isActiveFn() {
			variants = append(variants, dialog.Variant{
				Id:   variant.id,
				Text: trans(variant.textId),
				RowId: variant.rowId,
				AdditionalId: strconv.FormatInt(sessionId, 10),
			})
		}
	}
	return
}

func (factory *suggestedConfirmedDialogFactory) MakeDialog(userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
	db := staticFunctions.GetDb(staticData)

	sessionId, _ := db.GetUserSession(userId)

	return &dialog.Dialog{
		Text:     trans("suggested_command_sent"),
		Variants: factory.createVariants(trans, sessionId),
	}
}

func (factory *suggestedConfirmedDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	sessionId, _ := strconv.ParseInt(additionalId, 10, 64)
	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(sessionId, data)
		}
	}
	return false
}
