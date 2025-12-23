package orchestrator

// ConfigKeys define the keys used in the configuration file
type ConfigKeys struct {
	AppName            string `json:"appName"`            // App name
	AdminName          string `json:"adminName"`          // Admin name
	AdminEmail         string `json:"adminEmail"`         // Admin email
	AdminPassword      string `json:"adminPassword"`      // Admin password
	CorsAllowedOrigins string `json:"corsAllowedOrigins"` // CORS allowed origins
	Environment        string `json:"environment"`        // Environment where the app is running
	LogLevel           string `json:"logLevel"`           // Log level
	LogFilePath        string `json:"logFilePath"`        // File path for logging
	LogWriteToFile     string `json:"logWriteToFile"`     // Write logs to file
	Domain             string `json:"domain"`             // Domain
	DbDriver           string `json:"dbDriver"`           // Database driver
	DbFile             string `json:"dbFile"`             // Database file
	DbUrl              string `json:"dbUrl"`              // Database URL
	ServerHost         string `json:"serverHost"`         // Server host
	ServerPort         string `json:"serverPort"`
	SMTPHost           string `json:"smtpHost"`
	SMTPPort           string `json:"smtpPort"`
	SMTPUser           string `json:"smtpUser"`
	SMTPPassword       string `json:"smtpPassword"`
	SMTPSender         string `json:"smtpSender"`         // Server port
	CsrfToken          string `json:"csrfToken"`          // CSRF token
	FirebaseSecret     string `json:"firebaseSecret"`     // Firebase secret
	FirebaseApiKey     string `json:"firebaseApiKey"`     // Firebase API key
	GodToken           string `json:"godToken"`           // God token
	StoreMaxSize       string `json:"storeMaxSize"`       // Uploader max size in MB
	StoreSupportedMime string `json:"storeSupportedMime"` // Supported mime types for uploaded files
	StoreType          string `json:"storeType"`          // Uploader store type
	AwsBucket          string `json:"awsBucket"`          // AWS bucket
	AwsEndpoint        string `json:"awsEndpoint"`        // AWS endpoint
	AwsRegion          string `json:"awsRegion"`          // AWS region
	AwsSecretAccessKey string `json:"awsSecretAccessKey"` // AWS secret access key
	AwsAccessKeyId     string `json:"awsAccessKeyId"`     // AWS access key id
	BaseUrl            string `json:"baseUrl"`            // where the app is running
	RunScheduler       string `json:"runScheduler"`       // Run scheduler
}

// EnvKeys are the keys used in the configuration file
var EnvKeys = ConfigKeys{
	AppName:            "APP_NAME",
	AdminName:          "ADMIN_NAME",
	AdminEmail:         "ADMIN_EMAIL",
	AdminPassword:      "ADMIN_PASSWORD",
	CorsAllowedOrigins: "CORS_ALLOWED_ORIGINS",
	Environment:        "ENVIRONMENT",
	LogLevel:           "LOG_LEVEL",
	LogFilePath:        "LOG_FILE_PATH",
	LogWriteToFile:     "LOG_WRITE_TO_FILE",
	Domain:             "DOMAIN",
	DbDriver:           "DB_DRIVER",
	DbFile:             "DB_FILE",
	DbUrl:              "DB_URL",
	GodToken:           "GOD_TOKEN",
	ServerHost:         "SERVER_HOST",
	ServerPort:         "SERVER_PORT",
	CsrfToken:          "CSRF_TOKEN",
	FirebaseSecret:     "FIREBASE_SECRET",
	FirebaseApiKey:     "FIREBASE_API_KEY",
	SMTPHost:           "SMTP_HOST",
	SMTPPort:           "SMTP_PORT",
	SMTPUser:           "SMTP_USER",
	SMTPPassword:       "SMTP_PASSWORD",
	SMTPSender:         "SMTP_SENDER",
	StoreMaxSize:       "STORE_MAX_SIZE",
	StoreSupportedMime: "STORE_SUPPORTED_MIME_TYPES",
	StoreType:          "STORE_TYPE",
	AwsBucket:          "AWS_BUCKET",
	AwsRegion:          "AWS_REGION",
	AwsEndpoint:        "AWS_ENDPOINT",
	AwsSecretAccessKey: "AWS_SECRET_ACCESS_KEY",
	AwsAccessKeyId:     "AWS_ACCESS_KEY_ID",
	BaseUrl:            "BASE_URL",
	RunScheduler:       "RUN_SCHEDULER",
}
