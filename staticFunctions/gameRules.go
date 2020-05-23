package staticFunctions

import (
	"bytes"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	static "github.com/gameraccoon/telegram-the-king-says-bot/staticData"
	"math/rand"
	"sort"
	"strconv"
	"log"
)

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
	db := GetDb(data.Static)
	userIds := db.GetUsersInSession(sessionId)

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
	name string
}

func appendMatches(matches *[]placeholderMatch, sequence []byte, placeholder *static.PlaceholderInfo, matchType int) {
	resp := placeholder.Matcher.Match(sequence)
	defer resp.Release()

	userIndex := 0

	for resp.HasNext() {
		items := resp.NextMatchItem(sequence)
		for _, itr := range items {
			*matches = append(*matches, placeholderMatch{
				at:itr.At-itr.KLen+1,
				len:itr.KLen,
				matchType:matchType,
				userIndex:userIndex,
			})
			userIndex++
		}
	}
}

func appendOppositeMatches(matches *[]placeholderMatch, sequence []byte, placeholder *[2]string) {
	var oppositeGendersIndexes [2]int
	if rand.Intn(2) == 0 {
		oppositeGendersIndexes = [2]int{2, 1}
	} else {
		oppositeGendersIndexes = [2]int{1, 2}
	}

	for placeholderIdx, gender := range oppositeGendersIndexes {
		userIndex := 0
		placeholderSeq := []byte(placeholder[placeholderIdx])
		testSeq := sequence
		seqShift := 0
		for {
			at := bytes.Index(testSeq, placeholderSeq)
			if at == -1 {
				break
			}

			testSeq = testSeq[at+len(placeholderSeq):]

			*matches = append(*matches, placeholderMatch{
				at:at+seqShift,
				len:len(placeholderSeq),
				matchType:gender,
				userIndex:userIndex,
			})
			userIndex++
			seqShift+=at+len(placeholderSeq)
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

func findMatches(staticData *processing.StaticProccessStructs, sequence []byte) []placeholderMatch {
	placeholders := getPlaceholders(staticData)
	matches := make([]placeholderMatch, 0)

	appendMatches(&matches, sequence, &placeholders.Common, 0)
	appendMatches(&matches, sequence, &placeholders.Female, 1)
	appendMatches(&matches, sequence, &placeholders.Male, 2)
	appendOppositeMatches(&matches, sequence, &placeholders.Opposite)

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].at > matches[j].at
	})

	removeIntersectedMatches(&matches)
	return matches
}

func PreviewAdvancedCommand(data *processing.ProcessData, sessionId int64, command string) {
	sequence := []byte(command)
	matches := findMatches(data.Static, sequence)

	for _, match := range matches {
		log.Printf("seq: at=%d len=%d", match.at, match.len)
		sequence = []byte(string(sequence[:match.at]) + "{" + GetGenderPlaceholderFromId(match.matchType, data.Trans) + " #" + strconv.Itoa(match.userIndex + 1) + "}" + string(sequence[match.at+match.len:]))
	}

	data.SendMessage(string(sequence))
}

func SendAdvancedCommand(data *processing.ProcessData, sessionId int64, command string) {
	db := GetDb(data.Static)
	userIds := db.GetUsersInSession(sessionId)

	sequence := []byte(command)
	matches := findMatches(data.Static, sequence)

	rand.Shuffle(len(userIds), func(i, j int) {
		userIds[i], userIds[j] = userIds[j], userIds[i]
	})

	// fill the participating users
	participatingUsers := make([]participatingUser, 0)
	for _, userId := range userIds {
		participatingUsers = append(participatingUsers, participatingUser{userId, db.GetUserGender(userId), db.GetUserName(userId)})
	}

	// fill the names for the matches with specified genders
	for i, match := range matches {
		if match.matchType != 0 {
			matches[i].name = getAndRemoveParticipatingUserName(&participatingUsers, match.matchType)
		}
	}

	// fill the names for all the others
	for i, match := range matches {
		if len(match.name) == 0 {
			matches[i].name = getAndRemoveParticipatingUserName(&participatingUsers, match.matchType)	
		}
	}

	// replace names in the string
	for _, match := range matches {
		sequence = []byte(string(sequence[:match.at]) + "<b>" + match.name + "</b>" + string(sequence[match.at+match.len:]))
	}

	// transmit the message to all players in the session
	ResendSessionDialogs(sessionId, data.Static)
	for _, userId := range userIds {
		chatId := db.GetChatId(userId)
		data.Static.Chat.SendMessage(chatId, string(sequence), 0)
	}
}
