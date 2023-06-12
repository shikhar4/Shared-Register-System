package cmd

import (
	"cs598fts/benchmark"
	"cs598fts/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var configPath string
var clientID int

var clientCmd = &cobra.Command{
	Use:   "client [config]",
	Short: "Shared register client",
	Long:  `Shared register client, can be either reader or writer.`,
}

var readCmd = &cobra.Command{
	Use:   "read [key]",
	Short: "Shared register client read",
	Long:  `Shared register client read.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewClient(clientID, configPath)
		if err != nil {
			logrus.Fatal(err)
		}
		key := args[0]
		val, err := c.Read(key)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Read key [%s] with val [%s] ", key, val)
	},
}

var writeCmd = &cobra.Command{
	Use:   "write [key] [val]",
	Short: "Shared register client write",
	Long:  `Shared register client write.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewClient(clientID, configPath)
		if err != nil {
			logrus.Fatal(err)
		}
		key := args[0]
		val := args[1]
		if err := c.Write(key, val); err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Write key [%s] with val [%s] ", key, val)
	},
}

var clientNum int
var requestCnt int
var workloadStr string

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark [client number] [request count] [workload]",
	Short: "Shared register client benchmark",
	Long:  `Shared register client benchmark, workload can be selected from read-only, half-half, write-only.`,
	Run: func(cmd *cobra.Command, args []string) {
		b, err := benchmark.NewBenchmark(clientNum, requestCnt, workloadStr, configPath)
		if err != nil {
			logrus.Fatal(err)
		}
		if err := b.Init(); err != nil {
			logrus.Fatal(err)
		}
		if err := b.Run(); err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.PersistentFlags().IntVar(&clientID, "id", 0, "client ID")
	//_ = clientCmd.MarkPersistentFlagRequired("id")
	clientCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file")
	_ = clientCmd.MarkPersistentFlagRequired("config")

	clientCmd.AddCommand(readCmd)
	clientCmd.AddCommand(writeCmd)
	clientCmd.AddCommand(benchmarkCmd)

	readCmd.Args = cobra.ExactArgs(1)
	writeCmd.Args = cobra.ExactArgs(2)

	benchmarkCmd.PersistentFlags().IntVar(&clientNum, "client", 1, "client count")
	benchmarkCmd.PersistentFlags().IntVar(&requestCnt, "request", 100000, "request count")
	benchmarkCmd.PersistentFlags().StringVar(&workloadStr, "workload", "read-only", "workload type, from read-only, half-half, write-only")
}
