package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialogManager"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-the-king-says-bot/staticFunctions"
)

func GetTextInputProcessorManager() dialogManager.TextInputProcessorManager {
	return dialogManager.TextInputProcessorManager{
		Processors: dialogManager.TextProcessorsMap{
			"changeName":     processChangeName,
			"suggestCommand": processSuggestCommand,
		},
	}
}

func processChangeName(additionalId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)

	if len(data.Message) > 0 {
		if len(data.Message) < 30 {
			db.SetUserName(additionalId, data.Message)
			data.SendMessage(data.Trans("name_changed"), true)

			staticFunctions.FirstSetUpStep3(data)
			return true
		} else {
			data.SendMessage(data.Trans("name_too_long"), true)
		}
	} else {
		data.SendMessage(data.Trans("invalid_name"), true)
	}

	data.Static.SetUserStateTextProcessor(data.UserId, &processing.AwaitingTextProcessorData{
		ProcessorId: "changeName",
	})
	return true
}

func processSuggestCommand(additionalId int64, data *processing.ProcessData) bool {
	db := staticFunctions.GetDb(data.Static)
	db.AddSessionSuggestedCommand(additionalId, data.Message)
	staticFunctions.UpdateSessionDialogs(additionalId, data.Static)
	data.SendDialog(data.Static.MakeDialogFn("sc", data.UserId, data.Trans, data.Static, nil))
	return true
}
