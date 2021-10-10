package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/adamdevigili/skillz-cli/pkg/util"
	"github.com/spf13/cobra"
)

func init() {
	skillzRoot.AddCommand(userCmd)
	userCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(passwordCmd)
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage the currently logged in user. Look at statistics, change your password, etc.",
	RunE:  getUser,
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update account settings for the currently logged in user",
	RunE:  updateUser,
}

var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Update your password",
	RunE:  updatePassword,
}

func getUser(cmd *cobra.Command, args []string) error {
	usersDB, err := util.GetUsersDB(dbPath, usersBucket, currentUserBucket)
	if err != nil {
		fmt.Println("Error when connecting to database")
		os.Exit(1)
	}
	defer usersDB.Close()

	currentUser, err := util.GetCurrentUser(usersDB, currentUserBucket, currentUserKey)
	if err != nil {
		return fmt.Errorf("Error when trying to read from database: %w", err)
	}

	if currentUser == nil {
		return fmt.Errorf("No user currently logged in")
	}

	// Hide password from response, output JSON
	currentUser.HashedPassword = nil
	json.NewEncoder(os.Stdout).Encode(currentUser)

	return nil
}

func updateUser(cmd *cobra.Command, args []string) error {
	return nil
}

func updatePassword(cmd *cobra.Command, args []string) error {
	usersDB, err := util.GetUsersDB(dbPath, usersBucket, currentUserBucket)
	if err != nil {
		fmt.Println("Error when connecting to database")
		os.Exit(1)
	}
	defer usersDB.Close()

	currentUser, err := util.GetCurrentUser(usersDB, currentUserBucket, currentUserKey)
	if err != nil {
		return fmt.Errorf("Error when trying to read from database: %w", err)
	}

	if currentUser == nil {
		fmt.Println("No user currently logged in")
		return nil
	}

	rawPassword := promptForPassword()
	for i := passwordAttempts - 1; i >= 0; i-- {
		if i == 0 {
			return fmt.Errorf("No more attempts left, exiting")
		}

		if !util.PasswordsMatch(*currentUser.HashedPassword, rawPassword) {
			fmt.Println(fmt.Sprintf("Incorrect password provided. Attempts left: %d", i))
			rawPassword = promptForPassword()
		} else {
			break
		}
	}

	displayPasswordRestrictions()
	newPassword := promptForNewPassword(passwordAttempts)

	hashedPassword, err := util.EncryptPassword(newPassword)
	if err != nil {
		return fmt.Errorf("Error when trying to encrypt password: %w", err)
	}

	currentUser.HashedPassword = &hashedPassword
	if err := util.UpdateUser(usersDB, usersBucket, currentUser.Username, currentUser); err != nil {
		return fmt.Errorf("Error when trying to update database: %w", err)
	}

	fmt.Println(fmt.Sprintf("Password updated for %s", currentUser.Username))

	return nil
}
