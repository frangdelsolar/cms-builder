package handlers

// func (b *Builder) ListStoredFilesHandler(cfg *UploaderConfig) HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		err := svrUtils.ValidateRequestMethod(r, http.MethodGet)
// 		if err != nil {
// 			svrUtils.SendJsonResponse(w, http.StatusMethodNotAllowed, err, err.Error())
// 			return
// 		}

// 		files, err := b.Store.ListFiles()
// 		if err != nil {
// 			log.Error().Err(err).Msg("Error deleting file")
// 		}

// 		svrUtils.SendJsonResponse(w, http.StatusOK, files, "StoredFiles")
// 	}
// }
