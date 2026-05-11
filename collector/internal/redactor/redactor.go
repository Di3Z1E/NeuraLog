package redactor

import (
	"regexp"
	"sync"

	"github.com/Di3Z1E/neuralog/internal/config"
)

type compiledPattern struct {
	re      *regexp.Regexp
	replace string
}

type Redactor struct {
	mu       sync.RWMutex
	cfgMgr   *config.Manager
	patterns []compiledPattern
	enabled  bool
}

func New(cfgMgr *config.Manager) *Redactor {
	r := &Redactor{cfgMgr: cfgMgr}
	r.rebuild(cfgMgr.Get())
	return r
}

// Reload rebuilds the pattern list from current config — called after a config save.
func (r *Redactor) Reload() {
	r.rebuild(r.cfgMgr.Get())
}

func (r *Redactor) rebuild(cfg config.Config) {
	patterns := make([]compiledPattern, 0, len(builtinPatterns)+len(cfg.CustomPatterns))
	for _, bp := range builtinPatterns {
		re, err := regexp.Compile(bp.pattern)
		if err != nil {
			continue
		}
		patterns = append(patterns, compiledPattern{re, bp.replace})
	}
	for _, cp := range cfg.CustomPatterns {
		re, err := regexp.Compile(cp.Pattern)
		if err != nil {
			continue
		}
		patterns = append(patterns, compiledPattern{re, cp.Replace})
	}
	r.mu.Lock()
	r.enabled = cfg.RedactEnabled
	r.patterns = patterns
	r.mu.Unlock()
}

func (r *Redactor) Apply(line string) string {
	r.mu.RLock()
	enabled := r.enabled
	patterns := r.patterns
	r.mu.RUnlock()

	if !enabled {
		return line
	}
	for _, p := range patterns {
		line = p.re.ReplaceAllString(line, p.replace)
	}
	return line
}
