package env

import "os"

func IsProd() bool {
	return os.Getenv("GO_ENV") == "production"
}
