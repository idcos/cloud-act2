//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package cmd

import (
	"bytes"
	"encoding/json"
	"idcos.io/cloud-act2/config"
	"path/filepath"
)

type SaltState struct {
}

func NewSaltState() *SaltState {
	return &SaltState{
	}
}

func (sls *SaltState) Render(template string, context map[string]interface{}) (*bytes.Buffer, error) {
	stdinMap := map[string]interface{}{
		"template": template,
		"context":  context,
	}

	stdinBytes, err := json.Marshal(stdinMap)
	if err != nil {
		return nil, err
	}

	cmd := config.Conf.Salt.Python
	arg := filepath.Join(config.Conf.ProjectPath, "scripts/salt-state-render.py")

	executor := NewExecutor(cmd, arg)
	executor.Stdin = bytes.NewBuffer(stdinBytes)
	r := executor.Invoke()

	return r.Stdout, nil
}
