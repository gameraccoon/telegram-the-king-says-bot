package staticFunctions

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-the-king-says-bot/database"
	static "github.com/gameraccoon/telegram-the-king-says-bot/staticData"
	"github.com/nicksnyder/go-i18n/i18n"
	"log"
	"strings"
	"time"
)

func GetDb(staticData *processing.StaticProccessStructs) *database.SpyBotDb {
	if staticData == nil {
		log.Fatal("staticData is nil")
		return nil
	}

	db, ok := staticData.Db.(*database.SpyBotDb)
	if ok && db != nil {
		return db
	} else {
		log.Fatal("database is not set properly")
		return nil
	}
}

func getClosestLang(config *static.StaticConfiguration, lang string) string {
	for _, langCode := range config.AvailableLanguages {
		if strings.HasPrefix(langCode.Key, lang) {
			return langCode.Key
		}
	}
	return lang
}

func FindTransFunction(userId int64, staticData *processing.StaticProccessStructs) i18n.TranslateFunc {
	// ToDo: cache user's lang
	lang := GetDb(staticData).GetUserLanguage(userId)

	config, configCastSuccess := staticData.Config.(static.StaticConfiguration)

	if !configCastSuccess {
		config = static.StaticConfiguration{}
	}

	// replace empty language to default one (some clients don't send user's language)
	if len(lang) <= 0 {
		log.Printf("User %d has empty language. Setting to default.", userId)
		lang = config.DefaultLanguage
	}

	if foundTrans, ok := staticData.Trans[lang]; ok {
		GetDb(staticData).SetUserLanguage(userId, lang)
		return foundTrans
	}

	if foundTrans, ok := staticData.Trans[getClosestLang(&config, lang)]; ok {
		GetDb(staticData).SetUserLanguage(userId, lang)
		return foundTrans
	}	

	// unknown language, use default instead
	if foundTrans, ok := staticData.Trans[config.DefaultLanguage]; ok {
		log.Printf("User %d has unknown language (%s). Setting to default.", userId, lang)
		lang = config.DefaultLanguage
		GetDb(staticData).SetUserLanguage(userId, lang)
		return foundTrans
	}

	// something gone wrong
	log.Printf("Translator didn't found: %s", lang)
	// fall to the first available translator
	for lang, trans := range staticData.Trans {
		log.Printf("Using first available translator: %s", lang)
		return trans
	}

	// something gone completely wrong
	log.Fatal("There are no available translators")
	// we will probably crash but there is nothing else we can do
	translator, _ := i18n.Tfunc(config.DefaultLanguage)
	return translator
}

func FormatTimestamp(timestamp time.Time, timezone string) string {
	// the list of timezones https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	loc, err := time.LoadLocation(timezone)
	if err == nil {
		return timestamp.In(loc).Format("15:04:05 _2.01.2006")
	} else {
		return timestamp.Format("15:04:05 _2.01.2006")
	}
}

func SendSessionDialogToSomeone(userId int64, chatId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs) {
	db := GetDb(staticData)
	oldMessageId, isFound := db.GetSessionMessageId(userId)
	if isFound {
		staticData.Chat.RemoveMessage(chatId, oldMessageId)
	}

	newMssageId := staticData.Chat.SendDialog(chatId, staticData.MakeDialogFn("se", userId, trans, staticData, nil), 0)
	db.SetSessionMessageId(userId, newMssageId)
}

func SendSessionDialog(data *processing.ProcessData) {
	SendSessionDialogToSomeone(data.UserId, data.ChatId, data.Trans, data.Static)
}

func SendNoSessionDialog(data *processing.ProcessData) {
	db := GetDb(data.Static)
	oldMessageId, isFound := db.GetSessionMessageId(data.UserId)
	if isFound {
		data.Static.Chat.RemoveMessage(data.ChatId, oldMessageId)
	}

	newMssageId := data.SendDialog(data.Static.MakeDialogFn("ns", data.UserId, data.Trans, data.Static, nil))
	db.SetSessionMessageId(data.UserId, newMssageId)
}

func UpdateSessionDialogs(sessionId int64, staticData *processing.StaticProccessStructs) {
	users := GetDb(staticData).GetUsersInSession(sessionId)
	db := GetDb(staticData)

	for _, userId := range users {
		messageId, isFound := db.GetSessionMessageId(userId)
		if isFound {
			trans := FindTransFunction(userId, staticData)
			chatId := db.GetChatId(userId)
			staticData.Chat.SendDialog(chatId, staticData.MakeDialogFn("se", userId, trans, staticData, nil), messageId)
		}
	}
}

func ResendSessionDialogs(sessionId int64, staticData *processing.StaticProccessStructs) {
	users := GetDb(staticData).GetUsersInSession(sessionId)
	db := GetDb(staticData)

	for _, userId := range users {
		trans := FindTransFunction(userId, staticData)
		chatId := db.GetChatId(userId)

		SendSessionDialogToSomeone(userId, chatId, trans, staticData)
	}
}

func ConnectToSession(data *processing.ProcessData, token string) (successfull bool) {
	db := GetDb(data.Static)
	sessionId, isFound := db.GetSessionIdFromToken(token)
	if !isFound {
		return false
	}

	successfullyConnected, previousSessionId, wasInSession := db.ConnectToSession(data.UserId, sessionId)
	if !successfullyConnected {
		return false
	}

	UpdateSessionDialogs(sessionId, data.Static)

	if wasInSession {
		UpdateSessionDialogs(previousSessionId, data.Static)
	}

	return true
}

func GetGenderNameFromId(gender int, trans i18n.TranslateFunc) string {
	if gender == 1 {
		return trans("gender_female")
	} else if gender == 2 {
		return trans("gender_male")
	} else if gender == 3 {
		return trans("gender_both")
	} else {
		return trans("gender_none")
	}
}

func GetGenderPlaceholderFromId(gender int, trans i18n.TranslateFunc) string {
	if gender == 1 {
		return trans("gender_female")
	} else if gender == 2 {
		return trans("gender_male")
	} else {
		return trans("gender_any")
	}
}

func FirstSetUpStep1(data *processing.ProcessData) (inProgress bool) {
	db := GetDb(data.Static)
	if !db.IsUserCompletedFTUE(data.UserId) {
		data.SendDialog(data.Static.MakeDialogFn("lc", data.UserId, data.Trans, data.Static, nil))
		inProgress = true
	}
	return
}

func FirstSetUpStep2(data *processing.ProcessData) {
	db := GetDb(data.Static)

	if !db.IsUserCompletedFTUE(data.UserId) {
		data.SendMessage(data.Trans("enter_name"))
		data.Static.SetUserStateTextProcessor(data.UserId, &processing.AwaitingTextProcessorData{
			ProcessorId:  "changeName",
			AdditionalId: data.UserId,
		})
	}
}

func FirstSetUpStep3(data *processing.ProcessData) {
	db := GetDb(data.Static)
	if !db.IsUserCompletedFTUE(data.UserId) {
		data.SendDialog(data.Static.MakeDialogFn("gc", data.UserId, data.Trans, data.Static, nil))
	}
}

func FirstSetUpStep4(data *processing.ProcessData) {
	db := GetDb(data.Static)
	if !db.IsUserCompletedFTUE(data.UserId) {
		_, isInSession := db.GetUserSession(data.UserId)
		if isInSession {
			SendSessionDialog(data)
		} else {
			data.SendDialog(data.Static.MakeDialogFn("ns", data.UserId, data.Trans, data.Static, nil))
		}
		data.Static.SetUserStateTextProcessor(data.UserId, nil)
	}
	db.SetUserCompletedFTUE(data.UserId, true)
}
