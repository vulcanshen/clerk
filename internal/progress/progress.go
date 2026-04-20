package progress

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var spinnerChars = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// Steps provides a step-based progress indicator on stderr.
// If stderr is not a terminal, all methods are no-ops.
type Steps struct {
	enabled bool
	label   string
	stop    chan struct{}
	wg      sync.WaitGroup
}

func isTerminal() bool {
	info, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// New creates a new Steps. Progress is only shown when stderr is a terminal.
func New() *Steps {
	return &Steps{
		enabled: isTerminal(),
	}
}

// Start begins a new step with a spinning indicator.
func (s *Steps) Start(label string) {
	if !s.enabled {
		return
	}
	s.label = label
	s.stop = make(chan struct{})
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-s.stop:
				return
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "\r%c %s", spinnerChars[i%len(spinnerChars)], s.label)
				i++
			}
		}
	}()
}

// Done marks the current step as complete with a checkmark.
func (s *Steps) Done() {
	if !s.enabled {
		return
	}
	if s.stop != nil {
		close(s.stop)
		s.wg.Wait()
	}
	fmt.Fprintf(os.Stderr, "\r✓ %s\n", s.label)
}

// DoneMsg marks the current step as complete with a custom message.
func (s *Steps) DoneMsg(msg string) {
	if !s.enabled {
		return
	}
	if s.stop != nil {
		close(s.stop)
		s.wg.Wait()
	}
	fmt.Fprintf(os.Stderr, "\r✓ %s\n", msg)
}

// Fail marks the current step as failed with an error.
func (s *Steps) Fail(err error) {
	if !s.enabled {
		return
	}
	if s.stop != nil {
		close(s.stop)
		s.wg.Wait()
	}
	fmt.Fprintf(os.Stderr, "\r✗ %s — %v\n", s.label, err)
}
