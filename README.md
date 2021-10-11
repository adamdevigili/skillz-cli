# skillz-cli

## Features
Small Go CLI that has the following features
- Allows users to create accounts and set passwords
- Ensures passwords meet certain restrictions before creating account
- Logs users in, allowing them to perform more advanced commands for their account
- Allows users to change their passwords
- Uses [BoltDB](https://github.com/boltdb/bolt), [cobra](https://github.com/spf13/cobra)/[promptui](https://github.com/manifoldco/promptui), and [crypto](https://pkg.go.dev/golang.org/x/crypto)

## Usage
### Install
Ensure you have [Go installed](https://golang.org/doc/install)

Clone the repository..

`git clone https://github.com/adamdevigili/skillz-cli.git`

Go into the root directory and install the CLI

`cd skillz-cli && go install`

Ensure the CLI is installed by running `skillz-cli --help`

```
The Skillz CLI tool. Login to your account and manage your stats

Usage:
  skillz [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  login       Login to the Skillz platform using your email and password
  logout      Logout of the Skillz platform
  user        Manage the currently logged in user. Look at statistics, change your password, etc.

Flags:
  -h, --help   help for skillz

Use "skillz
```

### Commands
Currently, there are a simple set of commands. Here's a use-case path to exercise them all. 

### Windows users
NOTE: By default the CLI will try to make a local `.db` file in the `/tmp` directory (MacOS/Linux). Windows support would be a follow up. 

---
`skillz-cli login`

```
$ skillz-cli login

Username: adamdevigili
Username not found
Create a new account: y
Use provided username, (adamdevigili): y
Password restrictions:
- 10 character minimum, 32 character maximum
- 3 whitespace character min
- 1 digit between 4-9
Password:                
User adamdevigili created
successfully logged in
```

---
`skillz-cli user`

```
$ skillz-cli user
{"username":"adamdevigili","created":"2021-10-10T15:26:09.019831-05:00"}
```

---
`skillz-cli user update password`
``` 
$ skillz-cli user update password
Current password:                
Password restrictions:
- 10 character minimum, 32 character maximum
- 3 whitespace character min
- 1 digit between 4-9
Password:                
Confirm password:                
Password updated for adamdevigili
```

---
`skillz-cli logout`

```
$ skillz-cli logout
adamdevigili successfully logged out

$ skillz-cli logout
No user currently logged in
```
