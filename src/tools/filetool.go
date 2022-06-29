package tools

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"pvp_spiderfly/src/logger"
)

func WriteFile(finalPath string, content []byte) {
	if err := os.MkdirAll(filepath.Dir(finalPath), os.ModePerm); err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"Path": finalPath,
		}).Info("mkdir error")
	}

	if len(content) != 0 {
		logger.Logger.WithFields(logrus.Fields{
			"Path": finalPath,
		}).Info("Save url result")

		if err := ioutil.WriteFile(finalPath, content, os.ModePerm); err != nil {
			logger.Logger.Error(err)
		}
	}
}

func ExistFile(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
