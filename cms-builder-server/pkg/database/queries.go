package database

// // FindById retrieves a single record from the database that matches the provided ID.
// // It allows for an optional query extension to refine the search criteria.
// //
// // Parameters:
// //   - id: the unique identifier of the record to be retrieved.
// //   - entity: the destination where the result will be stored.
// //   - queryExtension: an optional additional query condition.
// //
// // Returns:
// //   - *gorm.DB: the result of the database query, which can be used to check for errors.
// func (db *Database) FindById(id string, entity interface{}, queryExtension string) *gorm.DB {
// 	q := "id = '" + id + "'"

// 	if queryExtension != "" {
// 		q += " AND " + queryExtension
// 	}

// 	return db.DB.Where(q).First(entity)
// }

// // FindUserByFirebaseId retrieves a user from the database by its Firebase ID.
// //
// // Parameters:
// //   - firebaseId: the Firebase ID of the user to be retrieved.
// //   - entity: the destination where the result will be stored.
// //
// // Returns:
// //   - *gorm.DB: the result of the database query, which can be used to check for errors.
// func (db *Database) FindUserByFirebaseId(firebaseId string, user *User) *gorm.DB {
// 	return db.DB.Where("firebase_id = ?", firebaseId).First(user)
// }

// // FindOne retrieves a single record from the database that matches the provided query.
// //
// // Parameters:
// //   - entity: the destination where the result will be stored.
// //   - query: the query to be executed, it can be a raw SQL query or a GORM query.
// //
// // Returns:
// //   - *gorm.DB: the result of the database query, which can be used to check for errors.
// func (db *Database) FindOne(entity interface{}, query string) *gorm.DB {
// 	return db.DB.Where(query).First(entity)
// }

// // Delete deletes the record in the database.
// //
// // Parameters:
// //   - entity: the model instance to be deleted.
// //
// // Returns:
// //   - *gorm.DB: the result of the database query, which can be used to check for errors.
// func (db *Database) Delete(entity interface{}, user *User, requestId string) *gorm.DB {

// 	result := db.DB.Delete(entity)
// 	if result.Error == nil {
// 		historyEntry, err := NewLogHistoryEntry(DeleteCRUDAction, user, entity, "", requestId)
// 		if err != nil {
// 			return nil
// 		}
// 		_ = db.DB.Create(historyEntry)
// 	}

// 	return result
// }

// // Save updates a record in the database if it already exists, or creates a new one if it does not.
// //
// // Parameters:
// //   - entity: the model instance to be saved.
// //
// // Returns:
// //   - *gorm.DB: the result of the database query, which can be used to check for errors.
// func (db *Database) Save(entity interface{}, user *User, differences interface{}, requestId string) *gorm.DB {

// 	result := db.DB.Save(entity)
// 	if result.Error == nil {
// 		historyEntry, err := NewLogHistoryEntry(UpdateCRUDAction, user, entity, differences, requestId)
// 		if err != nil {
// 			return db.DB
// 		}
// 		_ = db.DB.Create(historyEntry)
// 	}

// 	return result
// }

// // Migrate calls the AutoMigrate method on the GORM DB instance.
// func (db *Database) Migrate(model interface{}) error {
// 	if db == nil {
// 		return ErrDBNotInitialized
// 	}
// 	db.DB.AutoMigrate(model)
// 	return nil
// }
