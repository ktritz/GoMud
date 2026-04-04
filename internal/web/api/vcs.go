package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

var (
	vcsMu      sync.Mutex
	vcsDataDir string
)

// InitVCS starts the version control system for world data.
// Must be called after configs are loaded.
func InitVCS() {
	vcsDataDir = configs.GetFilePathsConfig().DataFiles.String()

	// Verify git repo exists
	if !isGitRepo() {
		mudlog.Info("VCS", "status", "No git repo in data dir, version control disabled")
		return
	}

	mudlog.Info("VCS", "status", "Version control active", "dataDir", vcsDataDir)

	// Start auto-commit timer
	go autoCommitLoop()
}

// gitCmd creates a git command with the correct environment for the data dir.
func gitCmd(args ...string) *exec.Cmd {
	fullArgs := append([]string{"-C", vcsDataDir}, args...)
	cmd := exec.Command("git", fullArgs...)
	// Ensure HOME is set so git can find .gitconfig (safe.directory)
	cmd.Env = append(cmd.Environ(), "HOME=/opt/gomud")
	return cmd
}

func isGitRepo() bool {
	return gitCmd("rev-parse", "--git-dir").Run() == nil
}

// isDirty returns true if there are uncommitted changes in the data dir.
func isDirty() bool {
	cmd := gitCmd( "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// changedFiles returns the list of modified/added/deleted files.
func changedFiles() []string {
	cmd := gitCmd( "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(string(out), "\n")
	files := []string{}
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		// Porcelain format: XY<space>path (first 3 chars are status + space)
		files = append(files, line[3:])
	}
	return files
}

// buildAutoMessage creates a commit message summarizing what changed.
func buildAutoMessage() string {
	files := changedFiles()
	if len(files) == 0 {
		return "Auto-save"
	}
	if len(files) == 1 {
		return fmt.Sprintf("Auto-save: %s", files[0])
	}
	if len(files) <= 5 {
		return fmt.Sprintf("Auto-save: %s", strings.Join(files, ", "))
	}
	return fmt.Sprintf("Auto-save: %d files changed", len(files))
}

// commitAll stages all changes and commits with the given message and author.
func commitAll(message string, author string) error {
	vcsMu.Lock()
	defer vcsMu.Unlock()

	if !isDirty() {
		return nil
	}

	// Stage all changes
	addCmd := gitCmd( "add", "-A")
	if out, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s: %w", string(out), err)
	}

	// Build commit args
	commitArgs := []string{"commit", "-m", message}
	if author != "" {
		commitArgs = append(commitArgs, "--author", fmt.Sprintf("%s <editor@gomud.local>", author))
	}

	commitCmd := gitCmd(commitArgs...)
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s: %w", string(out), err)
	}

	mudlog.Info("VCS", "action", "commit", "message", message, "author", author)
	return nil
}

// autoCommitLoop runs every 5 minutes and commits if dirty.
func autoCommitLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if !isDirty() {
			continue
		}
		msg := buildAutoMessage()
		if err := commitAll(msg, "auto"); err != nil {
			mudlog.Error("VCS", "error", "Auto-commit failed", "detail", err)
		}
	}
}

// --- API Handlers ---

type commitEntry struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"shortHash"`
	Author    string `json:"author"`
	Date      string `json:"date"`
	Message   string `json:"message"`
}

func handleListCommits(w http.ResponseWriter, r *http.Request) {
	if !isGitRepo() {
		writeError(w, http.StatusServiceUnavailable, "Version control not active")
		return
	}

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50"
	}

	cmd := gitCmd( "log",
		"--format=%H|%h|%an|%aI|%s",
		"-n", limit)
	out, err := cmd.Output()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to read history")
		return
	}

	entries := []commitEntry{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}
		entries = append(entries, commitEntry{
			Hash:      parts[0],
			ShortHash: parts[1],
			Author:    parts[2],
			Date:      parts[3],
			Message:   parts[4],
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"commits": entries,
		"total":   len(entries),
	})
}

func handleGetDiff(w http.ResponseWriter, r *http.Request) {
	if !isGitRepo() {
		writeError(w, http.StatusServiceUnavailable, "Version control not active")
		return
	}

	hash := r.PathValue("hash")
	if hash == "" {
		writeError(w, http.StatusBadRequest, "Commit hash required")
		return
	}

	cmd := gitCmd( "show", "--stat", "--patch", hash)
	out, err := cmd.Output()
	if err != nil {
		writeError(w, http.StatusNotFound, "Commit not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"hash": hash,
		"diff": string(out),
	})
}

type checkpointRequest struct {
	Message string `json:"message"`
}

func handleCheckpoint(w http.ResponseWriter, r *http.Request) {
	if !isGitRepo() {
		writeError(w, http.StatusServiceUnavailable, "Version control not active")
		return
	}

	var req checkpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Message == "" {
		req.Message = "Manual checkpoint"
	}

	// Get username from JWT claims
	author := "unknown"
	if claims, ok := getClaimsFromContext(r); ok {
		author = claims.Username
	}

	if !isDirty() {
		writeJSON(w, http.StatusOK, map[string]any{
			"committed": false,
			"message":   "No changes to commit",
		})
		return
	}

	if err := commitAll(req.Message, author); err != nil {
		writeError(w, http.StatusInternalServerError, "Commit failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"committed": true,
		"message":   req.Message,
	})
}

func handleRevert(w http.ResponseWriter, r *http.Request) {
	if !isGitRepo() {
		writeError(w, http.StatusServiceUnavailable, "Version control not active")
		return
	}

	hash := r.PathValue("hash")
	if hash == "" {
		writeError(w, http.StatusBadRequest, "Commit hash required")
		return
	}

	author := "unknown"
	if claims, ok := getClaimsFromContext(r); ok {
		author = claims.Username
	}

	// Use git revert to create a new commit that undoes the target commit
	vcsMu.Lock()
	defer vcsMu.Unlock()

	cmd := gitCmd( "revert", "--no-commit", hash)
	if out, err := cmd.CombinedOutput(); err != nil {
		writeError(w, http.StatusInternalServerError, "Revert failed: "+strings.TrimSpace(string(out)))
		return
	}

	// Commit the revert
	msg := fmt.Sprintf("Revert %s (by %s)", hash[:8], author)
	commitCmd := gitCmd( "commit", "-m", msg,
		"--author", fmt.Sprintf("%s <editor@gomud.local>", author))
	if out, err := commitCmd.CombinedOutput(); err != nil {
		// Abort if commit fails
		gitCmd( "reset", "--hard", "HEAD").Run()
		writeError(w, http.StatusInternalServerError, "Revert commit failed: "+strings.TrimSpace(string(out)))
		return
	}

	mudlog.Info("VCS", "action", "revert", "hash", hash, "author", author)

	writeJSON(w, http.StatusOK, map[string]any{
		"reverted": true,
		"hash":     hash,
	})
}

func handleVCSStatus(w http.ResponseWriter, r *http.Request) {
	if !isGitRepo() {
		writeJSON(w, http.StatusOK, map[string]any{
			"active": false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"active":  true,
		"dirty":   isDirty(),
		"changed": changedFiles(),
	})
}
