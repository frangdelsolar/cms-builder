-- type User struct {
-- 	*gorm.Model
-- 	ID         uint   `gorm:"primaryKey" json:"ID"`
-- 	Name       string `json:"name"`
-- 	Email      string `gorm:"unique" json:"email"`
-- 	FirebaseId string `json:"firebaseId"`
-- 	Roles      string `json:"roles"`
-- }

CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE, 
    name TEXT,
    email TEXT UNIQUE,  
    firebase_id TEXT,
    roles TEXT
);

CREATE INDEX idx_users_firebase_id ON users (firebase_id); 