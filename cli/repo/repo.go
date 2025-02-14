// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/repo/cron"
	"go.woodpecker-ci.org/woodpecker/v3/cli/repo/registry"
	"go.woodpecker-ci.org/woodpecker/v3/cli/repo/secret"
)

// Command exports the repository command.
var Command = &cli.Command{
	Name:  "repo",
	Usage: "manage repositories",
	Commands: []*cli.Command{
		repoAddCmd,
		repoChownCmd,
		cron.Command,
		repoListCmd,
		registry.Command,
		repoRemoveCmd,
		repoRepairCmd,
		secret.Command,
		repoShowCmd,
		repoSyncCmd,
		repoUpdateCmd,
	},
}
