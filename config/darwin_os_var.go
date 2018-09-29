// +build darwin

package config

import (
	"github.com/mitchellh/go-homedir"
)

func init() {
	dir, err := homedir.Dir()
	if err != nil {
		logger.Error("Error trying to find home directory:", err.Error())
		return
	}
	LibraryPath = dir + "/Library/Containers/com.electron.revis/Data/Documents/"
}