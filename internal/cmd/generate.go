package cmd

import (
	"fmt"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/generator"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates work load",
	RunE: func(cmd *cobra.Command, args []string) error {
		startOffset, err := cmd.Flags().GetInt("start-offset")
		if err != nil {
			return err
		}

		maxCount, err := cmd.Flags().GetInt("max-count")
		if err != nil {
			return err
		}

		maxOffset, err := cmd.Flags().GetInt("max-offset")
		if err != nil {
			return err
		}

		minInterval, err := cmd.Flags().GetInt("min-interval")
		if err != nil {
			return err
		}

		maxInterval, err := cmd.Flags().GetInt("max-interval")
		if err != nil {
			return err
		}

		indexToSearchRatio, err := cmd.Flags().GetInt("index-search-ratio")
		if err != nil {
			return err
		}

		g := generator.NewGenerator(generator.Config{
			StartOffsetSeconds: startOffset,
			MaxCount:           maxCount,
			MaxOffsetSeconds:   maxOffset,
			MinIntervalSeconds: minInterval,
			MaxIntervalSeconds: maxInterval,
			IndexToSearchRatio: indexToSearchRatio,
		})

		buf, err := g.Generate()
		fmt.Println(string(buf))

		return nil
	},
}

func init() {
	generateCmd.Flags().IntP("start-offset", "s", 0, "start offset, in seconds")
	generateCmd.Flags().IntP("max-count", "c", 1200, "max. number of operations")
	generateCmd.Flags().IntP("max-offset", "o", 3500, "max. offset, in seconds")
	generateCmd.Flags().IntP("min-interval", "i", 0, "min. interval between operations, in seconds")
	generateCmd.Flags().IntP("max-interval", "n", 3, "max. interval between operations, in seconds")
	generateCmd.Flags().IntP("index-search-ratio", "r", 4, "ratio of indexing to search operations")
}
