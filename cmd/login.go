package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/adamdevigili/skillz-cli/pkg/models"
	"github.com/adamdevigili/skillz-cli/pkg/util"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// Main invalid char slice for populating map on startup as well as for output
var invalidCharsSlice = []rune{'$', '^', '&', '*', '(', ')', '[', ']'}

// "set" of invalid characters for O(1) lookup (probably overkill)
var invalidCharsMap = map[rune]bool{}

// Configurable password restrictions
var passwordMinLength, passwordMaxLength = 10, 32
var passwordAttempts = 5

func init() {
	skillzRoot.AddCommand(loginCmd)
	skillzRoot.AddCommand(logoutCmd)

	// Init our invalid char map
	for _, c := range invalidCharsSlice {
		invalidCharsMap[c] = true
	}
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the Skillz platform using your email and password",
	Long: "Login to the Skillz platform using your email and password. " +
		"If the provided username is not registered, you will be able to create an account",
	RunE: loginUser,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of the Skillz platform",
	RunE:  logoutUser,
}

func loginUser(cmd *cobra.Command, args []string) error {
	newUser := false
	username := promptForUsername()

	usersDB, err := util.GetUsersDB(dbPath, usersBucket, currentUserBucket)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error when connecting to database: %w", err))
		os.Exit(1)
	}
	defer usersDB.Close()

	var rawPassword string
	var user *models.User

	// get the User model from the DB
	user, err = util.GetUser(usersDB, usersBucket, username)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error when connecting to database: %w", err))
		os.Exit(1)
	}

	// User not in DB, prompt for new user
	if user == nil {
		fmt.Println("Username not found")
		newUser = true
		if promptToCreateUser() {
			if !promptToReuseUsername(username) {
				username = promptForUsername()
			}
			displayPasswordRestrictions()
		} else {
			os.Exit(0)
		}
	}

	// Get correct password for context
	if newUser {
		rawPassword = promptForNewPassword(passwordAttempts)
	} else {
		rawPassword = promptForPassword()
	}

	// Add hashed password to table if new user, otherwise validate it matches what was present
	if newUser {
		hashedPassword, err := util.EncryptPassword(rawPassword)
		if err != nil {
			return fmt.Errorf("Error when trying to encrypt password: %w", err)
		}

		user, err = util.AddNewUser(usersDB, usersBucket, username, hashedPassword)
		if err != nil {
			return fmt.Errorf("Error when trying to update database: %w", err)
		}

		fmt.Println(fmt.Sprintf("User %s created", username))
	} else {
		for i := passwordAttempts - 1; i >= 0; i-- {
			if i == 0 {
				return fmt.Errorf("No more attempts left, exiting")
			}

			if !util.PasswordsMatch(*user.HashedPassword, rawPassword) {
				fmt.Println(fmt.Sprintf("Incorrect password provided. Attempts left: %d", i))
				rawPassword = promptForPassword()
			} else {
				break
			}
		}
	}

	if err := util.SetCurrentUser(usersDB, currentUserBucket, currentUserKey, user); err != nil {
		return fmt.Errorf("Error when trying to update database: %w", err)
	}

	fmt.Println("successfully logged in")
	return nil
}

func logoutUser(cmd *cobra.Command, args []string) error {
	usersDB, err := util.GetUsersDB(dbPath, usersBucket, currentUserBucket)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error when connecting to database: %w", err))
		os.Exit(1)
	}
	defer usersDB.Close()

	currentUser, err := util.GetCurrentUser(usersDB, currentUserBucket, currentUserKey)
	if err != nil {
		return fmt.Errorf("Error when trying to read from database: %w", err)
	}

	if currentUser == nil {
		fmt.Println("No user currently logged in")
	}

	if err := util.UnsetCurrentUser(usersDB, currentUserBucket, currentUserKey); err != nil {
		return fmt.Errorf("Error when trying to update database: %w", err)
	}

	fmt.Println(fmt.Sprintf("%s successfully logged out", currentUser.Username))
	return nil
}

func promptToCreateUser() bool {
	prompt := promptui.Prompt{
		Label:     "Create a new account",
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	if result == "y" || result == "Y" {
		return true
	}

	return false
}

func promptToReuseUsername(username string) bool {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Use provided username, (%s)", username),
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	if result == "y" || result == "Y" {
		return true
	}

	return false
}

func promptForUsername() string {
	validate := func(input string) error {
		if len(input) <= 0 {
			return errors.New("Username cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Username",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func promptForPassword() string {
	prompt := promptui.Prompt{
		Label: "Current password",
		Mask:  ' ',
	}

	rawPassword, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return rawPassword
}

// promptForNewPassword is called when creating a new user or when an existing user wants to change their password
func promptForNewPassword(attempts int) string {
	if attempts == 0 {
		fmt.Println("No more attempts left, exiting")
		os.Exit(0)
	}

	prompt := promptui.Prompt{
		Label: "Password",
		Mask:  ' ',
	}

	rawPassword, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	if err := util.IsValidPassword(rawPassword, invalidCharsMap, passwordMinLength, passwordMaxLength); err != nil {
		fmt.Println(fmt.Sprintf("Invalid password: %s. Attempts left: %d", err.Error(), attempts-1))
		rawPassword = promptForNewPassword(attempts - 1)
	}

	passwordConfirm := promptForConfirmPassword()
	if passwordConfirm != rawPassword {
		fmt.Println("Passwords do not match")
		os.Exit(0)
	}

	return rawPassword
}

// promptForConfirmPassword is called when creating a new user or when an existing user wants to change their password
func promptForConfirmPassword() string {
	prompt := promptui.Prompt{
		Label: "Confirm password",
		Mask:  ' ',
	}

	rawPassword, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return rawPassword
}

func displayPasswordRestrictions() {
	restrictionString := "Password restrictions:\n" +
		"- 10 character minimum, 32 character maximum\n" +
		"- 3 whitespace character min\n" +
		"- 1 digit between 4-9\n" +
		"- can not contain the following symbols: "

	var sb strings.Builder
	sb.WriteString(restrictionString)
	n := len(invalidCharsSlice)
	for i, c := range invalidCharsSlice {
		if i != n-1 {
			sb.WriteString(fmt.Sprintf("'%c', ", c))
		} else {
			sb.WriteString(fmt.Sprintf("'%c'", c))
		}

	}

	fmt.Print(sb.String())
}
