package uxi

import "time"

const (
	StatusProposed         = "proposed"
	StatusAwaitingApproval = "awaiting_approval"
	StatusApproved         = "approved"
	StatusQueued           = "queued"
	StatusRunning          = "running"
	StatusBlocked          = "blocked"
	StatusSucceeded        = "succeeded"
	StatusFailed           = "failed"
	StatusCancelled        = "cancelled"
	StatusRolledBack       = "rolled_back"
)

// Session is the persisted operator workspace and its governed history.
type Session struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Context   SessionContext `json:"context"`
	Turns     []Turn         `json:"turns"`
	Plans     []Plan         `json:"plans"`
	Actions   []Action       `json:"actions"`
	Approvals []Approval     `json:"approvals"`
	Jobs      []Job          `json:"jobs"`
	Artifacts []Artifact     `json:"artifacts"`
}

// SessionContext identifies the active repositories, application, and mode.
type SessionContext struct {
	RepositoryIDs []string `json:"repository_ids,omitempty"`
	ApplicationID string   `json:"application_id,omitempty"`
	Mode          string   `json:"mode,omitempty"`
}

// Turn is one persisted user or assistant message.
type Turn struct {
	ID        string         `json:"id"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	Evidence  []Evidence     `json:"evidence,omitempty"`
	Plan      *Plan          `json:"plan,omitempty"`
	Actions   []Action       `json:"actions,omitempty"`
	Artifacts []Artifact     `json:"artifacts,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Evidence records an observation supporting a status or answer.
type Evidence struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	Source     string    `json:"source"`
	Summary    string    `json:"summary"`
	Detail     string    `json:"detail,omitempty"`
	Status     string    `json:"status"`
	ObservedAt time.Time `json:"observed_at"`
}

// Plan is a proposed sequence of work and never implies execution.
type Plan struct {
	ID        string     `json:"id"`
	Objective string     `json:"objective"`
	Status    string     `json:"status"`
	Steps     []PlanStep `json:"steps"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// PlanStep is one proposed step in a plan.
type PlanStep struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Evidence string `json:"evidence,omitempty"`
}

// Action is a typed proposed operation subject to policy and approval.
type Action struct {
	ID               string         `json:"id"`
	Kind             string         `json:"kind"`
	Target           string         `json:"target"`
	Status           string         `json:"status"`
	Risk             string         `json:"risk"`
	ApprovalRequired bool           `json:"approval_required"`
	Parameters       map[string]any `json:"parameters,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// Approval records a human decision about an action.
type Approval struct {
	ID        string     `json:"id"`
	ActionID  string     `json:"action_id"`
	Decision  string     `json:"decision"`
	Actor     string     `json:"actor"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// Job records an execution lifecycle when an executor is introduced.
type Job struct {
	ID        string     `json:"id"`
	ActionID  string     `json:"action_id"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	LogPath   string     `json:"log_path,omitempty"`
}

// Artifact identifies a produced file or other durable output.
type Artifact struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Digest    string    `json:"digest,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Event is an append-only audit observation.
type Event struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id,omitempty"`
	Kind      string         `json:"kind"`
	Status    string         `json:"status"`
	Message   string         `json:"message"`
	Data      map[string]any `json:"data,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// Capability is one evidence-backed operation declared by a repository.
type Capability struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	Command     string `json:"command,omitempty"`
	Evidence    string `json:"evidence,omitempty"`
}

// Repository is the UXI view of one discovered Git repository.
type Repository struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Path         string       `json:"path"`
	Role         string       `json:"role"`
	Status       string       `json:"status"`
	Git          bool         `json:"git"`
	Languages    []string     `json:"languages,omitempty"`
	Capabilities []Capability `json:"capabilities,omitempty"`
	Evidence     []Evidence   `json:"evidence,omitempty"`
}
