package presence

import (
	"strings"
	"errors"
	"time"
	
	"golang.org/x/net/context"

	api "git.circuitco.de/self/greyhouse/api"
)

type Person struct {
	Id int32
	Name string
	Location api.Location
	Room api.Room // probably not known for the longest time, so most nodes won't have any reason to use this
	LastSeen time.Time
}

type PersonService struct {
	People map[int32]Person
}

func NewPersonService() PersonService {
	return PersonService{map[int32]Person{}}
}

func (ps PersonService) GetPerson(name string) *Person {
	for _, person := range ps.People {
		if strings.Compare(name, person.Name) == 0 {
			return &person
		}
	}
	return nil
}

func (ps PersonService) Find(ctx context.Context, find *api.FindPerson) (*api.PersonId, error) {
	person := ps.GetPerson(find.Name)
	if person != nil {
		return &api.PersonId{Id: person.Id}, nil
	} else {
		return nil, errors.New("Person not found")
	}
}
