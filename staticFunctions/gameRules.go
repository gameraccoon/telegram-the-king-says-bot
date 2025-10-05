package staticFunctions

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-the-king-says-bot/database"
	static "github.com/gameraccoon/telegram-the-king-says-bot/staticData"
)

const NumericPositionMatchTypeOffset = 100

func sendNumbers(staticData *processing.StaticProccessStructs, userIds []int64) {
	db := GetDb(staticData)

	maleIdx := 0
	femaleIdx := 0

	for i, userId := range userIds {
		role := db.GetUserRole(userId)

		if role == 0 {
			continue
		}

		trans := FindTransFunction(userId, staticData)

		message := trans("player_number_msg", map[string]interface{}{
			"Number": i + 1,
		})

		if role&1 != 0 {
			if len(message) > 0 {
				message += "\n"
			}
			message += trans("female_number", map[string]interface{}{
				"Number": femaleIdx + 1,
			})
			femaleIdx++
		}

		if role&2 != 0 {
			if len(message) > 0 {
				message += "\n"
			}
			message += trans("male_number", map[string]interface{}{
				"Number": maleIdx + 1,
			})
			maleIdx++
		}

		chatId, isFound := db.GetTelegramUserChatId(userId)
		if isFound {
			staticData.Chat.SendMessage(chatId, message, 0, true)
		} else {
			db.AddWebMessage(userId, message, 10)
		}
	}
}

func GiveRandomNumbersToPlayers(staticData *processing.StaticProccessStructs, sessionId int64) {
	db := GetDb(staticData)
	userIds := db.GetUsersInSession(sessionId)

	rand.Shuffle(len(userIds), func(i, j int) { userIds[i], userIds[j] = userIds[j], userIds[i] })

	sendNumbers(staticData, userIds)
}

func getPlaceholders(staticData *processing.StaticProccessStructs) *static.PlaceholderInfos {
	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		config = static.StaticConfiguration{}
	}

	return &config.Placeholders
}

type placeholderMatch struct {
	at        int
	len       int
	matchType int
	name      string
}

func appendMatches(matches *[]placeholderMatch, sequence []byte, placeholder *static.PlaceholderInfo, matchType int) {
	resp := placeholder.Matcher.Match(sequence)
	defer resp.Release()

	for resp.HasNext() {
		items := resp.NextMatchItem(sequence)
		for _, itr := range items {
			*matches = append(*matches, placeholderMatch{
				at:        itr.At - itr.KLen + 1,
				len:       itr.KLen,
				matchType: matchType,
			})
		}
	}
}

func appendOppositeMatches(matches *[]placeholderMatch, sequence []byte, placeholder *[2]static.PlaceholderInfo) {
	var oppositeRoleIndexes [2]int
	if rand.Intn(2) == 0 {
		oppositeRoleIndexes = [2]int{2, 1}
	} else {
		oppositeRoleIndexes = [2]int{1, 2}
	}

	for placeholderIdx, role := range oppositeRoleIndexes {
		resp := placeholder[placeholderIdx].Matcher.Match(sequence)

		for resp.HasNext() {
			items := resp.NextMatchItem(sequence)
			for _, itr := range items {
				*matches = append(*matches, placeholderMatch{
					at:        itr.At - itr.KLen + 1,
					len:       itr.KLen,
					matchType: role,
				})
			}
		}

		resp.Release()
	}
}

func appendRepeatedPositionMatches(matches *[]placeholderMatch, sequence []byte, placeholder *[10]static.PlaceholderInfo) {
	for placeholderIdx := range placeholder {
		resp := placeholder[placeholderIdx].Matcher.Match(sequence)

		for resp.HasNext() {
			items := resp.NextMatchItem(sequence)
			for _, itr := range items {
				*matches = append(*matches, placeholderMatch{
					at:        itr.At - itr.KLen + 1,
					len:       itr.KLen,
					matchType: placeholderIdx + NumericPositionMatchTypeOffset,
				})
			}
		}

		resp.Release()
	}
}

func removeIntersectedMatches(matches *[]placeholderMatch) {
	for i := 1; i < len(*matches); i++ {
		if (*matches)[i-1].at < (*matches)[i].at+(*matches)[i].len {
			*matches = append((*matches)[:i], (*matches)[i+1:]...)
			i--
		}
	}
}

func getAndRemoveParticipatingUserName(users *[]database.SessionUserInfo, matchType int) string {
	for i, user := range *users {
		if matchType == 0 || (matchType&user.Role != 0) {
			*users = append((*users)[:i], (*users)[i+1:]...)
			return user.Name
		}
	}
	return "[no match]"
}

func getNumericPlaceholderUserName(matches *[]placeholderMatch, matchType int) string {
	index := 0
	targetIndex := matchType - NumericPositionMatchTypeOffset
	// matches are sorted from back to front, so reverse the order
	for i := len(*matches) - 1; i >= 0; i -= 1 {
		if (*matches)[i].matchType < NumericPositionMatchTypeOffset {
			if index == targetIndex {
				return (*matches)[i].name
			}
			index += 1
		}
	}

	return "[bad number]"
}

func findMatches(staticData *processing.StaticProccessStructs, sequence []byte) []placeholderMatch {
	placeholders := getPlaceholders(staticData)
	matches := make([]placeholderMatch, 0)

	appendMatches(&matches, sequence, &placeholders.Common, 0)
	appendMatches(&matches, sequence, &placeholders.Female, 1)
	appendMatches(&matches, sequence, &placeholders.Male, 2)
	appendOppositeMatches(&matches, sequence, &placeholders.Opposite)
	appendRepeatedPositionMatches(&matches, sequence, &placeholders.Positions)

	// sort to reduce the algorithmic complexity of removing intersections
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].at > matches[j].at
	})

	removeIntersectedMatches(&matches)
	return matches
}

func getUserWeight(user *database.SessionUserInfo) int {
	weightPower := user.CurrentSessionIdleCount
	if weightPower > 31 {
		weightPower = 31
	}

	weight := 1 << weightPower

	if weight <= 0 {
		panic("User weight can never be zero or lower")
	}
	return weight
}

func deleteUnordered(s *[]database.SessionUserInfo, i int) {
	(*s)[i] = (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
}

func weightedShuffle(users []database.SessionUserInfo) []database.SessionUserInfo {
	if len(users) <= 1 {
		return users
	}

	usersToDrawFrom := make([]database.SessionUserInfo, len(users))
	copy(usersToDrawFrom, users)

	var result []database.SessionUserInfo

	sum := 0
	for _, user := range usersToDrawFrom {
		sum += getUserWeight(&user)
	}

	for len(usersToDrawFrom) > 1 {
		weight := rand.Intn(sum)
		for i, user := range usersToDrawFrom {
			userWeight := getUserWeight(&user)
			if weight < userWeight {
				result = append(result, user)
				deleteUnordered(&usersToDrawFrom, i)
				sum -= userWeight
				break
			}
			weight -= userWeight
		}
	}

	if sum != getUserWeight(&usersToDrawFrom[0]) {
		panic(fmt.Sprintf("The final list doesn't contain some of the records, missing weight %d", sum-getUserWeight(&usersToDrawFrom[0])))
	}
	result = append(result, usersToDrawFrom[0])

	return result
}

func contains(slice []int64, val int64) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func SendAdvancedCommand(staticData *processing.StaticProccessStructs, sessionId int64, command string) {
	db := GetDb(staticData)
	users := db.GetUsersInSessionInfo(sessionId)

	{
		commandLength := 0
		for {
			command = strings.ReplaceAll(strings.ReplaceAll(command, "<b>", ""), "</b>", "")
			if len(command) == commandLength {
				break
			}
			commandLength = len(command)
		}
	}

	sequence := []byte(command)
	matches := findMatches(staticData, sequence)

	// the players will be picked from the beginning first
	// the users with higher weight are more likely to end up in the beginning
	participatingUsers := weightedShuffle(users)

	// randomize the order of picking for placeholders, to avoid any bias
	// e.g. if we would keep it sorted from high to low, we would most likely pick
	// the players with higher weights for the last placeholders in the command
	rand.Shuffle(len(matches), func(i, j int) { matches[i], matches[j] = matches[j], matches[i] })

	// fill the names for the matches with specified roles
	for i, match := range matches {
		if match.matchType == 1 || match.matchType == 2 {
			matches[i].name = getAndRemoveParticipatingUserName(&participatingUsers, match.matchType)
		}
	}

	// fill the names for all the others
	for i, match := range matches {
		if len(match.name) == 0 && match.matchType < NumericPositionMatchTypeOffset {
			matches[i].name = getAndRemoveParticipatingUserName(&participatingUsers, match.matchType)
		}
	}

	// fill repeated indexes
	for i, match := range matches {
		if match.matchType >= NumericPositionMatchTypeOffset {
			matches[i].name = getNumericPlaceholderUserName(&matches, match.matchType)
		}
	}

	// sort again, from higher to lower, to not shift the indexes while using them
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].at > matches[j].at
	})

	// replace names in the string
	for _, match := range matches {
		sequence = []byte(string(sequence[:match.at]) + "<b>" + match.name + "</b>" + string(sequence[match.at+match.len:]))
	}

	message := string(sequence)

	// transmit the message to all players in the session
	ResendSessionDialogs(sessionId, staticData)
	for _, user := range users {
		if user.IsWebUser {
			db.AddWebMessage(user.UserId, message, 10)
		} else {
			staticData.Chat.SendMessage(user.ChatId, message, 0, true)
		}
	}

	// increase idle counters for players who didn't participate and reset for the ones who participated
	{
		var nonParticipatedIds []int64
		for _, user := range participatingUsers {
			nonParticipatedIds = append(nonParticipatedIds, user.UserId)
		}
		var participatedIds []int64
		for _, user := range users {
			if !contains(nonParticipatedIds, user.UserId) {
				participatedIds = append(participatedIds, user.UserId)
			}
		}
		db.UpdateUsersIdleCount(nonParticipatedIds, 1, participatedIds)
	}
}
