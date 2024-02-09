package match

import (
	"github.com/Double-DOS/randommer-go"
)

var CurrMaxGroup int

type UserInfo struct {
	ID         int         `json:"id"`
	FirstName  string      `json:"firstName"`
	LastName   string      `json:"lastName"`
	Email      string      `json:"email"`
	Gender     string      `json:"gender"`
	RandomName string      `json:"randomName"`
	Matched    bool        `json:"matched"`
	Matches    []*UserInfo `json:"matches"`
	MatchedTo  *UserInfo   `json:"matchedTo"`
}

type UserInfoDto struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Gender    string `json:"gender"`
}

func (uiDto *UserInfoDto) NewUserInfo() *UserInfo {
	// generate random name here
	randomName := randommer.GetRandomNames("firstname", 1)[0]
	// todo: run query against the database to create a new user
	newUser := &UserInfo{
		// todo:  add DB id
		FirstName:  uiDto.FirstName,
		LastName:   uiDto.LastName,
		Email:      uiDto.Email,
		Gender:     uiDto.Gender,
		RandomName: string(randomName),
	}
	// todo: if a user is a female;
	if newUser.Gender == "F" {
		// check if there are unmatched males.
		unMatchedMales := FindUnMatchedMales()
		// if there are, then attach enough males to female based on the current max match + 1

		if len(unMatchedMales) > 0 {
			for i := 0; i < CurrMaxGroup+1; i++ {
				male := unMatchedMales[i]
				AddNewMaleToFemale(male, newUser)
			}
		}
		// if there are no males unmatched, return
	} else if newUser.Gender == "M" {
		// if user is a male, add to female lowest number of female match.
		lowestMatchFemale := FindFemaleWithLowestMatch()
		if lowestMatchFemale != nil {
			AddNewMaleToFemale(newUser, lowestMatchFemale)
		}

	}

	// if no female in db yet, return
	return newUser
}

func AddNewMaleToFemale(male, female *UserInfo) {
	female.Matches = append(female.Matches, male)
	female.Matched = true
	male.Matched = true
	// update db with new data
}
func FindUnMatchedMales() []*UserInfo {
	// todo: run query
	return []*UserInfo{}
}
func FindFemaleWithLowestMatch() *UserInfo {
	// todo: run query
	return &UserInfo{}

}

func FindUser(id int) *UserInfo {
	// todo: run query
	return &UserInfo{}
}
