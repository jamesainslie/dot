package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yaklabco/dot/internal/cli/output"
	"github.com/yaklabco/dot/internal/gitsync"
	"github.com/yaklabco/dot/pkg/dot"
)

// maxListedCommits caps how many pulled commit subjects are printed.
const maxListedCommits = 10

// newSyncCommand creates the sync command.
func newSyncCommand() *cobra.Command {
	var push bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Reconcile this machine with the dotfiles repository",
		Long: `Reconcile this machine with the dotfiles repository.

sync is the single verb for multi-machine work. In the package directory,
which must be a git repository, it:

  1. Pulls with rebase and autostash, reporting the commits gained
  2. Remanages every package recorded in the manifest, so files added or
     removed upstream are linked or unlinked
  3. Prints how many packages are healthy
  4. Warns about uncommitted local edits, grouped by package

Edits made through a symlink land in the package directory, which leaves the
repository dirty. The warning block surfaces those edits; --push commits and
pushes them.

If the rebase stops with conflicts, sync prints the conflicted paths and
exits non-zero without attempting to resolve them. Resolve them in the
package directory, then run dot sync again.`,
		Example: `  # Pull, relink, and report local state
  dot sync

  # Also commit and push local edits
  dot sync --push

  # Preview without touching git
  dot --dry-run sync`,
		Args: argsWithUsage(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSync(cmd, push)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&push, "push", false, "commit and push local changes after syncing")

	return cmd
}

// runSync handles the sync command execution.
func runSync(cmd *cobra.Command, push bool) error {
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		return err
	}

	repo, err := gitsync.Open(cfg.PackageDir)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if err := repo.EnsureRepository(ctx); err != nil {
		return err
	}

	client, err := dot.NewClient(cfg)
	if err != nil {
		return err
	}

	colorize := shouldUseColor()
	formatter := output.NewFormatter(cmd.OutOrStdout(), colorize)
	errFormatter := output.NewFormatter(cmd.ErrOrStderr(), colorize)

	if err := syncPull(ctx, formatter, errFormatter, repo, cfg.DryRun); err != nil {
		return err
	}

	if err := syncRelink(ctx, formatter, client); err != nil {
		return err
	}

	return syncWorkingTree(ctx, formatter, repo, cfg.DryRun, push)
}

// syncPull pulls with rebase and reports the commits gained. Rebase conflicts
// are reported through errFormatter and returned so the caller exits non-zero.
func syncPull(ctx context.Context, formatter, errFormatter *output.Formatter, repo *gitsync.Repo, dryRun bool) error {
	if dryRun {
		formatter.Info(fmt.Sprintf("Dry run: skipping git pull in %s", repo.Dir()))
		return nil
	}

	result, err := repo.Pull(ctx)
	if err != nil {
		var conflict gitsync.ErrRebaseConflict
		if errors.As(err, &conflict) {
			reportRebaseConflict(errFormatter, conflict)
		}
		return err
	}

	reportPull(formatter, result)
	return nil
}

// reportPull prints the outcome of a pull.
func reportPull(formatter *output.Formatter, result gitsync.PullResult) {
	if result.UpToDate() {
		formatter.Info("Already up to date")
		return
	}

	formatter.Success("pulled", len(result.Commits), "commit", "commits")
	if len(result.Commits) > maxListedCommits {
		return
	}
	for _, commit := range result.Commits {
		formatter.Bullet(fmt.Sprintf("%s %s", commit.Hash, commit.Subject))
	}
}

// reportRebaseConflict prints the conflicted paths and how to recover.
func reportRebaseConflict(errFormatter *output.Formatter, conflict gitsync.ErrRebaseConflict) {
	errFormatter.Error("Rebase stopped with conflicts")
	for _, path := range conflict.Paths {
		errFormatter.Bullet(path)
	}
	errFormatter.Indent(1, fmt.Sprintf("Resolve the conflicts in %s, then run 'dot sync' again.", conflict.Dir))
}

// syncRelink remanages every package in the manifest and reports package health.
// A run with nothing to do stays silent apart from the health summary.
func syncRelink(ctx context.Context, formatter *output.Formatter, client *dot.Client) error {
	installed, err := client.List(ctx)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(installed))
	for _, pkg := range installed {
		names = append(names, pkg.Name)
	}

	if len(names) > 0 {
		err = client.Remanage(ctx, names...)
		var noChanges dot.ErrNoChanges
		switch {
		case err == nil:
			formatter.Success("remanaged", len(names), "package", "packages")
		case errors.As(err, &noChanges):
			// Nothing changed: stay quiet.
		default:
			return err
		}
	}

	return reportHealth(ctx, formatter, client)
}

// reportHealth prints the healthy-package summary line.
func reportHealth(ctx context.Context, formatter *output.Formatter, client *dot.Client) error {
	status, err := client.Status(ctx)
	if err != nil {
		return err
	}

	healthy := 0
	for _, pkg := range status.Packages {
		if pkg.IsHealthy {
			healthy++
		}
	}

	formatter.Info(fmt.Sprintf("Packages: %d/%d healthy", healthy, len(status.Packages)))
	return nil
}

// syncWorkingTree reports uncommitted local edits and, when push is set,
// commits and pushes them.
func syncWorkingTree(ctx context.Context, formatter *output.Formatter, repo *gitsync.Repo, dryRun, push bool) error {
	entries, err := repo.Status(ctx)
	if err != nil {
		return err
	}

	groups := gitsync.GroupByPackage(entries)
	if len(groups) == 0 {
		if push {
			formatter.Info("Nothing to commit; the package directory is clean")
		}
		return nil
	}

	reportDirty(formatter, repo.Dir(), groups, push)

	if !push {
		return nil
	}

	if dryRun {
		formatter.Info("Dry run: skipping commit and push")
		return nil
	}

	message := gitsync.CommitMessage(currentHostname(), groups)
	if err := repo.CommitAll(ctx, message); err != nil {
		return err
	}
	formatter.SuccessSimple(fmt.Sprintf("Committed %s: %s",
		formatCount(len(groups), "package", "packages"), message))

	if err := repo.Push(ctx); err != nil {
		return err
	}
	formatter.SuccessSimple("Pushed to remote")

	return nil
}

// reportDirty prints the uncommitted changes grouped by package.
func reportDirty(formatter *output.Formatter, dir string, groups []gitsync.PackageGroup, push bool) {
	formatter.Warning(fmt.Sprintf("Uncommitted changes in %s", dir))
	for _, group := range groups {
		formatter.Indent(1, group.Label())
		for _, entry := range group.Entries {
			formatter.Indent(2, fmt.Sprintf("%s %s", entry.Code(), entry.Path))
		}
	}
	if !push {
		formatter.Indent(1, "Run 'dot sync --push' to commit and push, or commit manually.")
	}
}

// currentHostname returns this machine's hostname, or an empty string when it
// cannot be determined.
func currentHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}
