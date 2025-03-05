package utils

import "strings"

func ValidateContentType(contentType string, supportedMimeTypes []string) (bool, error) {

	contentType = strings.Split(contentType, ";")[0]

	for _, supportedType := range supportedMimeTypes {
		if supportedType == "*" || supportedType == "*/*" {
			return true, nil
		}

		if strings.HasSuffix(supportedType, "/*") {
			// Check if the MIME type matches the prefix (e.g., "image/*" matches "image/png")
			prefix := strings.TrimSuffix(supportedType, "/*")
			if strings.HasPrefix(contentType, prefix) {
				return true, nil
			}
		} else if contentType == supportedType {
			return true, nil
		}
	}
	return false, nil
}
