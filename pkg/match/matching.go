package match

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Double-DOS/go-socket-chat/pkg/payment"
	"log"

	"github.com/Double-DOS/go-socket-chat/db"
	"github.com/Double-DOS/randommer-go"
)

var CurrMaxGroup int

type UserInfo struct {
	ID          int         `json:"id" db:"id"`
	FirstName   string      `json:"firstName" db:"first_name"`
	LastName    string      `json:"lastName" db:"last_name"`
	PhoneNumber string      `json:"phoneNumber" db:"phone_number"`
	Email       string      `json:"email" db:"email"`
	Gender      string      `json:"gender" db:"gender"`
	RandomName  string      `json:"randomName" db:"random_name"`
	Matched     bool        `json:"matched"  db:"matched"`
	Matches     []*UserInfo `json:"matches"`
	MatchCount  int         `json:"matchCount" db:"match_count"`
	MatchedTo   *UserInfo   `json:"matchedTo"`
	IsPaid      bool        `json:"isPaid" db:"ispaid"`
	PaymentUrl  string      `json:"paymentUrl,omitempty" db:"-"`
}

type UserInfoDto struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
	Gender      string `json:"gender"`
}

func (uiDto *UserInfoDto) NewUserInfo() (*UserInfo, bool, error) {
	var newUser *UserInfo
	existingUser, _ := FindUserByEmail(uiDto.Email)

	if existingUser != nil {
		// If the user is unpaid, IsPaid is set to false
		if !existingUser.IsPaid {
			paymentUrl, err := payment.InitializePayment(db.DB, existingUser.ID, existingUser.Email, map[string]interface{}{})
			if err != nil {
				return nil, false, err
			}
			existingUser.PaymentUrl = paymentUrl
			return existingUser, false, nil
		}

		// Check specific conditions for existing users
		if existingUser.Gender == "F" && existingUser.MatchCount > 0 {
			return existingUser, false, nil
		} else if existingUser.Gender == "M" && existingUser.MatchedTo != nil {
			return existingUser, false, nil
		} else {
			newUser = existingUser
		}
	} else {
		// Create a new user if not found
		randomNames := randommer.GetRandomNames("firstname", 1)
		if len(randomNames) == 0 {
			return nil, false, fmt.Errorf("failed to generate random name")
		}
		randomName := randomNames[0]

		// Create the new user in the database
		createdUser, err := CreateUser(*uiDto, string(randomName))
		if err != nil {
			return nil, false, err
		}
		newUser = createdUser
		paymentUrl, err := payment.InitializePayment(db.DB, newUser.ID, newUser.Email, map[string]interface{}{})
		if err != nil {
			return nil, false, err
		}
		newUser.PaymentUrl = paymentUrl
	}

	if newUser.IsPaid {
		if newUser.Gender == "F" {
			// Female-specific matching logic
			unMatchedMales := FindUnMatchedMales()
			if len(unMatchedMales) > 0 {
				for i := 0; i < CurrMaxGroup+1 && i < len(unMatchedMales); i++ {
					male := unMatchedMales[i]
					AddNewMaleToFemale(male, newUser)
				}
			}
		} else if newUser.Gender == "M" {
			// Male-specific matching logic
			lowestMatchFemale := FindFemaleWithLowestMatch()
			if lowestMatchFemale != nil {
				AddNewMaleToFemale(newUser, lowestMatchFemale)
			}
		}
	}

	return newUser, true, nil
}

func MatchUser(user *UserInfo) error {
	if user.IsPaid {
		if user.Gender == "F" {
			// Female-specific matching logic
			unMatchedMales := FindUnMatchedMales()
			if len(unMatchedMales) > 0 {
				for i := 0; i < CurrMaxGroup+1 && i < len(unMatchedMales); i++ {
					male := unMatchedMales[i]
					AddNewMaleToFemale(male, user)
				}
			}
		} else if user.Gender == "M" {
			// Male-specific matching logic
			lowestMatchFemale := FindFemaleWithLowestMatch()
			if lowestMatchFemale != nil {
				AddNewMaleToFemale(user, lowestMatchFemale)
			}
		}
		return nil
	}

	return errors.New("user is not paid, and cannot match")
}

func AddNewMaleToFemale(male, female *UserInfo) {
	// Begin a transaction
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return
	}

	// Insert match into Matches table
	matchQuery := `
        INSERT INTO Matches (user_id, matched_user_id)
        VALUES (:female_id, :male_id)
    `
	_, err = tx.NamedExec(matchQuery, map[string]interface{}{
		"male_id":   male.ID,
		"female_id": female.ID,
	})
	if err != nil {
		log.Printf("Error inserting match: %v", err)
		tx.Rollback() // Rollback in case of error
		return
	}

	// Update matched status for the male user
	updateMaleQuery := `
        UPDATE Users
        SET matched = TRUE
        WHERE id = :male_id
    `
	_, err = tx.NamedExec(updateMaleQuery, map[string]interface{}{
		"male_id": male.ID,
	})
	if err != nil {
		log.Printf("Error updating male's matched status: %v", err)
		tx.Rollback() // Rollback in case of error
		return
	}

	// Update matched status for the female user
	updateFemaleQuery := `
        UPDATE Users
        SET matched = TRUE
        WHERE id = :female_id
    `
	_, err = tx.NamedExec(updateFemaleQuery, map[string]interface{}{
		"female_id": female.ID,
	})
	if err != nil {
		log.Printf("Error updating female's matched status: %v", err)
		tx.Rollback() // Rollback in case of error
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}

	// Update local user info structs to reflect the new match
	female.Matches = append(female.Matches, male)
	female.Matched = true
	male.Matched = true
}

func FindUnMatchedMales() []*UserInfo {
	query := `
        SELECT * FROM Users 
        WHERE gender = 'M' AND matched = FALSE;
    `

	var users []*UserInfo
	err := db.DB.Select(&users, query)
	if err != nil {
		log.Printf("Error finding unmatched males: %v", err)
		return nil
	}

	return users
}

func FindFemaleWithLowestMatch() *UserInfo {
	query := `
        SELECT Users.*, COUNT(Matches.matched_user_id) as match_count
        FROM Users
        LEFT JOIN Matches ON Users.id = Matches.user_id
        WHERE Users.gender = 'F'
        GROUP BY Users.id
        ORDER BY match_count ASC
        LIMIT 1;
    `

	var user UserInfo
	err := db.DB.QueryRowx(query).StructScan(&user)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No female users found right now")

			return nil
		} else {

			log.Printf("Error finding female with the lowest match: %v", err)
			return nil
		}
	}

	return &user
}

func FindUserByEmail(email string) (*UserInfo, error) {
	// Query to find the user by ID
	user := UserInfo{}
	err := db.DB.Get(&user, "SELECT Users.*, (SELECT COUNT(*) FROM Matches WHERE Matches.user_id = Users.id) AS match_count FROM Users WHERE email = $1", email)
	if err != nil {
		log.Printf("Error finding user with email %s: %v", email, err)
		return nil, err
	}

	if user.Gender == "F" {
		// Query to find all matches for the user
		matches := []*UserInfo{}
		err = db.DB.Select(&matches, "SELECT u.* FROM Users u JOIN Matches m ON u.id = m.matched_user_id WHERE m.user_id = $1", user.ID)
		if err != nil {
			log.Printf("Error finding matches for user with email %s: %v", email, err)
			return &user, err
		}
		user.Matches = matches
		user.MatchCount = len(user.Matches)
	} else if user.Gender == "M" {
		// Query to find the user this user is matched to, if any
		var matchedTo UserInfo
		err = db.DB.Get(&matchedTo, "SELECT u.* FROM Users u JOIN Matches m ON u.id = m.user_id WHERE m.matched_user_id = $1", user.ID)
		if err != nil {
			log.Printf("Error finding matched user for user with ID %d: %v", user.ID, err)
		} else {
			user.MatchedTo = &matchedTo
		}
	}

	return &user, nil
}

func CreateUser(user UserInfoDto, randomName string) (*UserInfo, error) {
	query := `
	INSERT INTO Users (first_name, last_name, phone_number, email, gender, random_name, matched, isPaid)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id
`

	// Using QueryRowx to leverage ordinal parameters in the query
	row := db.DB.QueryRowx(
		query,
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
		user.Email,
		user.Gender,
		randomName, // Assuming 'randomName' is defined elsewhere in your code
		false,      // Assuming this is the value you want to set for 'matched'
		false,      // Assuming this is the value you want to set for 'isPaid'
	)

	var id int
	if err := row.Scan(&id); err != nil {
		log.Printf("Failed to retrieve last insert ID: %v", err)
		return nil, err
	}

	log.Printf("last insert ID: %d", id)

	// Retrieve the newly created user's information
	newUser := UserInfo{}
	err := db.DB.Get(&newUser, "SELECT * FROM Users WHERE id = $1", id)
	if err != nil {
		log.Printf("Error retrieving new user: %v", err)
		return nil, err
	}

	return &newUser, nil
}

func GetUserStats() (map[string]interface{}, error) {
	var allUsers []*UserInfo
	var allFemaleUsers []*UserInfo
	var allMaleUsers []*UserInfo
	allUserQuery := "SELECT u.*, COUNT(m.matched_user_id) as match_count FROM Users u LEFT JOIN Matches m on u.id = m.user_id GROUP BY u.id ORDER BY match_count DESC"
	err := db.DB.Select(&allUsers, allUserQuery)
	if err != nil {
		log.Printf("Error fethcing all users: %v", err)
		return nil, err
	}
	for _, user := range allUsers {
		if user.Gender == "F" {
			allFemaleUsers = append(allFemaleUsers, user)
			// Query to find all matches for the user
			matches := []*UserInfo{}
			err = db.DB.Select(&matches, "SELECT u.* FROM Users u JOIN Matches m ON u.id = m.matched_user_id WHERE m.user_id = $1", user.ID)
			if err != nil {
				log.Printf("Error finding matches for user with email %s: %v", user.Email, err)
			}
			user.Matches = matches
			// user.MatchCount = len(user.Matches)
		} else if user.Gender == "M" {
			allMaleUsers = append(allMaleUsers, user)

			// Query to find the user this user is matched to, if any
			var matchedTo UserInfo
			err = db.DB.Get(&matchedTo, "SELECT u.* FROM Users u JOIN Matches m ON u.id = m.user_id WHERE m.matched_user_id = $1", user.ID)
			if err != nil {
				log.Printf("Error finding matched user for user with ID %d: %v", user.ID, err)
			} else {
				user.MatchedTo = &matchedTo
			}
		}
	}
	return map[string]interface{}{
		"total_registration": len(allUsers),
		"total_females":      allFemaleUsers,
		"total_males":        allMaleUsers,
	}, nil
}
