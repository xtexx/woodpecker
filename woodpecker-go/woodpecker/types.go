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

package woodpecker

type ApprovalMode string

var (
	RequireApprovalNone         ApprovalMode = "none"          // require approval for no events
	RequireApprovalForks        ApprovalMode = "forks"         // require approval for PRs from forks
	RequireApprovalPullRequests ApprovalMode = "pull_requests" // require approval for all PRs (default)
	RequireApprovalAllEvents    ApprovalMode = "all_events"    // require approval for all events
)

func (mode ApprovalMode) Valid() bool {
	switch mode {
	case RequireApprovalNone,
		RequireApprovalForks,
		RequireApprovalPullRequests,
		RequireApprovalAllEvents:
		return true
	default:
		return false
	}
}

type (
	// User represents a user account.
	User struct {
		ID     int64  `json:"id"`
		Login  string `json:"login"`
		Email  string `json:"email"`
		Avatar string `json:"avatar_url"`
		Active bool   `json:"active"`
		Admin  bool   `json:"admin"`
	}

	TrustedConfiguration struct {
		Network  bool `json:"network"`
		Volumes  bool `json:"volumes"`
		Security bool `json:"security"`
	}

	// Repo represents a repository.
	Repo struct {
		ID                           int64                `json:"id,omitempty"`
		ForgeRemoteID                string               `json:"forge_remote_id"`
		Owner                        string               `json:"owner"`
		Name                         string               `json:"name"`
		FullName                     string               `json:"full_name"`
		Avatar                       string               `json:"avatar_url,omitempty"`
		ForgeURL                     string               `json:"forge_url,omitempty"`
		Clone                        string               `json:"clone_url,omitempty"`
		Branch                       string               `json:"default_branch,omitempty"`
		SCMKind                      string               `json:"scm,omitempty"`
		Timeout                      int64                `json:"timeout,omitempty"`
		Visibility                   string               `json:"visibility"`
		IsSCMPrivate                 bool                 `json:"private"`
		Trusted                      TrustedConfiguration `json:"trusted"`
		RequireApproval              ApprovalMode         `json:"require_approval"`
		IsActive                     bool                 `json:"active"`
		AllowPull                    bool                 `json:"allow_pr"`
		Config                       string               `json:"config_file"`
		CancelPreviousPipelineEvents []string             `json:"cancel_previous_pipeline_events"`
		NetrcTrustedPlugins          []string             `json:"netrc_trusted"`
	}

	// RepoPatch defines a repository patch request.
	RepoPatch struct {
		Config          *string       `json:"config_file,omitempty"`
		IsTrusted       *bool         `json:"trusted,omitempty"`
		RequireApproval *ApprovalMode `json:"require_approval,omitempty"`
		Timeout         *int64        `json:"timeout,omitempty"`
		Visibility      *string       `json:"visibility"`
		AllowPull       *bool         `json:"allow_pr,omitempty"`
		PipelineCounter *int          `json:"pipeline_counter,omitempty"`
	}

	PipelineError struct {
		Type      string `json:"type"`
		Message   string `json:"message"`
		IsWarning bool   `json:"is_warning"`
		Data      any    `json:"data"`
	}

	// Pipeline defines a pipeline object.
	Pipeline struct {
		ID        int64            `json:"id"`
		Number    int64            `json:"number"`
		Parent    int64            `json:"parent"`
		Event     string           `json:"event"`
		Status    string           `json:"status"`
		Errors    []*PipelineError `json:"errors"`
		Created   int64            `json:"created"`
		Updated   int64            `json:"updated"`
		Started   int64            `json:"started"`
		Finished  int64            `json:"finished"`
		Deploy    string           `json:"deploy_to"`
		Commit    string           `json:"commit"`
		Branch    string           `json:"branch"`
		Ref       string           `json:"ref"`
		Refspec   string           `json:"refspec"`
		Title     string           `json:"title"`
		Message   string           `json:"message"`
		Timestamp int64            `json:"timestamp"`
		Sender    string           `json:"sender"`
		Author    string           `json:"author"`
		Avatar    string           `json:"author_avatar"`
		Email     string           `json:"author_email"`
		ForgeURL  string           `json:"forge_url"`
		Reviewer  string           `json:"reviewed_by"`
		Reviewed  int64            `json:"reviewed"`
		Workflows []*Workflow      `json:"workflows,omitempty"`
	}

	// Workflow represents a workflow in the pipeline.
	Workflow struct {
		ID       int64             `json:"id"`
		PID      int               `json:"pid"`
		Name     string            `json:"name"`
		State    string            `json:"state"`
		Error    string            `json:"error,omitempty"`
		Started  int64             `json:"started,omitempty"`
		Stopped  int64             `json:"finished,omitempty"`
		AgentID  int64             `json:"agent_id,omitempty"`
		Platform string            `json:"platform,omitempty"`
		Environ  map[string]string `json:"environ,omitempty"`
		Children []*Step           `json:"children,omitempty"`
	}

	// Step represents a process in the pipeline.
	Step struct {
		ID       int64    `json:"id"`
		PID      int      `json:"pid"`
		PPID     int      `json:"ppid"`
		Name     string   `json:"name"`
		State    string   `json:"state"`
		Error    string   `json:"error,omitempty"`
		ExitCode int      `json:"exit_code"`
		Started  int64    `json:"started,omitempty"`
		Stopped  int64    `json:"finished,omitempty"`
		Type     StepType `json:"type,omitempty"`
	}

	// Registry represents a docker registry with credentials.
	Registry struct {
		ID       int64  `json:"id"`
		OrgID    int64  `json:"org_id"`
		RepoID   int64  `json:"repo_id"`
		Address  string `json:"address"`
		Username string `json:"username"`
		Password string `json:"password,omitempty"`
	}

	// Secret represents a secret variable, such as a password or token.
	Secret struct {
		ID     int64    `json:"id"`
		OrgID  int64    `json:"org_id"`
		RepoID int64    `json:"repo_id"`
		Name   string   `json:"name"`
		Value  string   `json:"value,omitempty"`
		Images []string `json:"images"`
		Events []string `json:"events"`
	}

	// Feed represents an item in the user's feed or timeline.
	Feed struct {
		RepoID   int64  `json:"repo_id"`
		ID       int64  `json:"id,omitempty"`
		Number   int64  `json:"number,omitempty"`
		Event    string `json:"event,omitempty"`
		Status   string `json:"status,omitempty"`
		Created  int64  `json:"created,omitempty"`
		Started  int64  `json:"started,omitempty"`
		Finished int64  `json:"finished,omitempty"`
		Commit   string `json:"commit,omitempty"`
		Branch   string `json:"branch,omitempty"`
		Ref      string `json:"ref,omitempty"`
		Refspec  string `json:"refspec,omitempty"`
		Remote   string `json:"remote,omitempty"`
		Title    string `json:"title,omitempty"`
		Message  string `json:"message,omitempty"`
		Author   string `json:"author,omitempty"`
		Avatar   string `json:"author_avatar,omitempty"`
		Email    string `json:"author_email,omitempty"`
	}

	// Version provides system version details.
	Version struct {
		Source  string `json:"source,omitempty"`
		Version string `json:"version,omitempty"`
		Commit  string `json:"commit,omitempty"`
	}

	QueueStats struct {
		Workers       int `json:"worker_count"`
		Pending       int `json:"pending_count"`
		WaitingOnDeps int `json:"waiting_on_deps_count"`
		Running       int `json:"running_count"`
		Complete      int `json:"completed_count"`
	}

	// Info provides queue stats.
	Info struct {
		Pending       []Task     `json:"pending"`
		WaitingOnDeps []Task     `json:"waiting_on_deps"`
		Running       []Task     `json:"running"`
		Stats         QueueStats `json:"stats"`
		Paused        bool       `json:"paused,omitempty"`
	}

	// LogLevel is for checking/setting logging level.
	LogLevel struct {
		Level string `json:"log-level"`
	}

	// LogEntry is a single log entry.
	LogEntry struct {
		ID     int64        `json:"id"`
		StepID int64        `json:"step_id"`
		Time   int64        `json:"time"`
		Line   int          `json:"line"`
		Data   []byte       `json:"data"`
		Type   LogEntryType `json:"type"`
	}

	// Cron is the JSON data of a cron job.
	Cron struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		RepoID    int64  `json:"repo_id"`
		CreatorID int64  `json:"creator_id"`
		NextExec  int64  `json:"next_exec"`
		Schedule  string `json:"schedule"`
		Created   int64  `json:"created"`
		Branch    string `json:"branch"`
	}

	// PipelineOptions is the JSON data for creating a new pipeline.
	PipelineOptions struct {
		Branch    string            `json:"branch"`
		Variables map[string]string `json:"variables"`
	}

	// Agent is the JSON data for an agent.
	Agent struct {
		ID           int64             `json:"id"`
		Created      int64             `json:"created"`
		Updated      int64             `json:"updated"`
		Name         string            `json:"name"`
		OwnerID      int64             `json:"owner_id"`
		OrgID        int64             `json:"org_id"`
		Token        string            `json:"token"`
		LastContact  int64             `json:"last_contact"`
		LastWork     int64             `json:"last_work"`
		Platform     string            `json:"platform"`
		Backend      string            `json:"backend"`
		Capacity     int32             `json:"capacity"`
		Version      string            `json:"version"`
		NoSchedule   bool              `json:"no_schedule"`
		CustomLabels map[string]string `json:"custom_labels"`
	}

	// Task is the JSON data for a task.
	Task struct {
		ID           string            `json:"id"`
		Labels       map[string]string `json:"labels"`
		Dependencies []string          `json:"dependencies"`
		RunOn        []string          `json:"run_on"`
		DepStatus    map[string]string `json:"dep_status"`
		AgentID      int64             `json:"agent_id"`
	}

	// Org is the JSON data for an organization.
	Org struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		IsUser bool   `json:"is_user"`
	}
)
