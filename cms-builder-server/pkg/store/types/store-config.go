package types

type StoreConfig struct {
	MaxSize            int64
	SupportedMimeTypes []string
	MediaFolder        string // the folder where the files are stored i. e. media/easy-files/
}
