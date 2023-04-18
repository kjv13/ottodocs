/*
Copyright © 2023 Chandler <chandler@chand1012.dev>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chand1012/ottodocs/ai"
	"github.com/chand1012/ottodocs/config"
	"github.com/chand1012/ottodocs/utils"
	"github.com/spf13/cobra"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Document a repository of files",
	Long: `Document an entire repository of files. Specify the path to the repo as the first positional argument. This command will recursively
search for files in the directory and document them.
	`,
	Args: cobra.PositionalArgs(func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires a path to a repository")
		}
		return nil
	}),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath = args[0]

		if (!markdownMode || inlineMode) && outputFile != "" {
			log.Error("Error: cannot specify an output file in inline mode")
			os.Exit(1)
		}

		if markdownMode && overwriteOriginal {
			log.Error("Error: cannot overwrite original file in markdown mode")
			os.Exit(1)
		}

		if markdownMode && outputFile == "" {
			log.Error("Error: must specify an output file in markdown mode")
			os.Exit(1)
		}

		if outputFile != "" {
			// if output file exists, throw error
			if _, err := os.Stat(outputFile); err == nil {
				log.Errorf("Error: output file %s already exists!", outputFile)
				os.Exit(1)
			}
		}

		conf, err := config.Load()
		if err != nil || conf.APIKey == "" {
			// if the API key is not set, prompt the user to login
			log.Error("Please login first.")
			log.Error("Run `ottodocs login` to login.")
			os.Exit(1)
		}

		repo, err := utils.GetRepo(repoPath, ignoreFilePath, ignoreGitignore)
		if err != nil {
			log.Errorf("Error: %s", err)
			os.Exit(1)
		}

		for _, file := range repo.Files {
			var contents string

			path := filepath.Join(repoPath, file.Path)

			if outputFile != "" {
				fmt.Println("Documenting file", file.Path)
			}

			if chatPrompt == "" {
				chatPrompt = "Write documentation for the following code snippet. The file name is" + file.Path + ":"
			}

			fileContents, err := utils.LoadFile(path)
			if err != nil {
				log.Warnf("Error loading file %s: %s", path, err)
				continue
			}

			if inlineMode || !markdownMode {
				contents, err = ai.SingleFile(path, fileContents, chatPrompt, conf.APIKey, conf.Model)
			} else {
				contents, err = ai.Markdown(path, fileContents, chatPrompt, conf.APIKey, conf.Model)
			}

			if err != nil {
				log.Warnf("Error documenting file %s: %s", path, err)
				continue
			}

			if outputFile != "" && markdownMode {
				// write the string to the output file
				// append if the file already exists
				file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					log.Errorf("Error: %s", err)
					os.Exit(1)
				}

				_, err = file.WriteString(contents)
				if err != nil {
					log.Errorf("Error: %s", err)
					os.Exit(1)
				}

				file.Close()
			} else if overwriteOriginal {
				// overwrite the original file
				// clear the contents of the file
				file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					log.Errorf("Error: %s", err)
					os.Exit(1)
				}

				// write the new contents to the file
				_, err = file.WriteString(contents)
				if err != nil {
					log.Errorf("Error: %s", err)
					os.Exit(1)
				}

				file.Close()
			} else {
				fmt.Println(contents)
			}
		}

	},
}

func init() {
	RootCmd.AddCommand(docsCmd)

	// see cmd/vars for the definition of these flags
	docsCmd.Flags().StringVarP(&chatPrompt, "prompt", "p", "", "Prompt to use for the ChatGPT API")
	docsCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Path to the output file. For use with --markdown")
	docsCmd.Flags().StringVarP(&ignoreFilePath, "ignore", "n", "", "path to .gptignore file")
	docsCmd.Flags().BoolVarP(&markdownMode, "markdown", "m", false, "Output in Markdown format")
	docsCmd.Flags().BoolVarP(&inlineMode, "inline", "i", false, "Output in inline format")
	docsCmd.Flags().BoolVarP(&overwriteOriginal, "overwrite", "w", false, "Overwrite the original file")
	docsCmd.Flags().BoolVarP(&ignoreGitignore, "ignore-gitignore", "g", false, "ignore .gitignore file")
}
