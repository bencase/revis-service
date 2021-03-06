// +build darwin

package config

import (
	"os"

	"github.com/mitchellh/go-homedir"
)

const bundleIdentifier = "com.github.bencase.revis"

func init() {
	dir, err := homedir.Dir()
	if err != nil {
		logger.Error("Error trying to find home directory:", err.Error())
		return
	}
	libraryPathWithoutTrailingSlash := dir + "/Library/Containers/" + bundleIdentifier + "/Data/Documents"
	LibraryPath = libraryPathWithoutTrailingSlash + "/"
	// If this directory doesn't exist, create it
	_, err = os.Stat(libraryPathWithoutTrailingSlash)
	if os.IsNotExist(err) {
		err = os.MkdirAll(libraryPathWithoutTrailingSlash, os.ModePerm)
		if err != nil {
			logger.Error("Error creating directory path:", err.Error())
			return
		}
	}
}