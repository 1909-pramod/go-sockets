package main

type UserRequests struct {
	UserConnections    map[string][]*User
	ConnectionRequests map[string][]*User
}

func FindUser(users []*User, userId string) (int, *User) {
	for ind, user := range users {
		if user.Id == userId {
			return ind, user
		}
	}
	return -1, nil
}

func (requests *UserRequests) CheckConnectionExists(Id1 string, Id2 string) bool {
	_, exists1 := requests.UserConnections[Id1]
	_, exists2 := requests.UserConnections[Id2]
	if exists1 && exists2 {
		return true
	}
	if exists1 {
		delete(requests.UserConnections, Id1)
	}
	if exists2 {
		delete(requests.UserConnections, Id2)
	}
	return false
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
