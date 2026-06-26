package cmd

import (
	"fmt"
	"os"

	"github.com/NexusSoftMDMA/TormentNexus/agent"
	"github.com/spf13/cobra"
)

var pipeCmd = &cobra.Command{
	Use:   "pipe [prompt]",
	Short: "Process piped input through the LLM",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		a := agent.NewAgent()
		result, err := a.ProcessPipe(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing pipe: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(pipeCmd)
}
