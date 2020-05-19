package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"math/rand"
)

func getUsersInSessionOfReportFailure(data *processing.ProcessData, sessionId int64) (userIds []int64, success bool) {
	db := GetDb(data.Static)

	userIds = db.GetUsersInSession(sessionId)
	success = true

	if len(userIds) < 2 {
		data.SendMessage(data.Trans("few_players"))
		success = false
	}

	return
}

func SendMessageToOthers(data *processing.ProcessData, sessionId int64, command string) {

}

func GiveRandomNumbersToPlayers(data *processing.ProcessData, sessionId int64) {
	db := GetDb(data.Static)

	userIds, success := getUsersInSessionOfReportFailure(data, sessionId)

	if !success {
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
	db := GetDb(data.Static)

	userIds, success := getUsersInSessionOfReportFailure(data, sessionId)

	if !success {
		return
	}

	rand.Shuffle(len(userIds), func(i, j int) { userIds[i], userIds[j] = userIds[j], userIds[i] })

	maleIdx := 0
	femaleIdx := 0

	for _, userId := range userIds {
		gender := db.GetUserGender(userId)

		if gender == 0 {
			continue
		}

		trans := FindTransFunction(userId, data.Static)

		var message string

		if gender & 1 != 0 {
			message = trans("female_number", map[string]interface{} {
				"Number": femaleIdx+1,
			})
			femaleIdx++
		}

		if gender & 2 != 0 {
			if len(message) > 0 {
				message += "\n"
			}
			message += trans("male_number", map[string]interface{} {
				"Number": maleIdx+1,
			})
			maleIdx++
		}

		chatId := db.GetChatId(userId)
		data.Static.Chat.SendMessage(chatId, message, 0)
	}
	return
}