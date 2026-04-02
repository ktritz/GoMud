package users

import "github.com/GoMudEngine/GoMud/internal/connections"

// Test helpers for registering/deregistering users in the global user manager.
// These are used by the test harness package for integration testing.

// RegisterTestUser adds a user to the global user manager without requiring
// a connection or login flow. For testing only.
func RegisterTestUser(user *UserRecord) {
	userManager.Users[user.UserId] = user
	userManager.Usernames[user.Username] = user.UserId
}

// DeregisterTestUser removes a user from the global user manager.
func DeregisterTestUser(userId int) {
	if user, ok := userManager.Users[userId]; ok {
		delete(userManager.Usernames, user.Username)
		delete(userManager.Users, userId)
		if connId, ok := userManager.UserConnections[userId]; ok {
			delete(userManager.Connections, connId)
			delete(userManager.UserConnections, userId)
		}
	}
}

// ClearAllTestUsers removes all users from the global user manager.
func ClearAllTestUsers() {
	userManager.Users = make(map[int]*UserRecord)
	userManager.Usernames = make(map[string]int)
	userManager.Connections = make(map[connections.ConnectionId]int)
	userManager.UserConnections = make(map[int]connections.ConnectionId)
	userManager.ZombieConnections = make(map[connections.ConnectionId]uint64)
}
