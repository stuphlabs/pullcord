package trigger

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
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

	handler := NewShellTriggerrer(testCommand, testArgs)
	err = handler.Trigger()

	assert.NoError(t, err)
	data, err := ioutil.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, testMessage, string(data))
}

func TestShellTriggerFail(t *testing.T) {
	testCommand := "["
	testArgs := []string{"1", "-eq", "0"}

	handler := NewShellTriggerrer(testCommand, testArgs)
	err := handler.Trigger()

	assert.Error(t, err)
}

func TestShellTriggerFromConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "shelltrigger",
		SyntacticallyBad: []configutil.ConfigTestData{
			{
				Data:        "",
				Explanation: "empty config",
			},
			{
				Data: `{
					"command": 7,
					"args": []
				}`,
				Explanation: "numeric command",
			},
			{
				Data: `{
					"command": {},
					"args": []
				}`,
				Explanation: "object command",
			},
			{
				Data: `{
					"command": "echo",
					"args": 42
				}`,
				Explanation: "numeric args",
			},
			{
				Data: `{
					"command": "echo",
					"args": "hello"
				}`,
				Explanation: "non-array string args",
			},
			{
				Data: `{
					"command": "echo",
					"args": [
						7,
						42
					]
				}`,
				Explanation: "numeric array args",
			},
			{
				Data:        "42",
				Explanation: "numeric config",
			},
		},
		Good: []configutil.ConfigTestData{
			{
				Data:        "{}",
				Explanation: "empty object",
			},
			{
				Data:        "null",
				Explanation: "null config",
			},
			{
				Data: `{
					"command": "echo"
				}`,
				Explanation: "missing args",
			},
			{
				Data: `{
					"command": "echo",
					"args": [
						"hello",
						"world"
					]
				}`,
				Explanation: "basic valid compound trigger",
			},
		},
	}
	test.Run(t)
}
