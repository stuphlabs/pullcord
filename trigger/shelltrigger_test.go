package trigger

import (
	"github.com/stretchr/testify/assert"
	// "github.com/stuphlabs/pullcord"
	"io/ioutil"
	"os"
	"testing"
)

func goRemoveAll(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		panic(err)
	}
}

func TestShellTriggerTee(t *testing.T) {
	testCommand := "/bin/sh"
	testMessage := "foo"
	testFileEnding := "testShellTrigger.txt"

	tmpdir, err := ioutil.TempDir("/tmp", "test_shell_trigger")
	defer goRemoveAll(tmpdir)
	assert.NoError(t, err)
	testFile := tmpdir + "/" + testFileEnding
	testArgs := []string{
		"-c",
		`printf "` + testMessage + `" > ` + testFile,
	}

	handler := NewShellTriggerHandler(testCommand, testArgs)
	err = handler.Trigger()

	assert.NoError(t, err)
	data, err := ioutil.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, testMessage, string(data))
}

func TestShellTriggerFail(t *testing.T) {
	testCommand := "["
	testArgs := []string{"1", "-eq", "0"}

	handler := NewShellTriggerHandler(testCommand, testArgs)
	err := handler.Trigger()

	assert.Error(t, err)
}

