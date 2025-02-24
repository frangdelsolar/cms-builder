package requestlogger

import "net/http"

func RequestLogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := ValidateRequestMethod(r, http.MethodGet)
		if err != nil {
			SendJsonResponse(w, http.StatusMethodNotAllowed, nil, err.Error())
			return
		}

		a, err := b.Admin.GetApp("requestlog")
		if err != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, err.Error())
			return
		}

		requestId := GetUrlParam("id", r)

		// Create instance for RequestLog
		requestLog := RequestLog{}

		query := "request_identifier = ?" // Use parameterized query
		res := b.DB.DB.Where(query, requestId).First(&requestLog)
		if res.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, res.Error.Error())
			return
		}

		// Create slice to store HistoryEntries
		var historyEntries []HistoryEntry // Assuming you have a HistoryEntry struct

		// Join HistoryEntries with RequestLog
		joinQuery := "history_entries.request_id = ?" // Use parameterized query
		historyRes := b.DB.DB.Where(joinQuery, requestId).Find(&historyEntries)
		if historyRes.Error != nil {
			SendJsonResponse(w, http.StatusInternalServerError, nil, historyRes.Error.Error())
			return
		}

		// Create a map to hold both RequestLog and HistoryEntries
		data := map[string]interface{}{
			"request_log":     requestLog,
			"history_entries": historyEntries,
		}

		SendJsonResponse(w, http.StatusOK, data, a.Name()+" details")
	}
}
