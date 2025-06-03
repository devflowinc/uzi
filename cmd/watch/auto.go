package watch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/devflowinc/uzi/pkg/state"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type AgentWatcher struct {
	stateManager    *state.StateManager
	watchedSessions map[string]*SessionMonitor
	mu              sync.RWMutex
	quit            chan bool
}

type SessionMonitor struct {
	sessionName    string
	prevOutputHash []byte
	lastUpdated    time.Time
	updateCount    int
	noUpdateCount  int
}

func NewAgentWatcher() *AgentWatcher {
	return &AgentWatcher{
		stateManager:    state.NewStateManager(),
		watchedSessions: make(map[string]*SessionMonitor),
		quit:            make(chan bool),
	}
}

func (aw *AgentWatcher) hashContent(content []byte) []byte {
	hash := sha256.Sum256(content)
	return hash[:]
}

func (aw *AgentWatcher) capturePaneContent(sessionName string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName+":agent", "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (aw *AgentWatcher) sendKeys(sessionName string, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName+":agent", keys)
	return cmd.Run()
}

func (aw *AgentWatcher) tapEnter(sessionName string) error {
	return aw.sendKeys(sessionName, "Enter")
}

func (aw *AgentWatcher) hasUpdated(sessionName string) (bool, bool, error) {
	content, err := aw.capturePaneContent(sessionName)
	if err != nil {
		return false, false, err
	}

	// Check for specific prompts that need auto-enter
	hasPrompt := false

	// Check for Claude trust prompt
	if strings.Contains(content, "Do you trust the files in this folder?") {
		hasPrompt = true
	}

	// Check for general continuation prompts
	if strings.Contains(content, "Press Enter to continue") ||
		strings.Contains(content, "Continue? (Y/n)") ||
		strings.Contains(content, "Do you want to proceed?") ||
		(strings.Contains(content, "Allow command") && !strings.Contains(content, "Thinking")) ||
		strings.Contains(content, "Do you want to") ||
		strings.Contains(content, "Proceed? (y/N)") {
		hasPrompt = true
	}

	aw.mu.Lock()
	monitor, exists := aw.watchedSessions[sessionName]
	if !exists {
		// First time monitoring this session
		aw.watchedSessions[sessionName] = &SessionMonitor{
			sessionName:    sessionName,
			prevOutputHash: aw.hashContent([]byte(content)),
			lastUpdated:    time.Now(),
			updateCount:    0,
			noUpdateCount:  0,
		}
		aw.mu.Unlock()
		return false, hasPrompt, nil
	}

	// Compare current content hash with previous
	currentHash := aw.hashContent([]byte(content))
	if !bytes.Equal(currentHash, monitor.prevOutputHash) {
		monitor.prevOutputHash = currentHash
		monitor.lastUpdated = time.Now()
		monitor.updateCount++
		monitor.noUpdateCount = 0
		aw.mu.Unlock()
		return true, hasPrompt, nil
	}

	monitor.noUpdateCount++
	aw.mu.Unlock()
	return false, hasPrompt, nil
}

func (aw *AgentWatcher) watchSession(sessionName string) {
	log.Info("Starting to watch session", "session", sessionName)

	for {
		select {
		case <-aw.quit:
			return
		default:
			updated, hasPrompt, err := aw.hasUpdated(sessionName)
			if err != nil {
				log.Error("Error checking session update", "session", sessionName, "error", err)
				time.Sleep(2 * time.Second)
				continue
			}

			if updated {
				log.Debug("Session updated", "session", sessionName)
			} else {
				// Check if session has no prompt and hasn't updated in 3 cycles
			}

			if hasPrompt {
				log.Info("Auto-pressing Enter for prompt", "session", sessionName)
				if err := aw.tapEnter(sessionName); err != nil {
					log.Error("Failed to send Enter", "session", sessionName, "error", err)
				} else {
					log.Info("Successfully sent Enter", "session", sessionName)
				}
			}

			time.Sleep(500 * time.Millisecond) // Check every 500ms
		}
	}
}

func (aw *AgentWatcher) refreshActiveSessions() error {
	activeSessions, err := aw.stateManager.GetActiveSessionsForRepo()
	if err != nil {
		return fmt.Errorf("failed to get active sessions: %w", err)
	}

	// Stop watching sessions that are no longer active
	aw.mu.Lock()
	for sessionName := range aw.watchedSessions {
		found := false
		for _, activeSession := range activeSessions {
			if activeSession == sessionName {
				found = true
				break
			}
		}
		if !found {
			log.Info("Session no longer active, stopping watch", "session", sessionName)
			delete(aw.watchedSessions, sessionName)
		}
	}
	aw.mu.Unlock()

	// Start watching new active sessions
	for _, sessionName := range activeSessions {
		aw.mu.RLock()
		_, exists := aw.watchedSessions[sessionName]
		aw.mu.RUnlock()
		if !exists {
			go aw.watchSession(sessionName)
		}
	}

	return nil
}

func (aw *AgentWatcher) Start() {
	log.Info("Starting Agent Watcher")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Refresh active sessions periodically
	refreshTicker := time.NewTicker(5 * time.Second)
	defer refreshTicker.Stop()

	go func() {
		for {
			select {
			case <-refreshTicker.C:
				if err := aw.refreshActiveSessions(); err != nil {
					log.Error("Failed to refresh active sessions", "error", err)
				}
			case <-aw.quit:
				return
			}
		}
	}()

	// Initial refresh
	if err := aw.refreshActiveSessions(); err != nil {
		log.Error("Failed initial session refresh", "error", err)
	}

	// Wait for signal
	<-sigChan
	log.Info("Shutting down Agent Watcher")
	close(aw.quit)
}

var CmdWatch = &ffcli.Command{
	Name:       "auto",
	ShortUsage: "uzi auto",
	ShortHelp:  "Automatically manage active agent sessions",
	LongHelp: `
The auto command monitors all active agent sessions in the current repository
and automatically presses Enter when it detects prompts that require user input,
such as trust prompts or continuation confirmations. It can also handle other
automated tasks in the future.

This is useful for hands-free operation of multiple agents.
`,
	FlagSet: func() *flag.FlagSet {
		fs := flag.NewFlagSet("auto", flag.ExitOnError)
		return fs
	}(),
	Exec: func(ctx context.Context, args []string) error {
		watcher := NewAgentWatcher()
		watcher.Start()
		return nil
	},
}
