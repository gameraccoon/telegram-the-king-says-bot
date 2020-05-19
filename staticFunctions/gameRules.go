package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"math/rand"
)

func SendMessageToOthers(data *processing.ProcessData, sessionId int64, command string) {

}

func GiveRandomNumbersToPlayers(data *processing.ProcessData, sessionId int64) {
	db := GetDb(data.Static)

	userIds := db.GetUsersInSession(sessionId)

	if len(userIds) < 2 {
		data.SendMessage(data.Trans("few_players"))
		return
	}

	rand.Shuffle(len(userIds), func(i, j int) { userIds[i], userIds[j] = userIds[j], userIds[i] })
	for i, userId := range userIds {
		trans := FindTransFunction(userId, data.Static)
		theme := trans("player_number_msg", map[string]interface{} {
			"Number": i+1,
		})

		chatId := db.GetChatId(userId)
		data.Static.Chat.SendMessage(chatId, theme, 0)
	}
	return
}


func GiveRandomNumbersToPlayersDividedByGender(data *processing.ProcessData, sessionId int64) {
}