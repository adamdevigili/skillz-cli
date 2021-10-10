package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Using BoltDB and local file system. This would be a path to a real DB
const dbPath = "/tmp/users.db"
const usersBucket = "skillz.users.bucket"
const currentUserBucket = "skillz.current.user.bucket"
const currentUserKey = "current.user"

var skillzRoot = &cobra.Command{
	Use:   "skillz",
	Short: "The Skillz CLI tool. Login to your account and manage your stats",
}

func Execute() {
	if err := skillzRoot.Execute(); err != nil {
		os.Exit(1)
	}
}
