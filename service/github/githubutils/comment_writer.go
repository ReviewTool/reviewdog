package githubutils

import (
	"context"
	"fmt"
	"sync"

	"github.com/haya14busa/go-actions-toolkit/core"

	"github.com/reviewtool/reviewdog"
	"github.com/reviewtool/reviewdog/proto/rdf"
)

const MaxLoggingAnnotationsPerStep = 10

var _ reviewdog.CommentService = &GitHubActionLogWriter{}

// GitHubActionLogWriter reports results via logging command to create
// annotations.
// https://help.github.com/en/actions/automating-your-workflow-with-github-actions/development-tools-for-github-actions#example-5
type GitHubActionLogWriter struct {
	level     string
	reportNum int
}

// NewGitHubActionLogWriter returns new GitHubActionLogWriter.
func NewGitHubActionLogWriter(level string) *GitHubActionLogWriter {
	return &GitHubActionLogWriter{level: level}
}

func (lw *GitHubActionLogWriter) Post(_ context.Context, c *reviewdog.Comment) error {
	lw.reportNum++
	if lw.reportNum == MaxLoggingAnnotationsPerStep {
		WarnTooManyAnnotationOnce()
	}
	ReportAsGitHubActionsLog(c.ToolName, lw.level, c.Result.Diagnostic)
	return nil
}

// Flush checks overall error at last.
func (lw *GitHubActionLogWriter) Flush(_ context.Context) error {
	if lw.reportNum > 9 {
		return fmt.Errorf("GitHubActionLogWriter: reported too many annotation (N=%d)", lw.reportNum)
	}
	return nil
}

// ReportAsGitHubActionsLog reports results via logging command to create
// annotations.
// https://help.github.com/en/actions/automating-your-workflow-with-github-actions/development-tools-for-github-actions#example-5
func ReportAsGitHubActionsLog(toolName, defaultLevel string, d *rdf.Diagnostic) {
	mes := fmt.Sprintf("[%s] reported by reviewdog 🐶\n%s\n\nRaw Output:\n%s",
		toolName, d.GetMessage(), d.GetOriginalOutput())
	loc := d.GetLocation()
	start := loc.GetRange().GetStart()
	opt := &core.LogOption{
		File: d.GetLocation().GetPath(),
		Line: int(start.GetLine()),
		Col:  int(start.GetColumn()),
	}

	level := defaultLevel
	switch d.Severity {
	case rdf.Severity_ERROR:
		level = "error"
	case rdf.Severity_INFO, rdf.Severity_WARNING:
		level = "warning"
	}

	switch level {
	// no info command with location data.
	case "warning", "info":
		core.Warning(mes, opt)
	case "error", "":
		core.Error(mes, opt)
	default:
		core.Error(fmt.Sprintf("Unknown level: %s", level), nil)
		core.Error(mes, opt)
	}
}

func WarnTooManyAnnotationOnce() {
	warnTooManyAnnotationOnce.Do(warnTooManyAnnotation)
}

var warnTooManyAnnotationOnce sync.Once

func warnTooManyAnnotation() {
	core.Error(`reviewdog: Too many results (annotations) in diff.
You may miss some annotations due to GitHub limitation for annotation created by logging command.
Please check GitHub Actions log console to see all results.

Limitation:
- 10 warning annotations and 10 error annotations per step
- 50 annotations per job (sum of annotations from all the steps)
- 50 annotations per run (separate from the job annotations, these annotations aren't created by users)

Source: https://github.community/t5/GitHub-Actions/Maximum-number-of-annotations-that-can-be-created-using-GitHub/m-p/39085`, nil)
}
