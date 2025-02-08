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

func CreateLockFile(lockFileName string) {
	fmt.Println(lockFileName)

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
		fmt.Println("ユーザー名を取得できませんでした")
	} else {
		fmt.Printf("Current user: %s\n", user)
	}

	pid := os.Getpid()

	lockFile := LockFile{ID: id, User: user, Pid: pid, TimeStamp: timeStamp}

	info, err := yaml.Marshal(&lockFile)
	if err != nil {
		fmt.Println(err)
	}

	err = os.WriteFile(lockFileName, []byte(info), 0644)
	if err != nil {
		fmt.Println(err)
	}

}
