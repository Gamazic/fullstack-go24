package internal

import "container/list"

type UserElement struct {
	list.Element
	User
}

type InMemoryStorage struct {
	userMap map[int]list.Element
	nextId  int
	order   *list.List
}

func NewInMemoryStorage() *InMemoryStorage {
	u := make(map[int]list.Element)
	storage := &InMemoryStorage{userMap: u, order: list.New()}
	storage.CreateUser(User{Username: newString("Admin")})
	return storage
}

func (i *InMemoryStorage) GetAll() []User {
	users := []User{}
	// go through list ...
	return users
}

func (i *InMemoryStorage) Delete(id int) {
	el := i.userMap[id]
	i.order.Remove(&el)
	delete(i.userMap, id)
}

func (i *InMemoryStorage) CreateUser(user User) User {
	idOfUser := i.nextId
	user.Id = &idOfUser
	nextEl := list.Element{Value: &user}
	i.order.PushFront(nextEl)
	i.userMap[i.nextId] = nextEl
	i.nextId += 1
	return user
}
