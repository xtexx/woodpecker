// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"context"
	"html/template"
	"os"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var registryListCmd = &cli.Command{
	Name:   "ls",
	Usage:  "list registries",
	Action: registryList,
	Flags: []cli.Flag{
		common.FormatFlag(tmplRegistryList, true),
	},
}

func registryList(ctx context.Context, c *cli.Command) error {
	format := c.String("format") + "\n"

	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	opt := woodpecker.RegistryListOptions{}

	list, err := client.GlobalRegistryList(opt)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	for _, registry := range list {
		if err := tmpl.Execute(os.Stdout, registry); err != nil {
			return err
		}
	}
	return nil
}

// Template for registry list information.
var tmplRegistryList = "\x1b[33m{{ .Address }} \x1b[0m" + `
Username: {{ .Username }}
Email: {{ .Email }}
`
