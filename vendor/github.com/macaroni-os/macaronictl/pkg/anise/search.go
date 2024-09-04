/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package anise

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/macaroni-os/macaronictl/pkg/logger"
	specs "github.com/macaroni-os/macaronictl/pkg/specs"
	"github.com/macaroni-os/macaronictl/pkg/utils"
)

func SearchStones(args []string) (*specs.StonesPack, error) {
	log := logger.GetDefaultLogger()
	var errBuffer bytes.Buffer
	var outBuffer bytes.Buffer
	var ans specs.StonesPack

	cmd := exec.Command(args[0], args[1:]...)

	log.Debug(fmt.Sprintf("Running search command: %s",
		strings.Join(args, " ")))

	cmd.Stdout = utils.NewNopCloseWriter(&outBuffer)
	cmd.Stderr = utils.NewNopCloseWriter(&errBuffer)

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return nil, fmt.Errorf("anise search exiting with %s: %s",
			cmd.ProcessState.ExitCode(),
			errBuffer.String())
	}

	// Read json output.
	err = json.Unmarshal(outBuffer.Bytes(), &ans)
	if err != nil {
		return nil, fmt.Errorf("Error on unmarshal json data: %s",
			err.Error())
	}

	return &ans, nil
}
