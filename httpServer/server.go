package httpServer

import (
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/gameraccoon/telegram-the-king-says-bot/database"
	"github.com/gameraccoon/telegram-the-king-says-bot/staticFunctions"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/html/index.html")
}

func invitePage(w http.ResponseWriter, r *http.Request, db *database.GameDb) {
	gameToken := r.URL.Path[len("/invite/"):]
	if gameToken == "" {
		http.Error(w, "Incorrect URL", http.StatusBadRequest)
		return
	}

	_, isFound := db.GetSessionIdFromToken(gameToken)
	if isFound {
		http.ServeFile(w, r, "data/html/invite.html")
	} else {
		http.ServeFile(w, r, "data/html/invite_no_session.html")
	}
}

func joinGame(w http.ResponseWriter, r *http.Request, db *database.GameDb, staticData *processing.StaticProccessStructs) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	gameId := r.Form.Get("gameId")

	if gameId == "" {
		http.Error(w, "Incorrect game id, reload the page and try again", http.StatusBadRequest)
		return
	}

	sessionId, isFound := db.GetSessionIdFromToken(gameId)
	if !isFound {
		http.Error(w, "Game not found. Was it ended?", http.StatusBadRequest)
		return
	}

	name := r.Form.Get("name")
	if name == "" {
		http.Error(w, "The name is empty", http.StatusBadRequest)
		return
	}

	if len(name) > 20 {
		http.Error(w, "The name is too long", http.StatusBadRequest)
		return
	}

	gender := r.Form.Get("gender")
	// convert from string to int
	genderInt := 0
	if gender == "g" {
		genderInt = 1
	} else if gender == "b" {
		genderInt = 2
	} else if gender == "a" {
		genderInt = 3
	} else if gender != "n" {
		genderInt = 0
	} else {
		http.Error(w, "Incorrect gender code", http.StatusBadRequest)
	}

	token := int64(rand.Uint64() & 0x7FFFFFFFFFFFFFFF)

	hasAdded := db.AddWebUser(sessionId, token, name, genderInt)

	if !hasAdded {
		http.Error(w, "Can't add new user, try again", http.StatusBadRequest)
		return
	}

	staticFunctions.UpdateSessionDialogs(sessionId, staticData)

	stringToken := strconv.FormatInt(token, 10)

	_, err = w.Write([]byte(stringToken))
	if err != nil {
		return
	}
}

func gamePage(w http.ResponseWriter, r *http.Request, db *database.GameDb) {
	if r.Method != "GET" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// get player token from URL
	playerTokenStr := r.URL.Path[len("/user/"):]
	if playerTokenStr == "" {
		http.Error(w, "Incorrect URL", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	_, isFound := db.GetWebUserId(playerToken)
	if isFound {
		http.ServeFile(w, r, "data/html/user.html")
	} else {
		http.ServeFile(w, r, "data/html/invite_no_session.html")
	}
}

func getLastMessages(w http.ResponseWriter, r *http.Request, db *database.GameDb) {
	if r.Method != "GET" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()

	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	playerTokenStr := r.Form.Get("playerToken")
	if playerTokenStr == "" {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	userId, isFound := db.GetWebUserId(playerToken)
	if !isFound {
		http.Error(w, "Player not found, has the game ended?", http.StatusNotFound)
		return
	}

	lastMessageIdxStr := r.Form.Get("lastMessageIdx")
	if lastMessageIdxStr == "" {
		http.Error(w, "Incorrect last message index", http.StatusBadRequest)
		return
	}

	lastMessageIdx, err := strconv.Atoi(lastMessageIdxStr)
	if err != nil {
		http.Error(w, "Incorrect last message index", http.StatusBadRequest)
		return
	}

	sessionId, isInSession := db.GetUserSession(userId)
	if !isInSession {
		http.Error(w, "Player not in session, has the game ended?", http.StatusNotFound)
		return
	}

	messages, newLastIdx := db.GetNewRecentlySentCommands(sessionId, lastMessageIdx)

	w.Header().Set("Content-Type", "application/json")
	messagesStr := ""
	for i, message := range messages {
		if i > 0 {
			messagesStr += "\",\""
		}

		sanitizedString := strings.Replace(message, "\n", " ", -1)
		sanitizedString = strings.Replace(sanitizedString, "\"", "\\\"", -1)
		messagesStr += sanitizedString
	}

	suggestedCount := db.GetSessionSuggestedCommandCount(sessionId)

	playersCount := db.GetUsersCountInSession(sessionId, false)

	_, err = w.Write([]byte("{\"lastMessageIdx\":" + strconv.Itoa(newLastIdx) + ",\"players\":" + strconv.FormatInt(playersCount, 10) + ",\"suggestions\":" + strconv.FormatInt(suggestedCount, 10) + ",\"messages\":[\"" + messagesStr + "\"]}"))
}

func suggestCommand(w http.ResponseWriter, r *http.Request, db *database.GameDb, staticData *processing.StaticProccessStructs) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	playerTokenStr := r.Form.Get("playerToken")
	if playerTokenStr == "" {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	userId, isFound := db.GetWebUserId(playerToken)
	if !isFound {
		http.Error(w, "Player not found, has the game ended?", http.StatusNotFound)
		return
	}

	sessionId, isInSession := db.GetUserSession(userId)
	if !isInSession {
		http.Error(w, "Player not in session, has the game ended?", http.StatusNotFound)
		return
	}

	command := r.Form.Get("command")
	if command == "" {
		http.Error(w, "The command is empty", http.StatusBadRequest)
		return
	}

	db.AddSessionSuggestedCommand(sessionId, command)

	staticFunctions.UpdateSessionDialogs(sessionId, staticData)

	_, err = w.Write([]byte("ok"))
	if err != nil {
		return
	}
}

func revealSuggestedCommand(w http.ResponseWriter, r *http.Request, db *database.GameDb, staticData *processing.StaticProccessStructs) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	playerTokenStr := r.Form.Get("playerToken")
	if playerTokenStr == "" {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	userId, isFound := db.GetWebUserId(playerToken)
	if !isFound {
		http.Error(w, "Player not found, has the game ended?", http.StatusNotFound)
		return
	}

	sessionId, isInSession := db.GetUserSession(userId)
	if !isInSession {
		http.Error(w, "Player not in session, has the game ended?", http.StatusNotFound)
		return
	}

	command, isSucceeded := db.PopRandomSessionSuggestedCommand(sessionId)
	staticFunctions.UpdateSessionDialogs(sessionId, staticData)

	if isSucceeded {
		staticFunctions.SendAdvancedCommand(staticData, sessionId, command)
	} else {
		_, err = w.Write([]byte("List of commands is empty"))
		if err != nil {
			return
		}
	}

	_, err = w.Write([]byte("ok"))
}

func leaveGame(w http.ResponseWriter, r *http.Request, db *database.GameDb, staticData *processing.StaticProccessStructs) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	playerTokenStr := r.Form.Get("playerToken")
	if playerTokenStr == "" {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	playerToken, err := strconv.ParseInt(playerTokenStr, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect player token", http.StatusBadRequest)
		return
	}

	userId, isFound := db.GetWebUserId(playerToken)
	if !isFound {
		http.Error(w, "Player not found, has the game ended?", http.StatusNotFound)
		return
	}

	sessionId, isInSession := db.GetUserSession(userId)
	if !isInSession {
		http.Error(w, "Player not in session, has the game ended?", http.StatusNotFound)
		return
	}

	db.RemoveWebUser(playerToken)

	staticFunctions.UpdateSessionDialogs(sessionId, staticData)

	_, err = w.Write([]byte("ok"))
	if err != nil {
		return
	}
}

func HandleHttpRequests(port int, staticData *processing.StaticProccessStructs) {
	db := staticFunctions.GetDb(staticData)

	http.HandleFunc("/", homePage)
	http.HandleFunc("/invite/", func(w http.ResponseWriter, r *http.Request) {
		invitePage(w, r, db)
	})
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		joinGame(w, r, db, staticData)
	})
	http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		gamePage(w, r, db)
	})
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		getLastMessages(w, r, db)
	})
	http.HandleFunc("/suggest", func(w http.ResponseWriter, r *http.Request) {
		suggestCommand(w, r, db, staticData)
	})
	http.HandleFunc("/reveal", func(w http.ResponseWriter, r *http.Request) {
		revealSuggestedCommand(w, r, db, staticData)
	})
	http.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) {
		leaveGame(w, r, db, staticData)
	})

	addr := ":" + strconv.Itoa(port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
