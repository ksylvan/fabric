package changelog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielmiessler/fabric/cmd/generate_changelog/internal/git"
	"github.com/danielmiessler/fabric/cmd/generate_changelog/internal/github"
)

var (
	mergePatterns     []*regexp.Regexp
	mergePatternsOnce sync.Once
)

// getMergePatterns returns the compiled merge patterns, initializing them lazily
func getMergePatterns() []*regexp.Regexp {
	mergePatternsOnce.Do(func() {
		mergePatterns = []*regexp.Regexp{
			regexp.MustCompile(`^Merge pull request #\d+`),      // "Merge pull request #123 from..."
			regexp.MustCompile(`^Merge branch '.*' into .*`),    // "Merge branch 'feature' into main"
			regexp.MustCompile(`^Merge remote-tracking branch`), // "Merge remote-tracking branch..."
			regexp.MustCompile(`^Merge '.*' into .*`),           // "Merge 'feature' into main"
		}
	})
	return mergePatterns
}

// isMergeCommit determines if a commit is a merge commit based on its parents and message patterns.
func isMergeCommit(commit github.PRCommit) bool {
	// Primary method: Check parent count (merge commits have multiple parents)
	if len(commit.Parents) > 1 {
		return true
	}

	// Fallback method: Check commit message patterns
	mergePatterns := getMergePatterns()
	for _, pattern := range mergePatterns {
		if pattern.MatchString(commit.Message) {
			return true
		}
	}

	return false
}

// calculateVersionDate determines the version date based on the most recent commit date from the provided PRs.
//
// If no valid commit dates are found, the function falls back to the current time.
// The function iterates through the provided PRs and their associated commits, comparing commit dates
// to identify the most recent one. If a valid date is found, it is returned; otherwise, the fallback is used.
func calculateVersionDate(fetchedPRs []*github.PR) time.Time {
	versionDate := time.Now() // fallback to current time
	if len(fetchedPRs) > 0 {
		var mostRecentCommitDate time.Time
		for _, pr := range fetchedPRs {
			for _, commit := range pr.Commits {
				if commit.Date.After(mostRecentCommitDate) {
					mostRecentCommitDate = commit.Date
				}
			}
		}
		if !mostRecentCommitDate.IsZero() {
			versionDate = mostRecentCommitDate
		}
	}
	return versionDate
}

// ProcessIncomingPR processes a single PR for changelog entry creation
func (g *Generator) ProcessIncomingPR(prNumber int) error {
	if err := g.validatePRState(prNumber); err != nil {
		return fmt.Errorf("pr validation failed: %w", err)
	}

	if err := g.validateGitStatus(); err != nil {
		return fmt.Errorf("git status validation failed: %w", err)
	}

	// Now fetch the full PR with commits for content generation
	pr, err := g.ghClient.GetPRWithCommits(prNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch PR %d: %w", prNumber, err)
	}

	content := g.formatPR(pr)

	if g.cfg.EnableAISummary {
		aiContent, err := SummarizeVersionContent(content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: AI summarization failed: %v\n", err)
		} else if !checkForAIError(aiContent) {
			content = strings.TrimSpace(aiContent)
		}
	}

	if err := g.ensureIncomingDir(); err != nil {
		return fmt.Errorf("failed to create incoming directory: %w", err)
	}

	filename := filepath.Join(g.cfg.IncomingDir, fmt.Sprintf("%d.txt", prNumber))

	// Ensure content ends with a single newline
	content = strings.TrimSpace(content) + "\n"

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write incoming file: %w", err)
	}

	if err := g.commitAndPushIncoming(prNumber, filename); err != nil {
		return fmt.Errorf("failed to commit and push: %w", err)
	}

	fmt.Printf("Successfully created incoming changelog entry: %s\n", filename)
	return nil
}

// CreateNewChangelogEntry aggregates all incoming PR files for release and includes direct commits
func (g *Generator) CreateNewChangelogEntry(version string) error {
	files, err := filepath.Glob(filepath.Join(g.cfg.IncomingDir, "*.txt"))
	if err != nil {
		return fmt.Errorf("failed to scan incoming directory: %w", err)
	}

	content, err := g.aggregateIncomingPRFiles(files)
	if err != nil {
		return err
	}

	processedPRs, processedCommitSHAs, fetchedPRs, prNumbers := g.extractPRDataFromFiles(files)

	directCommitsContent, err := g.getDirectCommitsSinceLastRelease(processedPRs, processedCommitSHAs)
	if err != nil {
		return fmt.Errorf("failed to get direct commits since last release: %w", err)
	}
	if directCommitsContent != "" {
		if content.Len() > 0 {
			content.WriteString("\n")
		}
		content.WriteString(directCommitsContent)
	}

	if content.Len() == 0 {
		g.reportNoContent(files)
		return nil
	}

	versionDate := calculateVersionDate(fetchedPRs)
	entry := fmt.Sprintf("## %s (%s)\n\n%s",
		version, versionDate.Format("2006-01-02"), strings.TrimLeft(content.String(), "\n"))

	if err := g.insertVersionAtTop(entry); err != nil {
		return fmt.Errorf("failed to update CHANGELOG.md: %w", err)
	}

	if err := g.cacheVersionData(version, versionDate, fetchedPRs, prNumbers, content.String()); err != nil {
		return err
	}

	if err := g.cleanupProcessedFiles(files); err != nil {
		return err
	}

	if err := g.stageChangesForRelease(); err != nil {
		return fmt.Errorf("critical: failed to stage changes for release: %w", err)
	}

	fmt.Printf("Successfully processed %d incoming PR files for version %s\n", len(files), version)
	return nil
}

// aggregateIncomingPRFiles reads and combines content from all incoming PR files
func (g *Generator) aggregateIncomingPRFiles(files []string) (*strings.Builder, error) {
	var content strings.Builder
	var processingErrors []string

	for i, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			processingErrors = append(processingErrors, fmt.Sprintf("failed to read %s: %v", file, err))
			continue
		}
		content.WriteString(string(data))
		if i < len(files)-1 {
			content.WriteString("\n")
		}
	}

	if len(processingErrors) > 0 {
		return nil, fmt.Errorf("encountered errors while processing incoming files: %s", strings.Join(processingErrors, "; "))
	}

	return &content, nil
}

// reportNoContent prints an appropriate message when no content is available
func (g *Generator) reportNoContent(files []string) {
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No incoming PR files found in %s and no direct commits since last release\n", g.cfg.IncomingDir)
	} else {
		fmt.Fprintf(os.Stderr, "No content found in incoming files and no direct commits since last release\n")
	}
}

// cacheVersionData saves PR and version data to the cache database
func (g *Generator) cacheVersionData(version string, versionDate time.Time, fetchedPRs []*github.PR, prNumbers []int, contentStr string) error {
	if g.cache == nil {
		return nil
	}

	if len(fetchedPRs) > 0 {
		if err := g.cache.SavePRBatch(fetchedPRs); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save PR batch to cache: %v\n", err)
		}

		if err := g.cache.SaveCommitPRMappings(fetchedPRs); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to cache commit mappings: %v\n", err)
		}

		for _, pr := range fetchedPRs {
			if err := g.cacheCommitsForPR(pr, version); err != nil {
				return err
			}
		}
	}

	newVersionEntry := &git.Version{
		Name:      version,
		Date:      versionDate,
		CommitSHA: "",
		PRNumbers: prNumbers,
		AISummary: contentStr,
	}

	if err := g.cache.SaveVersion(newVersionEntry); err != nil {
		return fmt.Errorf("failed to save new version entry to database: %w", err)
	}

	return nil
}

// cacheCommitsForPR saves all commits from a PR to the cache
func (g *Generator) cacheCommitsForPR(pr *github.PR, version string) error {
	for _, commit := range pr.Commits {
		commitDate := commit.Date
		if commitDate.IsZero() {
			commitDate = time.Now()
			fmt.Fprintf(os.Stderr, "Warning: Commit %s has invalid timestamp, using current time as fallback\n", commit.SHA)
		}

		gitCommit := &git.Commit{
			SHA:      commit.SHA,
			Message:  commit.Message,
			Author:   commit.Author,
			Email:    commit.Email,
			Date:     commitDate,
			IsMerge:  isMergeCommit(commit),
			PRNumber: pr.Number,
		}
		if err := g.cache.SaveCommit(gitCommit, version); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save commit %s to cache: %v\n", commit.SHA, err)
		}
	}
	return nil
}

// cleanupProcessedFiles removes processed incoming files from git and filesystem
func (g *Generator) cleanupProcessedFiles(files []string) error {
	for _, file := range files {
		relativeFile, err := filepath.Rel(g.cfg.RepoPath, file)
		if err != nil {
			relativeFile = file
		}

		if err := g.gitWalker.RemoveFile(relativeFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove %s from git index: %v\n", relativeFile, err)
			if err := os.Remove(file); err != nil {
				g.reportFileRemovalFailure(file, relativeFile, err)
			}
		}
	}
	return nil
}

// reportFileRemovalFailure prints detailed error messages for file removal failures
func (g *Generator) reportFileRemovalFailure(file, relativeFile string, err error) {
	fmt.Fprintf(os.Stderr, "Error: Failed to remove %s from the filesystem after failing to remove it from the git index.\n", relativeFile)
	fmt.Fprintf(os.Stderr, "Filesystem error: %v\n", err)
	fmt.Fprintf(os.Stderr, "Manual intervention required:\n")
	fmt.Fprintf(os.Stderr, "  1. Remove the file %s manually (using the OS-specific command)\n", file)
	fmt.Fprintf(os.Stderr, "  2. Remove from git index: git rm --cached %s\n", relativeFile)
	fmt.Fprintf(os.Stderr, "  3. Or reset git index: git reset HEAD %s\n", relativeFile)
}

// extractPRDataFromFiles extracts PR numbers and commit data from incoming files
func (g *Generator) extractPRDataFromFiles(files []string) (map[int]bool, map[string]bool, []*github.PR, []int) {
	processedPRs := make(map[int]bool)
	processedCommitSHAs := make(map[string]bool)
	var fetchedPRs []*github.PR
	var prNumbers []int

	for _, file := range files {
		prNum, pr := g.processSinglePRFile(file)
		if prNum == 0 {
			continue // Invalid file, skip
		}

		processedPRs[prNum] = true
		prNumbers = append(prNumbers, prNum)

		if pr != nil {
			fetchedPRs = append(fetchedPRs, pr)
			g.recordCommitSHAs(pr, processedCommitSHAs)
		}
	}

	return processedPRs, processedCommitSHAs, fetchedPRs, prNumbers
}

// processSinglePRFile extracts PR number from filename and fetches PR data
func (g *Generator) processSinglePRFile(file string) (int, *github.PR) {
	filename := filepath.Base(file)
	prNumStr, ok := strings.CutSuffix(filename, ".txt")
	if !ok {
		return 0, nil
	}

	prNum, err := strconv.Atoi(prNumStr)
	if err != nil {
		return 0, nil
	}

	pr, err := g.ghClient.GetPRWithCommits(prNum)
	if err != nil {
		return prNum, nil // Return PR number even if fetch fails
	}

	return prNum, pr
}

// recordCommitSHAs records all commit SHAs from a PR
func (g *Generator) recordCommitSHAs(pr *github.PR, processedCommitSHAs map[string]bool) {
	for _, commit := range pr.Commits {
		processedCommitSHAs[commit.SHA] = true
	}
}

// getDirectCommitsSinceLastRelease gets all direct commits (not part of PRs) since the last release
func (g *Generator) getDirectCommitsSinceLastRelease(processedPRs map[int]bool, processedCommitSHAs map[string]bool) (string, error) {
	// Get the latest tag to determine what commits are unreleased
	latestTag, err := g.gitWalker.GetLatestTag()
	if err != nil {
		return "", fmt.Errorf("failed to get latest tag: %w", err)
	}

	// Get all commits since the latest tag
	unreleasedVersion, err := g.gitWalker.WalkCommitsSinceTag(latestTag)
	if err != nil {
		return "", fmt.Errorf("failed to walk commits since tag %s: %w", latestTag, err)
	}

	if unreleasedVersion == nil || len(unreleasedVersion.Commits) == 0 {
		return "", nil // No unreleased commits
	}

	// Filter out commits that are part of PRs (we already have those from incoming files)
	// and format the direct commits
	directCommits := g.filterDirectCommits(unreleasedVersion.Commits, processedPRs, processedCommitSHAs)

	if len(directCommits) == 0 {
		return "", nil // No direct commits
	}

	// Format the direct commits similar to how it's done in generateRawVersionContent
	var sb strings.Builder
	sb.WriteString("### Direct commits\n\n")

	// Sort direct commits by date (newest first) for consistent ordering
	sort.Slice(directCommits, func(i, j int) bool {
		return directCommits[i].Date.After(directCommits[j].Date)
	})

	for _, commit := range directCommits {
		message := g.formatCommitMessage(strings.TrimSpace(commit.Message))
		if message != "" && !g.isDuplicateMessage(message, directCommits) {
			sb.WriteString(fmt.Sprintf("- %s\n", message))
		}
	}

	return sb.String(), nil
}

// filterDirectCommits filters commits to only include direct commits (not part of PRs)
func (g *Generator) filterDirectCommits(commits []*git.Commit, processedPRs map[int]bool, processedCommitSHAs map[string]bool) []*git.Commit {
	var directCommits []*git.Commit

	for _, commit := range commits {
		if !g.isDirectCommit(commit, processedPRs, processedCommitSHAs) {
			continue
		}
		directCommits = append(directCommits, commit)
	}

	return directCommits
}

// isDirectCommit checks if a commit is a direct commit (not part of a PR)
func (g *Generator) isDirectCommit(commit *git.Commit, processedPRs map[int]bool, processedCommitSHAs map[string]bool) bool {
	// Skip version bump commits
	if commit.IsVersion {
		return false
	}

	// Skip commits that belong to already-processed PRs
	if commit.PRNumber > 0 && processedPRs[commit.PRNumber] {
		return false
	}

	// Skip commits whose SHA is in processed PRs
	if processedCommitSHAs[commit.SHA] {
		return false
	}

	// Only include commits that are NOT part of any PR
	return commit.PRNumber == 0
}

// validatePRState validates that a PR is in the correct state for processing
func (g *Generator) validatePRState(prNumber int) error {
	// Use lightweight validation call that doesn't fetch commits
	details, err := g.ghClient.GetPRValidationDetails(prNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch pr %d: %w", prNumber, err)
	}

	if details.State != "open" {
		return fmt.Errorf("pr %d is not open (current state: %s)", prNumber, details.State)
	}

	if !details.Mergeable {
		return fmt.Errorf("pr %d is not mergeable: please resolve conflicts first", prNumber)
	}

	return nil
}

// validateGitStatus ensures the working directory is clean
func (g *Generator) validateGitStatus() error {
	isClean, err := g.gitWalker.IsWorkingDirectoryClean()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if !isClean {
		// Get detailed status for better error message
		statusDetails, statusErr := g.gitWalker.GetStatusDetails()
		if statusErr == nil && statusDetails != "" {
			return fmt.Errorf("working directory is not clean - please commit or stash changes before proceeding:\n%s", statusDetails)
		}
		return fmt.Errorf("working directory is not clean - please commit or stash changes before proceeding")
	}

	return nil
}

// ensureIncomingDir creates the incoming directory if it doesn't exist
func (g *Generator) ensureIncomingDir() error {
	if err := os.MkdirAll(g.cfg.IncomingDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", g.cfg.IncomingDir, err)
	}
	return nil
}

// commitAndPushIncoming commits and optionally pushes the incoming changelog file
func (g *Generator) commitAndPushIncoming(prNumber int, filename string) error {
	relativeFilename, err := filepath.Rel(g.cfg.RepoPath, filename)
	if err != nil {
		relativeFilename = filename
	}

	// Add file to git index
	if err := g.gitWalker.AddFile(relativeFilename); err != nil {
		return fmt.Errorf("failed to add file %s: %w", relativeFilename, err)
	}

	// Commit changes
	commitMessage := fmt.Sprintf("chore: incoming %d changelog entry", prNumber)
	_, err = g.gitWalker.CommitChanges(commitMessage)
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push to remote if enabled
	if g.cfg.Push {
		if err := g.gitWalker.PushToRemote(); err != nil {
			return fmt.Errorf("failed to push to remote: %w", err)
		}
	} else {
		fmt.Println("Commit created successfully. Please review and push manually.")
	}

	return nil
}

// detectVersion detects the current version from version.nix or git tags
func (g *Generator) detectVersion() (string, error) {
	versionNixPath := filepath.Join(g.cfg.RepoPath, "version.nix")
	if _, err := os.Stat(versionNixPath); err == nil {
		data, err := os.ReadFile(versionNixPath)
		if err != nil {
			return "", fmt.Errorf("failed to read version.nix: %w", err)
		}

		versionRegex := regexp.MustCompile(`"([^"]+)"`)
		matches := versionRegex.FindStringSubmatch(string(data))
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	latestTag, err := g.gitWalker.GetLatestTag()
	if err != nil {
		return "", fmt.Errorf("failed to get latest tag: %w", err)
	}

	if latestTag == "" {
		return "v1.0.0", nil
	}

	return latestTag, nil
}

// insertVersionAtTop inserts a new version entry at the top of CHANGELOG.md
func (g *Generator) insertVersionAtTop(entry string) error {
	changelogPath := filepath.Join(g.cfg.RepoPath, "CHANGELOG.md")
	header := "# Changelog"
	headerRegex := regexp.MustCompile(`(?m)^# Changelog\s*`)

	existingContent, err := os.ReadFile(changelogPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read existing CHANGELOG.md: %w", err)
		}
		// File doesn't exist, create it.
		newContent := fmt.Sprintf("%s\n\n%s\n", header, entry)
		return os.WriteFile(changelogPath, []byte(newContent), 0644)
	}

	contentStr := string(existingContent)
	var newContent string

	if loc := headerRegex.FindStringIndex(contentStr); loc != nil {
		// Found the header, insert after it.
		insertionPoint := loc[1]
		// Skip any existing newlines after the header to avoid double spacing
		for insertionPoint < len(contentStr) && (contentStr[insertionPoint] == '\n' || contentStr[insertionPoint] == '\r') {
			insertionPoint++
		}
		// Insert with proper spacing: single newline after header, then entry, then newline before existing content
		newContent = contentStr[:loc[1]] + entry + "\n" + contentStr[insertionPoint:]
	} else {
		// Header not found, prepend everything.
		newContent = fmt.Sprintf("%s\n\n%s\n\n%s", header, entry, contentStr)
	}

	return os.WriteFile(changelogPath, []byte(newContent), 0644)
}

// stageChangesForRelease stages the modified files for the release commit
func (g *Generator) stageChangesForRelease() error {
	changelogPath := filepath.Join(g.cfg.RepoPath, "CHANGELOG.md")
	relativeChangelog, err := filepath.Rel(g.cfg.RepoPath, changelogPath)
	if err != nil {
		relativeChangelog = "CHANGELOG.md"
	}

	relativeCacheFile, err := filepath.Rel(g.cfg.RepoPath, g.cfg.CacheFile)
	if err != nil {
		relativeCacheFile = g.cfg.CacheFile
	}

	// Add CHANGELOG.md to git index
	if err := g.gitWalker.AddFile(relativeChangelog); err != nil {
		return fmt.Errorf("failed to add %s: %w", relativeChangelog, err)
	}

	// Add cache file to git index
	if err := g.gitWalker.AddFile(relativeCacheFile); err != nil {
		return fmt.Errorf("failed to add %s: %w", relativeCacheFile, err)
	}

	// Note: Individual incoming files are now removed during the main processing loop
	// No need to remove the entire directory here

	return nil
}
