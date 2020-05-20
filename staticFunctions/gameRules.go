package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	static "github.com/gameraccoon/telegram-the-king-says-bot/staticData"
	"math/rand"
	"sort"
)

func getUsersInSessionOrReportFailure(data *processing.ProcessData, sessionId int64) (userIds []int64, success bool) {
	db := GetDb(data.Static)

	userIds = db.GetUsersInSession(sessionId)
	success = true

	if len(userIds) < 2 {
		data.SendMessage(data.Trans("few_players"))
		success = false
	}

	return
}

func sendNumbers(data *processing.ProcessData, userIds []int64) {
	db := GetDb(data.Static)

	maleIdx := 0
	femaleIdx := 0

	for i, userId := range userIds {
		gender := db.GetUserGender(userId)

		if gender == 0 {
			continue
		}

		trans := FindTransFunction(userId, data.Static)

		message := trans("player_number_msg", map[string]interface{} {
			"Number": i+1,
		})

		if gender & 1 != 0 {
			if len(message) > 0 {
				message += "\n"
			}
			message += trans("female_number", map[string]interface{} {
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
}

func GiveRandomNumbersToPlayers(data *processing.ProcessData, sessionId int64) {
	userIds, success := getUsersInSessionOrReportFailure(data, sessionId)

	if !success {
		return
	}

	rand.Shuffle(len(userIds), func(i, j int) { userIds[i], userIds[j] = userIds[j], userIds[i] })

	sendNumbers(data, userIds)
}

func getPlaceholders(staticData *processing.StaticProccessStructs) *static.PlaceholderInfos {
	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		config = static.StaticConfiguration{}
	}

	return &config.Placeholders
}

type placeholderMatch struct {
	at int
	len int
	matchType int
	userIndex int
}

func appendMatches(matches *[]placeholderMatch, sequence []byte, placeholder *static.PlaceholderInfo, matchType int) {
	resp := placeholder.Matcher.Match(sequence)
	defer resp.Release()

	index := 0

	for resp.HasNext() {
		items := resp.NextMatchItem(sequence)
		for _, itr := range items {
			*matches = append(*matches, placeholderMatch{
				at:itr.At,
				len:itr.KLen,
				matchType:matchType,
				userIndex:index,
			})
			index++
		}
	}
}

func removeIntersectedMatches(matches *[]placeholderMatch) {
	for i := 1; i < len(*matches); i++ {
		if (*matches)[i-1].at < (*matches)[i].at + (*matches)[i].len {
			*matches = append((*matches)[:i], (*matches)[i+1:]...)
			i--
		}
	}
}

type participatingUser struct {
	id int64
	gender int
	name string
}

func getAndRemoveParticipatingUserName(userIds *[]participatingUser, matchType int) string {
	for i, user := range *userIds {
		if matchType == 0 || (matchType & user.gender != 0) {
			*userIds = append((*userIds)[:i], (*userIds)[i+1:]...)
			return user.name
		}
	}
	return "error"
}

func SendAdvancedCommand(data *processing.ProcessData, sessionId int64, command string) {
	userIds, success := getUsersInSessionOrReportFailure(data, sessionId)

	if !success {
		return
	}

	sequence := []byte(command)
	placeholders := getPlaceholders(data.Static)
	matches := make([]placeholderMatch, 0)

	appendMatches(&matches, sequence, &placeholders.Common, 0)
	appendMatches(&matches, sequence, &placeholders.Female, 1)
	appendMatches(&matches, sequence, &placeholders.Male, 2)

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].at > matches[j].at
	})

	removeIntersectedMatches(&matches)

	/*for _, match := range matches {
		sequence = []byte(string(sequence[:match.at-match.len+1]) + "{" + strconv.Itoa(match.matchType) + ":" + strconv.Itoa(match.userIndex) + "}" + string(sequence[match.at+1:]))
	}*/

	rand.Shuffle(len(userIds), func(i, j int) {
		userIds[i], userIds[j] = userIds[j], userIds[i]
	})

	db := GetDb(data.Static)

	participatingUsers := make([]participatingUser, 0)

	for _, userId := range userIds {
		participatingUsers = append(participatingUsers, participatingUser{userId, db.GetUserGender(userId), db.GetUserName(userId)})
	}

	for _, match := range matches {
		sequence = []byte(string(sequence[:match.at-match.len+1]) + getAndRemoveParticipatingUserName(&participatingUsers, match.matchType) + string(sequence[match.at+1:]))
	}

	for _, userId := range userIds {
		chatId := db.GetChatId(userId)
		data.Static.Chat.SendMessage(chatId, string(sequence), 0)
	}
}
