package main

import "log"

type UserRequests struct {
	UserConnections    map[string][]*User
	ConnectionRequests map[string][]*User
}

func InitUserRequests() *UserRequests {
	return &UserRequests{
		UserConnections:    make(map[string][]*User),
		ConnectionRequests: make(map[string][]*User),
	}
}

func FindUser(users []*User, userId string) (int, *User) {
	for ind, user := range users {
		if user.Id == userId {
			return ind, user
		}
	}
	return -1, nil
}

func RemoveFromUserSlice(users []*User, Id string) []*User {
	ind, _ := FindUser(users, Id)
	if ind != -1 {
		return append(users[:ind], users[ind+1:]...)
	}
	return users
}

func (requests *UserRequests) CheckConnectionExists(Id1 string, Id2 string) bool {
	log.Printf("check connections %v %v", Id1, Id2)
	users1, exists1 := requests.UserConnections[Id1]
	users2, exists2 := requests.UserConnections[Id2]
	log.Printf("Checking connection %v, %v", exists1, exists2)
	if exists1 && exists2 {
		return true
	}
	if exists1 {
		requests.UserConnections[Id1] = RemoveFromUserSlice(users1, Id2)
	}
	if exists2 {
		requests.UserConnections[Id2] = RemoveFromUserSlice(users2, Id1)
	}
	return false
}

func (requests *UserRequests) CreateConnections(fromId string, toId string) {
	log.Printf("%v %v", fromId, toId)
	users, exists := requests.ConnectionRequests[toId]
	if exists {
		requests.ConnectionRequests[toId] = RemoveFromUserSlice(users, fromId)
	}
	from, fromExists := requests.UserConnections[fromId]
	if fromExists {
		requests.UserConnections[fromId] = RemoveFromUserSlice(from, toId)
		requests.UserConnections[fromId] = append(requests.UserConnections[fromId], &User{
			Id: toId,
		})
	} else {
		fromUser := &User{
			Id: toId,
		}
		requests.UserConnections[fromId] = []*User{fromUser}
	}
	log.Printf("from user %V \n", requests.UserConnections[fromId])
	to, toExists := requests.UserConnections[toId]
	if toExists {
		requests.UserConnections[toId] = RemoveFromUserSlice(to, fromId)
		requests.UserConnections[toId] = append(requests.UserConnections[toId], &User{
			Id: fromId,
		})
	} else {
		toUser := &User{
			Id: fromId,
		}
		requests.UserConnections[toId] = []*User{toUser}
	}
	log.Printf("to user %V \n", requests.UserConnections[toId])
}

func (requests *UserRequests) GetConnectionRequests(Id string) []*User {
	users, exists := requests.UserConnections[Id]
	if exists {
		return users
	}
	return nil
}

func (requests *UserRequests) CheckRequests(FromId string, ToId string) bool {
	userReqs, exists := requests.ConnectionRequests[ToId]
	if exists {
		_, user := FindUser(userReqs, FromId)
		if user != nil {
			return true
		}
	}
	return false
}

func (requests *UserRequests) RemoveRequest(FromId string, ToId string) {
	userReqs, exists := requests.ConnectionRequests[ToId]
	if exists {
		ind, _ := FindUser(userReqs, FromId)
		requests.ConnectionRequests[ToId] = append(userReqs[:ind], userReqs[ind+1:]...)
	}
}

func (requests *UserRequests) AddRequests(FromId string, ToId string) {
	userReqs, exists := requests.ConnectionRequests[ToId]
	if exists {
		requests.ConnectionRequests[ToId] = append(userReqs, &User{
			Id: FromId,
		})

	} else {
		user := &User{
			Id: FromId,
		}
		requests.ConnectionRequests[ToId] = []*User{user}
	}
}
