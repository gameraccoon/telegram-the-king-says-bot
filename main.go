package main

import (
	"encoding/json"
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogManager"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-bot-skeleton/telegramChat"
	"github.com/gameraccoon/telegram-the-king-says-bot/database"
	"github.com/gameraccoon/telegram-the-king-says-bot/dialogFactories"
	static "github.com/gameraccoon/telegram-the-king-says-bot/staticData"
	"github.com/nicksnyder/go-i18n/i18n"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func getFileStringContent(filePath string) (content string, err error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err == nil {
		content = strings.TrimSpace(string(fileContent))
	}
	return
}

func getApiToken() (token string, err error) {
	return getFileStringContent("./telegramApiToken.txt")
}

func loadConfig(path string) (config static.StaticConfiguration, err error) {
	jsonString, err := getFileStringContent(path)
	if err == nil {
		dec := json.NewDecoder(strings.NewReader(jsonString))
		err = dec.Decode(&config)
	}

	if err == nil {
		jsonString, err := getFileStringContent("./data/placeholders.json")
		if err == nil {
			dec := json.NewDecoder(strings.NewReader(jsonString))
			err = dec.Decode(&config.Placeholders)
			if err == nil {
				config.Placeholders.Compile()
			}
		}
	}
	return
}

func main() {
	rand.Seed(time.Now().UnixNano())

	apiToken, err := getApiToken()
	if err != nil {
		log.Fatal(err.Error())
	}

	config, err := loadConfig("./config.json")
	if err != nil {
		log.Fatal(err.Error())
	}

	translators := make(map[string]i18n.TranslateFunc)

	for _, lang := range config.AvailableLanguages {
		i18n.MustLoadTranslationFile("./data/strings/" + lang.Key + ".all.json")

		trans, err1 := i18n.Tfunc(lang.Key)
		if err1 != nil {
			log.Fatal(err.Error())
		}
		translators[lang.Key] = trans
	}

	if len(translators) <= 0 {
		log.Fatal("Need at least one language available")
	}

	if _, ok := translators[config.DefaultLanguage]; !ok {
		log.Fatal("Default language should be in the list of available languages")
	}

	db, err := database.ConnectDb("./bot-data.db")
	defer db.Disconnect()

	if err != nil {
		log.Fatal("Can't connect database")
	}

	database.UpdateVersion(db)

	chat, err := telegramChat.MakeTelegramChat(apiToken)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Authorized on account %s", chat.GetBotUsername())

	chat.SetDebugModeEnabled(config.ExtendedLog)

	dialogManager := &(dialogManager.DialogManager{})
	dialogManager.RegisterDialogFactory("us", dialogFactories.MakeUserSettingsDialogFactory())
	dialogManager.RegisterDialogFactory("lc", dialogFactories.MakeLanguageSelectDialogFactory())
	dialogManager.RegisterDialogFactory("gc", dialogFactories.MakeGenderSelectDialogFactory())
	dialogManager.RegisterDialogFactory("se", dialogFactories.MakeSessionDialogFactory())
	dialogManager.RegisterDialogFactory("ns", dialogFactories.MakeNoSessionDialogFactory())
	dialogManager.RegisterDialogFactory("sc", dialogFactories.MakeSuggestedConfirmedDialogFactory())
	dialogManager.RegisterTextInputProcessorManager(dialogFactories.GetTextInputProcessorManager())

	staticData := &processing.StaticProccessStructs{
		Chat:   chat,
		Db:     db,
		Config: config,
		Trans:  translators,
		MakeDialogFn: func(id string, userId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs, customData interface{}) *dialog.Dialog {
			return dialogManager.MakeDialog(id, userId, trans, staticData, customData)
		},
		BotName: chat.GetBotUsername(),
	}

	staticData.Init()

	startUpdating(chat, dialogManager, staticData)
}
