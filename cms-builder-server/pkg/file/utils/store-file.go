package file

import (
	"context"
	"fmt"
	"mime/multipart"

	authModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/auth/models"
	dbQueries "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/queries"
	dbTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/database/types"
	fileModels "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/file/models"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
	rmTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager/types"
	storeTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/store/types"
	utilsPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/utils"
)

// StoreUploadedFile handles the full lifecycle of storing an uploaded file.
//
// This includes:
//   - Saving the raw file to the configured Store (st).
//   - Creating the file metadata record in the database.
//   - Running Resource-level validation before persisting.
//   - Generating and updating the final public download URL.
//   - Rolling back the stored file on any error (ensuring no orphaned files).
//
// Parameters:
//
//	ctx         - Request context
//	log         - Request-scoped logger
//	db          - Database connection
//	st          - Implemented Store (local, S3, etc.)
//	apiBaseUrl  - Base URL used to construct the final file download URL
//	resource    - Resource definition used for validation and metadata handling
//	file        - multipart.File representing the uploaded file
//	header      - multipart.FileHeader containing metadata such as the file name
//	user        - Authenticated user performing the upload
//	requestId   - Request identifier for logging and auditing
//
// Returns:
//   - *fileModels.File: The fully created file object, after DB persistence.
//   - error: Non-nil if any part of the storage or DB operations failed.
//
// Side Effects:
//   - If any step fails, the function will delete the partially stored file
//     from the underlying storage engine.
//   - Database records are only created if validation succeeds.
//   - The file metadata is updated twice: once for creation, once for URL update.
func StoreUploadedFile(
	ctx context.Context,
	log *loggerTypes.Logger,
	db *dbTypes.DatabaseConnection,
	st storeTypes.Store,
	apiBaseUrl string,
	resource *rmTypes.Resource,
	file multipart.File,
	header *multipart.FileHeader,
	user *authModels.User,
	requestId string,
) (*fileModels.File, error) {

	// Store file in configured Store
	fileData, err := st.StoreFile(header.Filename, file, header, log)
	if err != nil {
		// Cleanup partially saved file
		st.DeleteFile(fileData, log)
		return nil, err
	}

	// Attach system metadata
	fileData.SystemData = &authModels.SystemData{
		CreatedByID: user.ID,
		UpdatedByID: user.ID,
	}

	// Run Resource validations
	validation := resource.Validate(fileData, log)
	if len(validation.Errors) > 0 {
		st.DeleteFile(fileData, log)
		return nil, fmt.Errorf("validation failed: %v", validation.Errors)
	}

	// Insert DB record
	if err := dbQueries.Create(ctx, log, db, fileData, user, requestId); err != nil {
		st.DeleteFile(fileData, log)
		return nil, err
	}

	// Update public URL and persist diff
	before := *fileData
	fileData.Url = apiBaseUrl + "/private/api/files/" + fileData.StringID() + "/download"

	diff := utilsPkg.CompareInterfaces(&before, fileData)
	if err := dbQueries.Update(ctx, log, db, fileData, user, diff, requestId); err != nil {
		st.DeleteFile(fileData, log)
		return nil, err
	}

	return fileData, nil
}
