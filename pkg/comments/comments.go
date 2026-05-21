package comments

import (
	"bytes"
	"io"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/augmentable-dev/lege"
	"github.com/go-enry/go-enry/v2"
	"github.com/karrick/godirwalk"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Comments is a list of comments
type Comments []*Comment

// DefaultIgnorePatterns skips common repository metadata, dependency, and build paths.
var DefaultIgnorePatterns = []string{
	".git",
	".hg",
	".svn",
	"node_modules",
	"vendor",
	"dist",
	"build",
	"target",
	"bin",
	"obj",
	".terraform",
	".venv",
	"venv",
	"__pycache__",
	".next",
	"coverage",
	".github/tickgit-*.csv",
	".github/tickgit-candidates.md",
	"tickgit-*.csv",
	"tickgit-candidates.md",
}

// SearchOptions configures directory comment searches.
type SearchOptions struct {
	IgnorePatterns []string
}

// Comment represents a comment in a source code file
type Comment struct {
	lege.Collection
	FilePath string
}

// SearchFile searches a file for comments. It infers the language
func SearchFile(filePath string, reader io.Reader, cb func(*Comment)) error {
	slashPath := filepath.ToSlash(filePath)
	if enry.IsVendor(slashPath) {
		return nil
	}

	// create a preview reader that reads in some of the file for enry to better identify the language
	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)
	previewReader := io.LimitReader(tee, 1000)
	preview, err := io.ReadAll(previewReader)
	if err != nil {
		return err
	}

	// create a new reader concatenating the preview and the original reader (which has now been read from)
	fullReader := io.MultiReader(strings.NewReader(buf.String()), reader)

	lang := Language(enry.GetLanguage(filepath.Base(filePath), preview))
	if lang == "Markdown" {
		return searchMarkdownFile(filePath, fullReader, cb)
	}

	options, ok := LanguageParseOptions[lang]
	if !ok {
		options = CStyleCommentOptions
	}
	commentParser, err := lege.NewParser(options)
	if err != nil {
		return err
	}

	collections, err := commentParser.Parse(fullReader)
	if err != nil {
		return err
	}

	for _, c := range collections {
		comment := Comment{*c, filePath}
		cb(&comment)
	}

	return nil
}

func searchMarkdownFile(filePath string, reader io.Reader, cb func(*Comment)) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	for index, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineNumber := index + 1
		collection := lege.NewCollection(
			lege.Location{Line: lineNumber, Pos: 1},
			lege.Location{Line: lineNumber, Pos: len(line)},
			lege.Boundary{},
			line,
		)
		comment := Comment{*collection, filePath}
		cb(&comment)
	}

	return nil
}

// SearchDir searches a directory for comments
func SearchDir(dirPath string, cb func(comment *Comment)) error {
	return SearchDirWithOptions(dirPath, SearchOptions{}, cb)
}

// SearchDirWithOptions searches a directory for comments with explicit options.
func SearchDirWithOptions(dirPath string, options SearchOptions, cb func(comment *Comment)) error {
	ignorePatterns := append([]string{}, DefaultIgnorePatterns...)
	ignorePatterns = append(ignorePatterns, options.IgnorePatterns...)

	err := godirwalk.Walk(dirPath, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			localPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			ignored, err := matchesIgnorePattern(localPath, ignorePatterns)
			if err != nil {
				return err
			}
			if ignored && de.IsDir() {
				return filepath.SkipDir
			}
			if ignored {
				return nil
			}
			if de.IsRegular() {
				p, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				f, err := os.Open(p)
				if err != nil {
					return err
				}
				err = SearchFile(localPath, f, cb)
				closeErr := f.Close()
				if err != nil {
					return err
				}
				if closeErr != nil {
					return closeErr
				}
			}
			return nil
		},
		Unsorted: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func matchesIgnorePattern(localPath string, patterns []string) (bool, error) {
	slashPath := filepath.ToSlash(localPath)
	if slashPath == "." {
		return false, nil
	}

	components := strings.Split(slashPath, "/")
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(filepath.ToSlash(pattern))
		pattern = strings.Trim(pattern, "/")
		if pattern == "" {
			continue
		}

		if strings.Contains(pattern, "/") {
			matched, err := pathpkg.Match(pattern, slashPath)
			if err != nil {
				return false, err
			}
			if matched || slashPath == pattern || strings.HasPrefix(slashPath, pattern+"/") {
				return true, nil
			}
			continue
		}

		for _, component := range components {
			matched, err := pathpkg.Match(pattern, component)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
	}

	return false, nil
}

// SearchCommit searches all files in the tree of a given commit
func SearchCommit(commit *object.Commit, cb func(*Comment)) error {
	var wg sync.WaitGroup
	errs := make(chan error)

	fileIter, err := commit.Files()
	if err != nil {
		return err
	}
	defer fileIter.Close()
	err = fileIter.ForEach(func(file *object.File) error {
		if file.Mode.IsFile() {
			wg.Add(1)
			go func() {
				defer wg.Done()

				r, err := file.Reader()
				if err != nil {
					errs <- err
					return
				}
				err = SearchFile(file.Name, r, cb)
				if err != nil {
					errs <- err
					return
				}

			}()
		}
		return nil
	})

	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}
