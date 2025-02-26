package store

// func (b *Builder) ListStoredFilesHandler(cfg *UploaderConfig) HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		err := ValidateRequestMethod(r, http.MethodGet)
// 		if err != nil {
// 			SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
// 			return
// 		}

// 		files, err := b.Store.ListFiles()
// 		if err != nil {
// 			log.Error().Err(err).Msg("Error deleting file")
// 		}

// 		SendJsonResponse(w, http.StatusOK, files, "StoredFiles")
// 	}
// }
