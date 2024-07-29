package database

import (
	"fmt"
	dbBase "github.com/gameraccoon/telegram-bot-skeleton/database"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
	"sync"
)

type GameDb struct {
	db    dbBase.Database
	mutex sync.Mutex
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func ConnectDb(path string) (database *GameDb, err error) {
	database = &GameDb{}

	err = database.db.Connect(path)

	if err != nil {
		return
	}

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" global_vars(name TEXT PRIMARY KEY" +
		",integer_value INTEGER" +
		",string_value TEXT" +
		")")

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" sessions(id INTEGER NOT NULL PRIMARY KEY" +
		",token TEXT NOT NULL" +
		")")

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" users(id INTEGER NOT NULL PRIMARY KEY" +
		",name TEXT NOT NULL" +
		",gender INTEGER NOT NULL" +

		// session related data
		",current_session INTEGER" +
		",current_session_idle_count INTEGER NOT NULL" + // how many steps player didn't participate in
		")")

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" telegram_users(id INTEGER NOT NULL PRIMARY KEY" +
		",user_id INTEGER UNIQUE NOT NULL" +
		",chat_id INTEGER UNIQUE NOT NULL" +
		",language TEXT NOT NULL" +
		",ftue_completed INTEGER NOT NULL" +

		// session related data
		",current_session_message INTEGER" +
		")")

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" web_users(id INTEGER NOT NULL PRIMARY KEY" +
		",user_id INTEGER UNIQUE NOT NULL" +
		",token INTEGER UNIQUE NOT NULL" +
		")")

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" session_commands(id INTEGER NOT NULL PRIMARY KEY" +
		",session_id INTEGER NOT NULL" +
		",command TEXT NOT NULL" +
		")")

	database.db.Exec("CREATE TABLE IF NOT EXISTS" +
		" recent_web_messages(id INTEGER NOT NULL PRIMARY KEY" +
		",user_id INTEGER NOT NULL" +
		",index_for_user INTEGER NOT NULL" +
		",message TEXT NOT NULL" +
		")")

	database.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS" +
		" token_index ON sessions(token)")

	database.db.Exec("CREATE INDEX IF NOT EXISTS" +
		" current_session_index ON users(current_session)")

	database.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS" +
		" chat_id_index ON telegram_users(chat_id)")

	database.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS" +
		" user_id_index ON telegram_users(user_id)")

	database.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS" +
		" token_index ON web_users(token)")

	database.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS" +
		" user_id_index ON web_users(user_id)")

	database.db.Exec("CREATE INDEX IF NOT EXISTS" +
		" session_id_index ON session_commands(session_id)")

	database.db.Exec("CREATE INDEX IF NOT EXISTS" +
		" user_id_index ON recent_web_messages(user_id)")

	return
}

func (database *GameDb) IsConnectionOpened() bool {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	return database.db.IsConnectionOpened()
}

func (database *GameDb) Disconnect() {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Disconnect()
}

func (database *GameDb) GetDatabaseVersion() (version string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query("SELECT string_value FROM global_vars WHERE name='version'")

	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		// that means it's a new clean database
		version = latestVersion
	}

	return
}

func (database *GameDb) SetDatabaseVersion(version string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec("DELETE FROM global_vars WHERE name='version'")

	safeVersion := dbBase.SanitizeString(version)
	database.db.Exec(fmt.Sprintf("INSERT INTO global_vars (name, string_value) VALUES ('version', '%s')", safeVersion))
}

func (database *GameDb) GetOrCreateTelegramUserId(chatId int64, userLangCode string, userName string) (userId int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	// first try to find an existing user
	rows, err := database.db.Query(fmt.Sprintf("SELECT id FROM telegram_users WHERE chat_id=%d", chatId))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		// user is found, we don't need to do anything, return the id
		err := rows.Scan(&userId)
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	err = rows.Close()
	if err != nil {
		log.Fatal(err.Error())
	}

	database.db.Exec(fmt.Sprintf("INSERT INTO users(name, gender, current_session_idle_count) "+
		"VALUES ('%s', 0, 0)", dbBase.SanitizeString(userName)))

	userId = database.getLastInsertedItemId()

	database.db.Exec(fmt.Sprintf("INSERT INTO telegram_users(user_id, chat_id, language, ftue_completed) "+
		"VALUES (%d, %d, '%s', 0)", userId, chatId, userLangCode))

	return
}

func (database *GameDb) GetTelegramUserChatId(userId int64) (chatId int64, isFound bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT chat_id FROM telegram_users WHERE user_id=%d", userId))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&chatId)
		if err != nil {
			log.Fatal(err.Error())
		}
		isFound = true
	} else {
		isFound = false
	}

	return
}

func (database *GameDb) getLastInsertedItemId() (id int64) {
	rows, err := database.db.Query("SELECT last_insert_rowid()")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal("No item found")
	}
	return -1
}

func (database *GameDb) SetUserName(userId int64, name string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET name='%s' WHERE id=%d", dbBase.SanitizeString(name), userId))
}

func (database *GameDb) GetUserName(userId int64) (name string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT name FROM users WHERE id=%d", userId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("can't find name for player %d", userId)
	}

	return
}

func (database *GameDb) SetUserLanguage(userId int64, language string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK telegram_users SET language='%s' WHERE user_id=%d", dbBase.SanitizeString(language), userId))
}

func (database *GameDb) GetUserLanguage(userId int64) (language string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT language FROM telegram_users WHERE user_id=%d AND language IS NOT NULL", userId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&language)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		// empty language
	}

	return
}

func (database *GameDb) SetUserGender(userId int64, gender int) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET gender='%d' WHERE id=%d", gender, userId))
}

func (database *GameDb) GetUserGender(userId int64) (gender int) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT gender FROM users WHERE id=%d", userId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&gender)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("can't find gender for player %d", userId)
	}

	return
}

func (database *GameDb) SetUserCompletedFTUE(userId int64, isCompleted bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	value := 0
	if isCompleted {
		value = 1
	}

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK telegram_users SET ftue_completed='%d' WHERE user_id=%d", value, userId))
}

func (database *GameDb) IsUserCompletedFTUE(userId int64) (isCompleted bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT ftue_completed FROM telegram_users WHERE user_id=%d", userId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		var value int
		err := rows.Scan(&value)
		if err != nil {
			log.Fatal(err.Error())
		}
		if value != 0 {
			isCompleted = true
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("can't find ftue_completed for player %d", userId)
	}

	return
}

func (database *GameDb) GetUserSession(userId int64) (sessionId int64, isInSession bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT current_session FROM users WHERE id=%d AND current_session IS NOT NULL", userId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&sessionId)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			isInSession = true
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *GameDb) DoesSessionExist(sessionId int64) (isExists bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT 1 FROM sessions WHERE id=%d LIMIT 1", sessionId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	isExists = rows.Next()

	return
}

func (database *GameDb) CreateSession(userId int64) (sessionId int64, previousSessionId int64, wasInSession bool) {
	previousSessionId, wasInSession = database.LeaveSession(userId)

	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec("INSERT INTO sessions (token) VALUES (strftime('%s', 'now') || '-' || abs(random() % 100000))")

	sessionId = database.getLastInsertedItemId()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET current_session=%d WHERE id=%d", sessionId, userId))

	return
}

func (database *GameDb) ConnectToSession(userId int64, sessionId int64) (isSucceeded bool, previousSessionId int64, wasInSession bool) {
	if !database.DoesSessionExist(sessionId) {
		return
	}

	previousSessionId, wasInSession = database.LeaveSession(userId)

	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET current_session=%d WHERE id=%d", sessionId, userId))

	isSucceeded = true
	return
}

func (database *GameDb) GetUsersCountInSession(sessionId int64, onlyTelegramUsers bool) (usersCount int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	return database.getUsersCountInSessionUnsafe(sessionId, onlyTelegramUsers)
}

func (database *GameDb) getUsersCountInSessionUnsafe(sessionId int64, onlyTelegramUsers bool) (usersCount int64) {
	var request string
	if onlyTelegramUsers {
		request = fmt.Sprintf("SELECT COUNT(*) FROM users JOIN telegram_users ON users.id=telegram_users.user_id WHERE current_session=%d", sessionId)
	} else {
		request = fmt.Sprintf("SELECT COUNT(*) FROM users WHERE current_session=%d", sessionId)
	}

	rows, err := database.db.Query(request)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&usersCount)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *GameDb) GetUsersInSession(sessionId int64) (users []int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT id FROM users WHERE current_session=%d", sessionId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	for rows.Next() {
		var userId int64
		err := rows.Scan(&userId)
		if err != nil {
			log.Fatal(err.Error())
		}
		users = append(users, userId)
	}

	return
}

func (database *GameDb) LeaveSession(userId int64) (sessionId int64, wasInSession bool) {
	sessionId, wasInSession = database.GetUserSession(userId)

	if !wasInSession {
		return
	}

	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET current_session=NULL, current_session_idle_count=0 WHERE id=%d", userId))

	// delete session if it doesn't have Telegram users in it
	if database.getUsersCountInSessionUnsafe(sessionId, true) == 0 {
		database.db.Exec(fmt.Sprintf("DELETE FROM session_commands WHERE session_id=%d", sessionId))
		database.db.Exec(fmt.Sprintf("DELETE FROM recent_web_messages WHERE user_id in (select user_id from users where current_session=%d)", sessionId))
		database.db.Exec(fmt.Sprintf("DELETE FROM sessions WHERE id=%d", sessionId))
		database.db.Exec(fmt.Sprintf("DELETE FROM web_users WHERE user_id IN (SELECT id FROM users WHERE current_session=%d)", sessionId))
		// the remaining users that have this session is the web users that we just deleted
		database.db.Exec(fmt.Sprintf("DELETE FROM users WHERE current_session=%d", sessionId))
	}

	return
}

func (database *GameDb) SetSessionMessageId(userId int64, messageId int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK telegram_users SET current_session_message=%d WHERE user_id=%d", messageId, userId))
}

func (database *GameDb) GetSessionMessageId(userId int64) (messageId int64, isFound bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT current_session_message FROM telegram_users WHERE user_id=%d AND current_session_message IS NOT NULL", userId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&messageId)
		if err != nil {
			log.Fatal(err.Error())
		}
		isFound = true
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *GameDb) GetSessionIdFromToken(token string) (sessionId int64, isFound bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT id FROM sessions WHERE token='%s' LIMIT 1", dbBase.SanitizeString(token)))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&sessionId)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			isFound = true
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *GameDb) GetTokenFromSessionId(sessionId int64) (token string, isFound bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT token FROM sessions WHERE id=%d LIMIT 1", sessionId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&token)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			isFound = true
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *GameDb) AddSessionSuggestedCommand(sessionId int64, command string) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("INSERT INTO session_commands (session_id, command) VALUES (%d, '%s')", sessionId, dbBase.SanitizeString(command)))
}

func (database *GameDb) PopRandomSessionSuggestedCommand(sessionId int64) (command string, isSucceeded bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT id, command FROM session_commands WHERE session_id=%d ORDER BY RANDOM() LIMIT 1", sessionId))
	if err != nil {
		log.Fatal(err.Error())
	}

	if rows.Next() {
		var rowId int64
		err := rows.Scan(&rowId, &command)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}

		database.db.Exec(fmt.Sprintf("DELETE FROM session_commands WHERE id=%d", rowId))

		isSucceeded = true
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	return
}

func (database *GameDb) GetSessionSuggestedCommandCount(sessionId int64) (commandsCount int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT COUNT(*) FROM session_commands WHERE session_id=%d", sessionId))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&commandsCount)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

type SessionUserInfo struct {
	UserId                  int64
	ChatId                  int64 // token for web users
	Name                    string
	Gender                  int
	CurrentSessionIdleCount int
	IsWebUser               bool
}

func (database *GameDb) GetUsersInSessionInfo(sessionId int64) (users []SessionUserInfo) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	// join users, telegram_users and web_users tables to get chat id as either chat id or token
	request := fmt.Sprintf("SELECT users.id, IFNULL(telegram_users.chat_id, web_users.token) AS chat_id, users.name, users.gender, users.current_session_idle_count, IFNULL(web_users.token, 0) AS is_web_user FROM users LEFT JOIN telegram_users ON users.id=telegram_users.user_id LEFT JOIN web_users ON users.id=web_users.user_id WHERE users.current_session=%d", sessionId)

	rows, err := database.db.Query(request)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	for rows.Next() {
		var userInfo SessionUserInfo
		var isWebUser int
		err := rows.Scan(&userInfo.UserId, &userInfo.ChatId, &userInfo.Name, &userInfo.Gender, &userInfo.CurrentSessionIdleCount, &isWebUser)
		if err != nil {
			log.Fatal(err.Error())
		}
		userInfo.IsWebUser = isWebUser != 0
		users = append(users, userInfo)
	}

	return
}

func (database *GameDb) UpdateUsersIdleCount(usersToIncrease []int64, countIncrease int, usersToReset []int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	usersToIncreaseIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(usersToIncrease)), ","), "[]")
	usersToResetIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(usersToReset)), ","), "[]")

	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET current_session_idle_count=current_session_idle_count+'%d' WHERE id IN (%s)", countIncrease, usersToIncreaseIds))
	database.db.Exec(fmt.Sprintf("UPDATE OR ROLLBACK users SET current_session_idle_count='0' WHERE id IN (%s)", usersToResetIds))
}

func (database *GameDb) AddWebUser(sessionId int64, token int64, name string, gender int) (wasAdded bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT 1 FROM web_users WHERE token=%d", token))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		return false
	}

	err = rows.Close()
	if err != nil {
		log.Fatal(err.Error())
	}

	database.db.Exec(fmt.Sprintf("INSERT INTO users (name, gender, current_session, current_session_idle_count) VALUES ('%s', %d, %d, 0)", dbBase.SanitizeString(name), gender, sessionId))

	userId := database.getLastInsertedItemId()

	database.db.Exec(fmt.Sprintf("INSERT INTO web_users (user_id, token) VALUES (%d, %d)", userId, token))

	return true
}

func (database *GameDb) RemoveWebUser(token int64) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT user_id FROM web_users WHERE token=%d", token))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	var userId int64
	if rows.Next() {
		err := rows.Scan(&userId)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		return
	}

	err = rows.Close()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	database.db.Exec(fmt.Sprintf("DELETE FROM web_users WHERE token=%d", token))
	database.db.Exec(fmt.Sprintf("DELETE FROM users WHERE id=%d", userId))
	database.db.Exec(fmt.Sprintf("DELETE FROM recent_web_messages WHERE user_id=%d", userId))
}

func (database *GameDb) DoesWebUserExist(token int64) (isExists bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT 1 FROM web_users WHERE token=%d", token))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	isExists = rows.Next()

	return
}

func (database *GameDb) GetWebUserId(token int64) (userId int64, isFound bool) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	rows, err := database.db.Query(fmt.Sprintf("SELECT user_id FROM web_users WHERE token=%d", token))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	if rows.Next() {
		err := rows.Scan(&userId)
		if err != nil {
			log.Fatal(err.Error())
		}
		isFound = true
	} else {
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func (database *GameDb) AddWebMessage(userId int64, command string, limit int) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	database.db.Exec(fmt.Sprintf("INSERT INTO recent_web_messages (user_id, index_for_user, message) VALUES (%d, (SELECT IFNULL(MAX(index_for_user), -1) FROM recent_web_messages WHERE user_id=%d) + 1, '%s')", userId, userId, dbBase.SanitizeString(command)))
	database.db.Exec(fmt.Sprintf("DELETE FROM recent_web_messages WHERE user_id=%d AND index_for_user<=((SELECT MAX(index_for_user) FROM recent_web_messages WHERE user_id=%d) - %d)", userId, userId, limit))
}

func (database *GameDb) GetNewRecentWebMessages(userId int64, lastIndex int) (commands []string, newLastIndex int) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	newLastIndex = lastIndex

	rows, err := database.db.Query(fmt.Sprintf("SELECT message, index_for_user FROM recent_web_messages WHERE user_id=%d AND index_for_user>%d", userId, lastIndex))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	for rows.Next() {
		var command string
		err := rows.Scan(&command, &newLastIndex)
		if err != nil {
			log.Fatal(err.Error())
		}
		commands = append(commands, command)
	}

	return
}
