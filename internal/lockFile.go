package internal

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type LockFile struct {
	ID        string `yaml:"id"`
	User      string `yaml:"user"`
	Pid       int    `yaml:"pid"`
	TimeStamp string `yaml:"timestamp"`
}

func CreateLockFile(lockFileName string) error {
	fmt.Println("Creating lock file:", lockFileName)

	t := time.Now()
	id := fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	timeStamp := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		return fmt.Errorf("failed to retrieve the username")
	}

	pid := os.Getpid()

	lockFile := LockFile{ID: id, User: user, Pid: pid, TimeStamp: timeStamp}

	info, err := yaml.Marshal(&lockFile)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	err = os.WriteFile(lockFileName, info, 0644)
	if err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}
