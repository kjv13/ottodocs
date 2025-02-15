/*
Copyright © 2023 Chandler <chandler@chand1012.dev>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chand1012/git2gpt/prompt"
	"github.com/chand1012/memory"
	l "github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/chand1012/ottodocs/pkg/ai"
	"github.com/chand1012/ottodocs/pkg/calc"
	"github.com/chand1012/ottodocs/pkg/config"
	"github.com/chand1012/ottodocs/pkg/git"
	"github.com/chand1012/ottodocs/pkg/utils"
)

// move this to memory package eventually
func sortByScore(fragments []memory.MemoryFragment) []memory.MemoryFragment {
	sort.Slice(fragments, func(i, j int) bool {
		return fragments[i].Score > fragments[j].Score
	})
	return fragments
}

// askCmd represents the ask command
var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Ask a question about a file or repo",
	Long: `Uses full text search to find relevant code and ask questions about said code.
Requires a path to a repository or file as a positional argument.`,
	Args: cobra.PositionalArgs(func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires a path to a repository or file")
		}
		return nil
	}),
	PreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetLevel(l.DebugLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var answer string
		repoPath := args[0]
		var fileName string
		conf, err := config.Load()
		if err != nil || conf.APIKey == "" {
			// if the API key is not set, prompt the user to login
			log.Error("Please login first.")
			log.Error("Run `ottodocs login` to login.")
			os.Exit(1)
		}

		if chatPrompt == "" {
			log.Debug("User did not enter a question. Prompting for one...")
			fmt.Println("Please enter a question: ")
			fmt.Scanln(&chatPrompt)
			// strip the newline character
			chatPrompt = strings.TrimRight(chatPrompt, " \n")
		}

		log.Debug("Getting file contents...")
		info, err := os.Stat(repoPath)
		if err != nil {
			log.Errorf("Error getting file info: %s", err)
			os.Exit(1)
		}
		// check if the first arg is a file or a directory
		// if it's a file, ask a question about that file directly
		if info.IsDir() {
			log.Debug("Constructing repo memory...")
			// Define a temporary path for the index file
			testIndexPath := filepath.Join(args[0], ".index.memory")

			log.Debug("Creating memory index...")
			// Create a new memory index
			m, _, err := memory.New(testIndexPath)
			if err != nil {
				log.Errorf("Failed to create memory index: %s", err)
				os.Exit(1)
			}

			log.Debug("Indexing repo...")
			// index the repo
			repo, err := git.GetRepo(repoPath, ignoreFilePath, ignoreGitignore)
			if err != nil {
				log.Errorf("Error processing repo: %s", err)
				os.Exit(1)
			}

			// index the files
			for _, file := range repo.Files {
				err = m.Add(file.Path, file.Contents)
				if err != nil {
					log.Errorf("Error indexing file: %s", err)
					os.Exit(1)
				}
			}

			log.Debug("Searching memory index...")
			// search the memory index
			results, err := m.Search(chatPrompt)
			if err != nil {
				log.Errorf("Failed to search memory index: %s", err)
				os.Exit(1)
			}

			log.Debug("Results extracted. Destroying memory index...")
			// close the memory index
			m.Destroy()

			log.Debug("Sorting results...")
			sortedFragments := sortByScore(results)

			log.Debug("Getting file contents...")
			var files []prompt.GitFile
			for _, result := range sortedFragments {
				for _, file := range repo.Files {
					if file.Path == result.ID {
						files = append(files, file)
					}
				}
			}

			if len(files) == 0 {
				log.Error("No results found.")
				os.Exit(1)
			}

			log.Debug("Asking chatGPT question...")
			answer, err = ai.Question(files, chatPrompt, conf)
			if err != nil {
				log.Errorf("Error asking question: %s", err)
				os.Exit(1)
			}
		} else {
			log.Debug("Getting file contents...")
			fileName = repoPath
			content, err := utils.LoadFile(fileName)
			if err != nil {
				log.Errorf("Error loading file: %s", err)
				os.Exit(1)
			}

			log.Debug("Calculating tokens and constructing file...")
			tokens, err := calc.PreciseTokens(content)
			if err != nil {
				log.Errorf("Error calculating tokens: %s", err)
				os.Exit(1)
			}

			files := []prompt.GitFile{
				{
					Path:     fileName,
					Contents: content,
					Tokens:   int64(tokens),
				},
			}

			log.Debug("Asking chatGPT question...")
			answer, err = ai.Question(files, chatPrompt, conf)
			if err != nil {
				log.Errorf("Error asking question: %s", err)
				os.Exit(1)
			}
		}

		fmt.Println("Answer:", answer)
	},
}

func init() {
	RootCmd.AddCommand(askCmd)
	askCmd.Flags().StringVarP(&chatPrompt, "question", "q", "", "The question to ask")
	askCmd.Flags().BoolVarP(&ignoreGitignore, "ignore-gitignore", "g", false, "ignore .gitignore file")
	askCmd.Flags().StringVarP(&ignoreFilePath, "ignore", "n", "", "path to .gptignore file")
	askCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
