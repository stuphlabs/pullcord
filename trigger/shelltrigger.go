package trigger

import (
	"bytes"
	"fmt"
	// "github.com/stuphlabs/pullcord"
	"os/exec"
)

// ShellTriggerHandler is a basic TriggerHandler that calls a stored shell
// command (along with arguments) when triggered.
//
// The message given to TriggerString will be passed to the command via stdin.
type ShellTriggerHandler struct {
	command string
	args []string
}

// TriggerString for the ShellTriggerHandler is an implementation of the
// TriggerString function required by all TriggerHandler instances.
//
// In this case, the message will be passed to the command via stdin.
func (handler *ShellTriggerHandler) Trigger() (err error) {
	log().Debug("shelltrigger running trigger")
	cmd := exec.Command(handler.command, handler.args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err = cmd.Run()
	log().Debug(
		fmt.Sprintf(
			"shelltrigger command wrote to stdout: %s",
			stdout.String(),
		),
	)
	if err != nil {
		log().Err(
			fmt.Sprintf(
				"shelltrigger failed during trigger: %v",
				err,
			),
		)
		return err
	} else {
		log().Info("shelltrigger trigger sent")
		return nil
	}
}

// NewShellTriggerHandler constructs a new ShellTriggerHandler given the
// command (and arguments) to be run each time TriggerString is called. Entire
// shell scripts could potentially be stored in the arguments, though the
// trigger could just as easily call an external shell script. As a result, a
// wide variety of actions could be taken based on the message passed in via
// stdin.
func NewShellTriggerHandler(
	command string,
	args []string,
) *ShellTriggerHandler {
	log().Info("initializing shell trigger handler")
	log().Debug(
		fmt.Sprintf(
			"shelltrigger will run command: %s %v",
			command,
			args,
		),
	)

	var handler ShellTriggerHandler
	handler.command = command
	handler.args = args

	return &handler
}
