package orchestrator

// ConfigKeys define the keys used in the configuration file
type ConfigKeys struct {
	AppName               string `json:"appName"`               // App name
	AdminName             string `json:"adminName"`             // Admin name
	AdminEmail            string `json:"adminEmail"`            // Admin email
	AdminPassword         string `json:"adminPassword"`         // Admin password
	CorsAllowedOrigins    string `json:"corsAllowedOrigins"`    // CORS allowed origins
	Environment           string `json:"environment"`           // Environment where the app is running
	LogLevel              string `json:"logLevel"`              // Log level
	LogFilePath           string `json:"logFilePath"`           // File path for logging
	LogWriteToFile        string `json:"logWriteToFile"`        // Write logs to file
	Domain                string `json:"domain"`                // Domain
	DbDriver              string `json:"dbDriver"`              // Database driver
	DbFile                string `json:"dbFile"`                // Database file
	DbUrl                 string `json:"dbUrl"`                 // Database URL
	ServerHost            string `json:"serverHost"`            // Server host
	ServerPort            string `json:"serverPort"`            // Server port
	CsrfToken             string `json:"csrfToken"`             // CSRF token
	FirebaseSecret        string `json:"firebaseSecret"`        // Firebase secret
	FirebaseApiKey        string `json:"firebaseApiKey"`        // Firebase API key
	GodToken              string `json:"godToken"`              // God token
	UploaderMaxSize       string `json:"uploaderMaxSize"`       // Uploader max size in MB
	UploaderAuthenticate  string `json:"uploaderAuthenticate"`  // Whether files will be public or private accessible
	UploaderSupportedMime string `json:"uploaderSupportedMime"` // Supported mime types for uploaded files
	UploaderFolder        string `json:"uploaderFolder"`        // Uploader folder
	StoreType             string `json:"storeType"`             // Uploader store type
	AwsBucket             string `json:"awsBucket"`             // AWS bucket
	AwsRegion             string `json:"awsRegion"`             // AWS region
	AwsSecretAccessKey    string `json:"awsSecretAccessKey"`    // AWS secret access key
	AwsAccessKeyId        string `json:"awsAccessKeyId"`        // AWS access key id
	BaseUrl               string `json:"baseUrl"`               // where the app is running
}

// EnvKeys are the keys used in the configuration file
var EnvKeys = ConfigKeys{
	AppName:               "APP_NAME",
	AdminName:             "ADMIN_NAME",
	AdminEmail:            "ADMIN_EMAIL",
	AdminPassword:         "ADMIN_PASSWORD",
	CorsAllowedOrigins:    "CORS_ALLOWED_ORIGINS",
	Environment:           "ENVIRONMENT",
	LogLevel:              "LOG_LEVEL",
	LogFilePath:           "LOG_FILE_PATH",
	LogWriteToFile:        "LOG_WRITE_TO_FILE",
	Domain:                "DOMAIN",
	DbDriver:              "DB_DRIVER",
	DbFile:                "DB_FILE",
	DbUrl:                 "DB_URL",
	GodToken:              "GOD_TOKEN",
	ServerHost:            "SERVER_HOST",
	ServerPort:            "SERVER_PORT",
	CsrfToken:             "CSRF_TOKEN",
	FirebaseSecret:        "FIREBASE_SECRET",
	FirebaseApiKey:        "FIREBASE_API_KEY",
	UploaderMaxSize:       "UPLOADER_MAX_SIZE",
	UploaderAuthenticate:  "UPLOADER_AUTHENTICATE",
	UploaderSupportedMime: "UPLOADER_SUPPORTED_MIME_TYPES",
	UploaderFolder:        "UPLOADER_FOLDER",
	StoreType:             "STORE_TYPE",
	AwsBucket:             "AWS_BUCKET",
	AwsRegion:             "AWS_REGION",
	AwsSecretAccessKey:    "AWS_SECRET_ACCESS_KEY",
	AwsAccessKeyId:        "AWS_ACCESS_KEY_ID",
	BaseUrl:               "BASE_URL",
}
