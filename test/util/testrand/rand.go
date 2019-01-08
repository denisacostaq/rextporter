package testrand

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// testingFolder return a default testing folder root
func testingFolder() (testingFolder string, err error) {
	testingFolder = filepath.Join(os.TempDir(), "testingFolder")
	if err = os.MkdirAll(testingFolder, 0750); err != nil {
		log.WithError(err).Errorln("creating testing folder")
		return testingFolder, err
	}
	return testingFolder, err
}

// dynamicTestingFolder return a default testing folder root with a timestamp as last folder path
func dynamicTestingFolder() (timestampedTestingFolder string, err error) {
	if timestampedTestingFolder, err = testingFolder(); err != nil {
		log.WithError(err).Errorln("getting testing root folder")
		return timestampedTestingFolder, err
	}
	timestamp := strconv.FormatInt(int64(time.Now().Nanosecond()), 10)
	timestampedTestingFolder = filepath.Join(timestampedTestingFolder, timestamp)
	if err = os.MkdirAll(timestampedTestingFolder, 0750); err != nil {
		log.WithError(err).Errorln("creating dynamic testing folder")
		return timestampedTestingFolder, err
	}
	return timestampedTestingFolder, err
}

// FilePathToSharePort path in which you should write/read the port number where fake server is listinning
func FilePathToSharePort() (path string, err error) {
	var testFolder string
	testFolder, err = testingFolder()
	return filepath.Join(testFolder, "listenport.txt"), err
}

// RName return a random string from a predefined list
func RName() string {
	names := []string{"a", "bsfdf", "test", "integration", "integration_test", "fake", "32", "other", "dfdf", "c", "d", "e", "f", "g", "h", "i"}
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	i := r.Intn(len(names) - 1)
	log.WithFields(log.Fields{"nameIndex": i, "name": names[i]}).Infoln("random folder name")
	return names[i]
}

// RFolderPath return a random folder path under a directory from tmp
func RFolderPath() string {
	testFolder, err := dynamicTestingFolder()
	if err != nil {
		return ""
	}
	path := filepath.Join(testFolder, RName())
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	deep := r.Intn(5)
	for deep > 0 {
		path = filepath.Join(path, RName())
		deep--
	}
	log.WithField("path", path).Infoln("random path")
	return path
}

// RandomPort returns a port number from 5000-5100
func RandomPort() uint16 {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	port := 5000 + uint16(r.Intn(100))
	log.WithField("port", port).Infoln("random port number")
	return port
}
