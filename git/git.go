package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	originNamesInLookupOrder = []string{"upstream", "github", "origin"}
	remotesRegexp            = regexp.MustCompile(`(.+)\s+(.+)\s+\((push|fetch)\)`)
)

func BranchShortName(ref string) string {
	reg := regexp.MustCompile("^refs/(remotes/)?.+?/")
	return reg.ReplaceAllString(ref, "")
}

func ConfigAll(name string) ([]string, error) {
	mode := "--get-all"
	if strings.Contains(name, "*") {
		mode = "--get-regexp"
	}

	output, err := execGit("config", mode, name)
	if err != nil {
		return nil, fmt.Errorf("unknown config %s", name)
	}
	return splitLines(output), nil
}

func CurrentBranch() (string, error) {
	head, err := Head()
	if err != nil {
		return "", fmt.Errorf("aborted: not currently on any branch")
	}
	return head, nil
}

var cachedDir string

func Dir() (string, error) {
	if cachedDir != "" {
		return cachedDir, nil
	}

	output, err := execGitQuiet("rev-parse", "-q", "--git-dir")
	if err != nil {
		return "", fmt.Errorf("not a git repository (or any of the parent directories): .git")
	}

	var chdir string
	// for i, flag := range GlobalFlags {
	// 	if flag == "-C" {
	// 		dir := GlobalFlags[i+1]
	// 		if filepath.IsAbs(dir) {
	// 			chdir = dir
	// 		} else {
	// 			chdir = filepath.Join(chdir, dir)
	// 		}
	// 	}
	// }

	gitDir := firstLine(output)

	if !filepath.IsAbs(gitDir) {
		if chdir != "" {
			gitDir = filepath.Join(chdir, gitDir)
		}

		gitDir, err = filepath.Abs(gitDir)
		if err != nil {
			return "", err
		}

		gitDir = filepath.Clean(gitDir)
	}

	cachedDir = gitDir
	return gitDir, nil
}

func DefaultBranch(remote string) string {
	if name, err := SymbolicRef(fmt.Sprintf("refs/remotes/%s/HEAD", remote)); err != nil {
		return name
	}
	return "refs/heads/main"
}

func HasFile(segments ...string) bool {
	// For Git >= 2.5.0
	if output, err := execGitQuiet("rev-parse", "-q", "--git-path", filepath.Join(segments...)); err == nil {
		if lines := splitLines(output); len(lines) == 1 {
			if _, err := os.Stat(lines[0]); err == nil {
				return true
			}
		}
	}

	return false
}

func Head() (string, error) {
	return SymbolicRef("HEAD")
}

func LocalBranches() ([]string, error) {
	output, err := execGit("branch", "--list")
	if err != nil {
		return nil, err
	}
	branches := []string{}
	for _, branch := range splitLines(output) {
		branches = append(branches, branch[2:])
	}
	return branches, nil
}

func MainRemote() (string, error) {
	remotes, err := Remotes()
	if err != nil || len(remotes) == 0 {
		return "", fmt.Errorf("aborted: no git remotes found")
	}
	return remotes[0], nil
}

func NewRange(a, b string) (*Range, error) {
	output, err := execGitQuiet("rev-parse", "-q", a, b)
	if err != nil {
		return nil, err
	}
	lines := splitLines(output)
	if len(lines) != 2 {
		return nil, fmt.Errorf("can't parse range %s..%s", a, b)
	}
	return &Range{lines[0], lines[1]}, nil
}

func Remotes() ([]string, error) {
	output, err := execGit("remote", "-v")
	if err != nil {
		return nil, fmt.Errorf("aborted: can't load git remotes")
	}

	remoteLines := splitLines(output)

	remotesMap := make(map[string]map[string]string)
	for _, r := range remoteLines {
		if remotesRegexp.MatchString(r) {
			match := remotesRegexp.FindStringSubmatch(r)
			name := strings.TrimSpace(match[1])
			url := strings.TrimSpace(match[2])
			urlType := strings.TrimSpace(match[3])
			utm, ok := remotesMap[name]
			if !ok {
				utm = make(map[string]string)
				remotesMap[name] = utm
			}
			utm[urlType] = url
		}
	}

	remotes := []string{}

	for _, name := range originNamesInLookupOrder {
		if _, ok := remotesMap[name]; ok {
			remotes = append(remotes, name)
			delete(remotesMap, name)
		}
	}

	for name := range remotesMap {
		remotes = append(remotes, name)
	}

	return remotes, nil
}

func Spawn(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Quiet(args ...string) bool {
	fmt.Printf("%v\n", args)
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}

func SymbolicFullName(name string) (string, error) {
	output, err := execGitQuiet("rev-parse", "--symbolic-full-name", name)
	if err != nil {
		return "", fmt.Errorf("unknown revision or path not in the working tree: %s", name)
	}
	return firstLine(output), nil
}

func SymbolicRef(ref string) (string, error) {
	output, err := execGit("symbolic-ref", ref)
	if err != nil {
		return "", err
	}
	return firstLine(output), err
}

func execGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	return string(output), err
}

func execGitQuiet(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	return string(output), err
}

func splitLines(output string) []string {
	output = strings.TrimSuffix(output, "\n")
	if output == "" {
		return []string{}
	}
	return strings.Split(output, "\n")
}

func firstLine(output string) string {
	if i := strings.Index(output, "\n"); i >= 0 {
		return output[0:i]
	}
	return output
}
