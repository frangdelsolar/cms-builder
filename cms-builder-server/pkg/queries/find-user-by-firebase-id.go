package queries

import (
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"
	"gorm.io/gorm"
)

func FindUserByFirebaseId(db *database.Database, firebaseId string, user *models.User) *gorm.DB {
	return db.DB.Where("firebase_id = ?", firebaseId).First(user)
}
