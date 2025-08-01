// Copyright 2022 Woodpecker Authors
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

package exec

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/drone/envsubst"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/lint"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/docker"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/kubernetes"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/local"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/linter"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/matrix"
	pipelineLog "go.woodpecker-ci.org/woodpecker/v3/pipeline/log"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

// Command exports the exec command.
var Command = &cli.Command{
	Name:      "exec",
	Usage:     "execute a local pipeline",
	ArgsUsage: "[path/to/.woodpecker.yaml]",
	Action:    run,
	Flags:     utils.MergeSlices(flags, docker.Flags, kubernetes.Flags, local.Flags),
}

var backends = []backend_types.Backend{
	kubernetes.New(),
	docker.New(),
	local.New(),
}

func run(ctx context.Context, c *cli.Command) error {
	return common.RunPipelineFunc(ctx, c, execFile, execDir)
}

func execDir(ctx context.Context, c *cli.Command, dir string) error {
	// TODO: respect pipeline dependency
	repoPath := c.String("repo-path")
	if repoPath != "" {
		repoPath, _ = filepath.Abs(repoPath)
	} else {
		repoPath, _ = filepath.Abs(filepath.Dir(dir))
	}
	if runtime.GOOS == "windows" {
		repoPath = convertPathForWindows(repoPath)
	}
	// TODO: respect depends_on and do parallel runs with output to multiple _windows_ e.g. tmux like
	return filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			fmt.Println("#", info.Name())
			_ = runExec(ctx, c, path, repoPath, false) // TODO: should we drop errors or store them and report back?
			fmt.Println("")
			return nil
		}

		return nil
	})
}

func execFile(ctx context.Context, c *cli.Command, file string) error {
	repoPath := c.String("repo-path")
	if repoPath != "" {
		repoPath, _ = filepath.Abs(repoPath)
	} else {
		repoPath, _ = filepath.Abs(filepath.Dir(file))
	}
	if runtime.GOOS == "windows" {
		repoPath = convertPathForWindows(repoPath)
	}
	return runExec(ctx, c, file, repoPath, true)
}

func runExec(ctx context.Context, c *cli.Command, file, repoPath string, singleExec bool) error {
	dat, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	axes, err := matrix.ParseString(string(dat))
	if err != nil {
		return fmt.Errorf("parse matrix fail")
	}

	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}
	for _, axis := range axes {
		err := execWithAxis(ctx, c, file, repoPath, axis, singleExec)
		if err != nil {
			return err
		}
	}
	return nil
}

func execWithAxis(ctx context.Context, c *cli.Command, file, repoPath string, axis matrix.Axis, singleExec bool) error {
	metadataWorkflow := &metadata.Workflow{}
	if !singleExec {
		// TODO: proper try to use the engine to generate the same metadata for workflows
		// https://github.com/woodpecker-ci/woodpecker/pull/3967
		metadataWorkflow.Name = strings.TrimSuffix(strings.TrimSuffix(file, ".yaml"), ".yml")
	}
	metadata, err := metadataFromContext(ctx, c, axis, metadataWorkflow)
	if err != nil {
		return fmt.Errorf("could not create metadata: %w", err)
	} else if metadata == nil {
		return fmt.Errorf("metadata is nil")
	}

	environ := metadata.Environ()
	maps.Copy(environ, metadata.Workflow.Matrix)
	var secrets []compiler.Secret
	for key, val := range c.StringMap("secrets") {
		secrets = append(secrets, compiler.Secret{
			Name:  key,
			Value: val,
		})
	}

	pipelineEnv := make(map[string]string)
	for _, env := range c.StringSlice("env") {
		before, after, _ := strings.Cut(env, "=")
		pipelineEnv[before] = after
		if oldVar, exists := environ[before]; exists {
			// override existing values, but print a warning
			log.Warn().Msgf("environment variable '%s' had value '%s', but got overwritten", before, oldVar)
		}
		environ[before] = after
	}

	tmpl, err := envsubst.ParseFile(file)
	if err != nil {
		return err
	}
	confStr, err := tmpl.Execute(func(name string) string {
		return environ[name]
	})
	if err != nil {
		return err
	}

	conf, err := yaml.ParseString(confStr)
	if err != nil {
		return err
	}

	// emulate server behavior https://github.com/woodpecker-ci/woodpecker/blob/eebaa10d104cbc3fa7ce4c0e344b0b7978405135/server/pipeline/stepbuilder/stepBuilder.go#L289-L295
	prefix := "wp_" + ulid.Make().String()

	// configure volumes for local execution
	volumes := c.StringSlice("volumes")
	if c.Bool("local") {
		var (
			workspaceBase = conf.Workspace.Base
			workspacePath = conf.Workspace.Path
		)
		if workspaceBase == "" {
			workspaceBase = c.String("workspace-base")
		}
		if workspacePath == "" {
			workspacePath = c.String("workspace-path")
		}

		volumes = append(volumes, prefix+"_default:"+workspaceBase)
		volumes = append(volumes, repoPath+":"+path.Join(workspaceBase, workspacePath))
	}

	privilegedPlugins := c.StringSlice("plugins-privileged")

	// lint the yaml file
	err = linter.New(
		linter.WithTrusted(linter.TrustedConfiguration{
			Security: c.Bool("repo-trusted-security"),
			Network:  c.Bool("repo-trusted-network"),
			Volumes:  c.Bool("repo-trusted-volumes"),
		}),
		linter.PrivilegedPlugins(privilegedPlugins),
		linter.WithTrustedClonePlugins(constant.TrustedClonePlugins),
	).Lint([]*linter.WorkflowConfig{{
		File:      path.Base(file),
		RawConfig: confStr,
		Workflow:  conf,
	}})
	if err != nil {
		str, err := lint.FormatLintError(file, err, false)
		fmt.Print(str)
		if err != nil {
			return err
		}
	}

	// compiles the yaml file
	compiled, err := compiler.New(
		compiler.WithEscalated(
			privilegedPlugins...,
		),
		compiler.WithVolumes(volumes...),
		compiler.WithWorkspace(
			c.String("workspace-base"),
			c.String("workspace-path"),
		),
		compiler.WithNetworks(
			c.StringSlice("network")...,
		),
		compiler.WithPrefix(prefix),
		compiler.WithProxy(compiler.ProxyOptions{
			NoProxy:    c.String("backend-no-proxy"),
			HTTPProxy:  c.String("backend-http-proxy"),
			HTTPSProxy: c.String("backend-https-proxy"),
		}),
		compiler.WithLocal(
			c.Bool("local"),
		),
		compiler.WithNetrc(
			c.String("netrc-username"),
			c.String("netrc-password"),
			c.String("netrc-machine"),
		),
		compiler.WithMetadata(*metadata),
		compiler.WithSecret(secrets...),
		compiler.WithEnviron(pipelineEnv),
	).Compile(conf)
	if err != nil {
		return err
	}

	backendCtx := context.WithValue(ctx, backend_types.CliCommand, c)
	backendEngine, err := backend.FindBackend(backendCtx, backends, c.String("backend-engine"))
	if err != nil {
		return err
	}

	if _, err = backendEngine.Load(backendCtx); err != nil {
		return err
	}

	pipelineCtx, cancel := context.WithTimeout(context.Background(), c.Duration("timeout"))
	defer cancel()
	pipelineCtx = utils.WithContextSigtermCallback(pipelineCtx, func() {
		fmt.Printf("ctrl+c received, terminating current pipeline '%s'\n", confStr)
	})

	return pipeline.New(compiled,
		pipeline.WithContext(pipelineCtx), //nolint:contextcheck
		pipeline.WithTracer(pipeline.DefaultTracer),
		pipeline.WithLogger(defaultLogger),
		pipeline.WithBackend(backendEngine),
		pipeline.WithDescription(map[string]string{
			"CLI": "exec",
		}),
	).Run(ctx)
}

// convertPathForWindows converts a path to use slash separators
// for Windows. If the path is a Windows volume name like C:, it
// converts it to an absolute root path starting with slash (e.g.
// C: -> /c). Otherwise it just converts backslash separators to
// slashes.
func convertPathForWindows(path string) string {
	base := filepath.VolumeName(path)

	// Check if path is volume name like C:
	//nolint:mnd
	if len(base) == 2 {
		path = strings.TrimPrefix(path, base)
		base = strings.ToLower(strings.TrimSuffix(base, ":"))
		return "/" + base + filepath.ToSlash(path)
	}

	return filepath.ToSlash(path)
}

var defaultLogger = pipeline.Logger(func(step *backend_types.Step, rc io.ReadCloser) error {
	logWriter := NewLineWriter(step.Name, step.UUID)
	return pipelineLog.CopyLineByLine(logWriter, rc, pipeline.MaxLogLineLength)
})
