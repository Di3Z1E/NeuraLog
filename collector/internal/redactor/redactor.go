package redactor

import "regexp"

type compiledPattern struct {
	re      *regexp.Regexp
	replace string
}

type Redactor struct {
	patterns []compiledPattern
	enabled  bool
}

func New(enabled bool) *Redactor {
	r := &Redactor{enabled: enabled}
	for _, bp := range builtinPatterns {
		re, err := regexp.Compile(bp.pattern)
		if err != nil {
			continue
		}
		r.patterns = append(r.patterns, compiledPattern{re, bp.replace})
	}
	return r
}

func (r *Redactor) AddPattern(expr, replace string) error {
	re, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	r.patterns = append(r.patterns, compiledPattern{re, replace})
	return nil
}

func (r *Redactor) Apply(line string) string {
	if !r.enabled {
		return line
	}
	for _, p := range r.patterns {
		line = p.re.ReplaceAllString(line, p.replace)
	}
	return line
}
