package builder

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name       string `json:"name"`
	Email      string `json:"email"`
	FirebaseId string `json:"firebase_id"`
}

// ID returns the ID of the SystemData as a string.
//
// Returns:
// - string: the ID of the SystemData.
func (u *User) GetIDString() string {
	return fmt.Sprint(u.ID)
}
