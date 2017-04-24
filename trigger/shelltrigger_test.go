package trigger

import (
	"github.com/stretchr/testify/assert"
	configutil "github.com/stuphlabs/pullcord/config/util"
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

func TestShellTriggerFromConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "shelltrigger",
		SyntacticallyBad: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: "",
				Explanation: "empty config",
			},
			configutil.ConfigTestData{
				Data: `{
					"command": 7,
					"args": []
				}`,
				Explanation: "numeric command",
			},
			configutil.ConfigTestData{
				Data: `{
					"command": {},
					"args": []
				}`,
				Explanation: "object command",
			},
			configutil.ConfigTestData{
				Data: `{
					"command": "echo",
					"args": 42
				}`,
				Explanation: "numeric args",
			},
			configutil.ConfigTestData{
				Data: `{
					"command": "echo",
					"args": "hello"
				}`,
				Explanation: "non-array string args",
			},
			configutil.ConfigTestData{
				Data: `{
					"command": "echo",
					"args": [
						7,
						42
					]
				}`,
				Explanation: "numeric array args",
			},
			configutil.ConfigTestData{
				Data: "42",
				Explanation: "numeric config",
			},
		},
		Good: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: "{}",
				Explanation: "empty object",
			},
			configutil.ConfigTestData{
				Data: "null",
				Explanation: "null config",
			},
			configutil.ConfigTestData{
				Data: `{
					"command": "echo"
				}`,
				Explanation: "missing args",
			},
			configutil.ConfigTestData{
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
