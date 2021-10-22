package cmd

import "github.com/spf13/cobra"

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate autocompletion for commonly used shells",
	Long: `
Completion is used to output completion code for bash and zsh shells.
Before using completion features, you have to source completion code
from your .profile or .bashrc/.zshrc file. This is done by adding
following line to one of above files:

	source <(algoid completion SHELL)

Bash users can as well save it to the file and copy it to:
	/etc/bash_completion.d/

Correct arguments for SHELL are: "bash" and "zsh".

Notes:
1) zsh completions requires zsh 5.2 or newer.

2) macOS users have to install bash-completion framework to utilize
completion features. This can be done using homebrew:
	brew install bash-completion

Once installed, you must load bash_completion by adding following
line to your .profile or .bashrc/.zshrc:
	source $(brew --prefix)/etc/bash_completion
`,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
