package database

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	testDbPath = "./testDb.db"
)

func dropDatabase(fileName string) {
	_ = os.Remove(fileName)
}

func clearDb() {
	dropDatabase(testDbPath)
}

func connectDb(t *testing.T) *GameDb {
	assert := require.New(t)
	db, err := ConnectDb(testDbPath)

	if err != nil {
		assert.Fail("Problem with creation db connection:" + err.Error())
		return nil
	}
	return db
}

func createDbAndConnect(t *testing.T) *GameDb {
	clearDb()
	return connectDb(t)
}

func TestConnection(t *testing.T) {
	assert := require.New(t)
	dropDatabase(testDbPath)

	db, err := ConnectDb(testDbPath)

	defer dropDatabase(testDbPath)
	if err != nil {
		assert.Fail("Problem with creation db connection:" + err.Error())
		return
	}

	assert.True(db.IsConnectionOpened())

	db.Disconnect()

	assert.False(db.IsConnectionOpened())
}

func TestSanitizeString(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	testText := "text'test''test\"test\\"

	db.SetDatabaseVersion(testText)
	assert.Equal(testText, db.GetDatabaseVersion())
}

func TestDatabaseVersion(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}

	{
		version := db.GetDatabaseVersion()
		assert.Equal(latestVersion, version)
	}

	{
		db.SetDatabaseVersion("1.0")
		version := db.GetDatabaseVersion()
		assert.Equal("1.0", version)
	}

	db.Disconnect()

	{
		db = connectDb(t)
		version := db.GetDatabaseVersion()
		assert.Equal("1.0", version)
		db.Disconnect()
	}

	{
		db = connectDb(t)
		db.SetDatabaseVersion("1.2")
		db.Disconnect()
	}

	{
		db = connectDb(t)
		version := db.GetDatabaseVersion()
		assert.Equal("1.2", version)
		db.Disconnect()
	}
}

func TestGetUserId(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	var chatId1 int64 = 321
	var chatId2 int64 = 123

	id1 := db.GetOrCreateTelegramUserId(chatId1, "", "")
	id2 := db.GetOrCreateTelegramUserId(chatId1, "", "")
	id3 := db.GetOrCreateTelegramUserId(chatId2, "", "")

	assert.Equal(id1, id2)
	assert.NotEqual(id1, id3)

	userChatId1, found := db.GetTelegramUserChatId(id1)
	assert.True(found)
	assert.Equal(chatId1, userChatId1)
	userChatId3, found := db.GetTelegramUserChatId(id3)
	assert.True(found)
	assert.Equal(chatId2, userChatId3)
}

func TestUserName(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	testNameDefault1 := "default"
	testName1 := "Te'stName"
	testName2 := "Test'name2"

	userId1 := db.GetOrCreateTelegramUserId(321, "", testNameDefault1)
	userId2 := db.GetOrCreateTelegramUserId(123, "", testName2)

	{
		assert.Equal(testNameDefault1, db.GetUserName(userId1))
		assert.Equal(testName2, db.GetUserName(userId2))
	}

	db.SetUserName(userId1, testName1)

	{
		assert.Equal(testName1, db.GetUserName(userId1))
		assert.Equal(testName2, db.GetUserName(userId2))
	}
}

func TestUserLanguage(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "", "")
	userId2 := db.GetOrCreateTelegramUserId(321, "", "")

	db.SetUserLanguage(userId1, "en-US")

	{
		lang1 := db.GetUserLanguage(userId1)
		lang2 := db.GetUserLanguage(userId2)
		assert.Equal("en-US", lang1)
		assert.Equal("", lang2)
	}

	// in case of some side effects
	{
		lang1 := db.GetUserLanguage(userId1)
		lang2 := db.GetUserLanguage(userId2)
		assert.Equal("en-US", lang1)
		assert.Equal("", lang2)
	}
}

func TestUserGender(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "", "")
	userId2 := db.GetOrCreateTelegramUserId(321, "", "")

	db.SetUserGender(userId1, 1)

	{
		assert.Equal(1, db.GetUserGender(userId1))
		assert.Equal(0, db.GetUserGender(userId2))
	}

	db.SetUserGender(userId2, 2)

	{
		assert.Equal(1, db.GetUserGender(userId1))
		assert.Equal(2, db.GetUserGender(userId2))
	}
}

func TestUserSession(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "", "")
	userId2 := db.GetOrCreateTelegramUserId(321, "", "")

	sessionId, _, _ := db.CreateSession(userId1)
	assert.True(db.DoesSessionExist(sessionId))

	{
		token, isFound1 := db.GetTokenFromSessionId(sessionId)
		newSessionId, isFound2 := db.GetSessionIdFromToken(token)
		assert.True(isFound1)
		assert.True(isFound2)
		assert.Equal(sessionId, newSessionId)
	}

	{
		sessionId1, isInSession1 := db.GetUserSession(userId1)
		_, isInSession2 := db.GetUserSession(userId2)
		assert.True(isInSession1)
		assert.False(isInSession2)
		assert.Equal(sessionId, sessionId1)
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId1, true))
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId1, false))

		users := db.GetUsersInSession(sessionId)
		assert.Equal(1, len(users))
		if len(users) > 0 {
			assert.Equal(userId1, users[0])
		}
	}

	db.ConnectToSession(userId2, sessionId)

	{
		sessionId1, isInSession1 := db.GetUserSession(userId1)
		sessionId2, isInSession2 := db.GetUserSession(userId2)
		assert.True(isInSession1)
		assert.True(isInSession2)
		assert.Equal(sessionId, sessionId1)
		assert.Equal(sessionId, sessionId2)
		assert.Equal(int64(2), db.GetUsersCountInSession(sessionId, true))
		assert.Equal(int64(2), db.GetUsersCountInSession(sessionId, false))
	}

	db.LeaveSession(userId1)
	assert.True(db.DoesSessionExist(sessionId))

	{
		_, isInSession1 := db.GetUserSession(userId1)
		sessionId2, isInSession2 := db.GetUserSession(userId2)
		assert.False(isInSession1)
		assert.True(isInSession2)
		assert.Equal(sessionId, sessionId2)
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId, true))
		assert.Equal(int64(1), db.GetUsersCountInSession(sessionId, false))
	}

	db.LeaveSession(userId2)
	assert.False(db.DoesSessionExist(sessionId))

	{
		_, isInSession1 := db.GetUserSession(userId1)
		_, isInSession2 := db.GetUserSession(userId2)
		assert.False(isInSession1)
		assert.False(isInSession2)
		assert.Equal(int64(0), db.GetUsersCountInSession(sessionId, true))
		assert.Equal(int64(0), db.GetUsersCountInSession(sessionId, false))
	}
}

func TestSessionMessageId(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "", "")
	sessionMessageId := int64(32)

	{
		_, isFound := db.GetSessionMessageId(userId1)
		assert.False(isFound)
	}
	db.SetSessionMessageId(userId1, sessionMessageId)

	{
		sessionId, isFound := db.GetSessionMessageId(userId1)
		assert.True(isFound)
		assert.Equal(sessionMessageId, sessionId)
	}
}

func TestSuggestedCommands(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "", "")

	testCommand1 := "test'asd"
	testCommand2 := "tefaasd'a"

	sessionId, _, _ := db.CreateSession(userId1)

	assert.Equal(int64(0), db.GetSessionSuggestedCommandCount(sessionId))

	{
		_, isSucceeded := db.PopRandomSessionSuggestedCommand(sessionId)
		assert.False(isSucceeded)
	}

	db.AddSessionSuggestedCommand(sessionId, testCommand1)
	db.AddSessionSuggestedCommand(sessionId, testCommand2)

	assert.Equal(int64(2), db.GetSessionSuggestedCommandCount(sessionId))

	{
		command, isSucceeded := db.PopRandomSessionSuggestedCommand(sessionId)
		assert.True(isSucceeded)
		assert.True(command == testCommand1 || command == testCommand2)
		assert.Equal(int64(1), db.GetSessionSuggestedCommandCount(sessionId))
	}

	db.LeaveSession(userId1)
	assert.Equal(int64(0), db.GetSessionSuggestedCommandCount(sessionId))
}

func TestFTUE(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "", "")

	assert.False(db.IsUserCompletedFTUE(userId1))

	db.SetUserCompletedFTUE(userId1, true)

	assert.True(db.IsUserCompletedFTUE(userId1))
}

func TestIdleCount(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId1 := db.GetOrCreateTelegramUserId(123, "", "a")
	db.SetUserGender(userId1, 1)
	userId2 := db.GetOrCreateTelegramUserId(234, "", "b")
	db.SetUserGender(userId2, 2)

	sessionId, _, _ := db.CreateSession(userId1)
	db.ConnectToSession(userId2, sessionId)

	assert.Equal([]SessionUserInfo{{userId1, 123, "a", 1, 0, false}, {userId2, 234, "b", 2, 0, false}}, db.GetUsersInSessionInfo(sessionId))

	db.UpdateUsersIdleCount([]int64{userId1, userId2}, 1, []int64{})

	assert.Equal([]SessionUserInfo{{userId1, 123, "a", 1, 1, false}, {userId2, 234, "b", 2, 1, false}}, db.GetUsersInSessionInfo(sessionId))

	db.UpdateUsersIdleCount([]int64{userId1}, 2, []int64{userId2})

	assert.Equal([]SessionUserInfo{{userId1, 123, "a", 1, 3, false}, {userId2, 234, "b", 2, 0, false}}, db.GetUsersInSessionInfo(sessionId))

	db.LeaveSession(userId1)
	db.ConnectToSession(userId1, sessionId)

	assert.Equal([]SessionUserInfo{{userId1, 123, "a", 1, 0, false}, {userId2, 234, "b", 2, 0, false}}, db.GetUsersInSessionInfo(sessionId))
}

func TestAddWebUser(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	webUserToken := int64(10)

	// we can add web users only if we have a session
	userId := db.GetOrCreateTelegramUserId(123, "", "test")
	sessionId, _, _ := db.CreateSession(userId)

	assert.False(db.DoesWebUserExist(webUserToken))

	wasAdded := db.AddWebUser(sessionId, webUserToken, "test name", 2)
	assert.True(wasAdded)

	assert.True(db.DoesWebUserExist(webUserToken))

	wasAdded = db.AddWebUser(sessionId, webUserToken, "test name 2", 1)
	assert.False(wasAdded) // same token

	assert.Equal(int64(1), db.GetUsersCountInSession(sessionId, true))
	assert.Equal(int64(2), db.GetUsersCountInSession(sessionId, false))

	users := db.GetUsersInSession(sessionId)
	assert.Equal(2, len(users))

	for _, user := range users {
		if user == userId {
			continue
		}
		assert.Equal("test name", db.GetUserName(user))
		assert.Equal(2, db.GetUserGender(user))
		userSessionId, isInSession := db.GetUserSession(user)
		assert.True(isInSession)
		assert.Equal(sessionId, userSessionId)
	}

	webUserId, isFound := db.GetWebUserId(webUserToken)
	assert.True(isFound)

	assert.Equal([]SessionUserInfo{{userId, 123, "test", 0, 0, false}, {webUserId, 10, "test name", 2, 0, true}}, db.GetUsersInSessionInfo(sessionId))
	sessionToken, _ := db.GetTokenFromSessionId(sessionId)

	// web users are not counted for the session survival
	db.LeaveSession(userId)

	assert.False(db.DoesSessionExist(sessionId))
	_, isSessionFound := db.GetSessionIdFromToken(sessionToken)
	assert.False(isSessionFound)
	assert.False(db.DoesWebUserExist(webUserToken))
	_, isFound = db.GetWebUserId(webUserToken)
	assert.False(isFound)
}

func TestRemoveWebUser(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	webUserToken := int64(10)

	userId := db.GetOrCreateTelegramUserId(123, "", "test")
	sessionId, _, _ := db.CreateSession(userId)

	db.AddWebUser(sessionId, webUserToken, "test name", 2)

	assert.True(db.DoesWebUserExist(webUserToken))

	db.RemoveWebUser(webUserToken)

	assert.False(db.DoesWebUserExist(webUserToken))

	assert.Equal(int64(1), db.GetUsersCountInSession(sessionId, false))
}

func TestRecentlySentCommands(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	userId := db.GetOrCreateTelegramUserId(123, "", "test")
	sessionId, _, _ := db.CreateSession(userId)

	// check that there were no messages

	db.AddRecentlySentCommand(sessionId, "command1", 10)
	db.AddRecentlySentCommand(sessionId, "command2", 10)
	db.AddRecentlySentCommand(sessionId, "command3", 10)

	{
		commands, newLastIndex := db.GetNewRecentlySentCommands(sessionId, -1)
		assert.Equal(3, len(commands))
		assert.Equal(2, newLastIndex)
		assert.Equal("command1", commands[0])
		assert.Equal("command2", commands[1])
		assert.Equal("command3", commands[2])
	}

	{
		commands, newLastIndex := db.GetNewRecentlySentCommands(sessionId, 0)
		assert.Equal(2, len(commands))
		assert.Equal(2, newLastIndex)
		assert.Equal("command2", commands[0])
		assert.Equal("command3", commands[1])
	}

	{
		commands, newLastIndex := db.GetNewRecentlySentCommands(sessionId, 1)
		assert.Equal(1, len(commands))
		assert.Equal(2, newLastIndex)
		assert.Equal("command3", commands[0])
	}

	{
		commands, newLastIndex := db.GetNewRecentlySentCommands(sessionId, 2)
		assert.Equal(0, len(commands))
		assert.Equal(2, newLastIndex)
	}

	db.AddRecentlySentCommand(sessionId, "command4", 2)

	{
		commands, newLastIndex := db.GetNewRecentlySentCommands(sessionId, 0)
		assert.Equal(2, len(commands))
		assert.Equal(3, newLastIndex)
		assert.Equal("command3", commands[0])
		assert.Equal("command4", commands[1])
	}

	db.LeaveSession(userId)

	{
		commands, newLastIndex := db.GetNewRecentlySentCommands(sessionId, 0)
		assert.Equal(0, len(commands))
		assert.Equal(0, newLastIndex)
	}
}
