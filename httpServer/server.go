package httpServer

import (
	"bytes"
	"github.com/gameraccoon/telegram-the-king-says-bot/database"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "data/html/index.html")
}

func invitePage(w http.ResponseWriter, r *http.Request, db *database.GameDb) {
	content, err := os.ReadFile("data/html/invite.html")
	if err != nil {
		http.Error(w, "Can't read invite.html", http.StatusInternalServerError)
		return
	}

	gameToken := r.URL.Path[len("/invite/"):]
	if gameToken == "" {
		http.Error(w, "Incorrect URL", http.StatusBadRequest)
		return
	}

	_, isFound := db.GetSessionIdFromToken(gameToken)
	if !isFound {
		gameToken = ""
	}

	content = bytes.ReplaceAll(content, []byte("{{.GameId}}"), []byte(gameToken))

	_, err = w.Write(content)
}

func joinGame(w http.ResponseWriter, r *http.Request, db *database.GameDb) {
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
		http.Error(w, "Game not found. Did the host delete it?", http.StatusBadRequest)
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

	stringToken := strconv.FormatInt(token, 10)

	_, err = w.Write([]byte(stringToken))
	if err != nil {
		return
	}
}

func gamePage(w http.ResponseWriter, r *http.Request, db *database.GameDb) {
	http.ServeFile(w, r, "data/html/game.html")
}

func HandleHttpRequests(port int, db *database.GameDb) {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/invite/", func(w http.ResponseWriter, r *http.Request) {
		invitePage(w, r, db)
	})
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		joinGame(w, r, db)
	})
	http.HandleFunc("/game/", func(w http.ResponseWriter, r *http.Request) {
		gamePage(w, r, db)
	})

	addr := ":" + strconv.Itoa(port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
