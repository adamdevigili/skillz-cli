package util

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func EncryptPassword(rawPassword string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hashedPassword, nil
}

func PasswordsMatch(storedPassword []byte, rawPassword string) bool {
	return bcrypt.CompareHashAndPassword(storedPassword, []byte(rawPassword)) == nil
}

func IsValidPassword(rawPassword string, invalidCharsMap map[rune]bool, minLength, maxLength int) error {
	if lenErr := lengthCheck(rawPassword, minLength, maxLength); lenErr != nil {
		return lenErr
	}

	charCountMap, charErr := containsInvalidChars(rawPassword, invalidCharsMap)
	if charErr != nil {
		return charErr
	}

	if wsErr := hasEnoughWhitespace(charCountMap); wsErr != nil {
		return wsErr
	}

	if digitErr := hasRequiredDigits(rawPassword); digitErr != nil {
		return digitErr
	}

	return nil
}

func lengthCheck(rawPassword string, minLength, maxLength int) error {
	l := len(rawPassword)
	if l < minLength {
		return fmt.Errorf("Password must be %d characters or more. Provided %d", minLength, l)
	}
	if l > maxLength {
		return fmt.Errorf("Password must be %d characters or less. Provided %d", maxLength, l)
	}

	return nil
}

func containsInvalidChars(rawPassword string, invalidCharsMap map[rune]bool) (map[rune]int, error) {
	// Easy lookup table for future checks
	charCountMap := make(map[rune]int)

	// Populate the count map to return, also check for invalid characters
	for _, char := range rawPassword {
		if _, ok := charCountMap[char]; ok {
			charCountMap[char] += 1
		} else {
			charCountMap[char] = 1
		}

		if invalidCharsMap[char] {
			return nil, fmt.Errorf("password contains invalid characters: %c", char)
		}
	}

	return charCountMap, nil
}

func hasEnoughWhitespace(charCountMap map[rune]int) error {
	count := charCountMap['_']
	if count < 3 {
		return fmt.Errorf("Not enough whitespace characters provided. Provided %d, Required: %d", count, 3)
	}

	return nil
}

func hasRequiredDigits(rawPassword string) error {
	// We can do this faster when we iterate through the string the initial time, placeholder for now
	for i := 4; i <= 9; i++ {
		s := strconv.Itoa(i)
		if strings.Contains(rawPassword, s) {
			return nil
		}
	}
	return fmt.Errorf("Password must contain a digit between 4-9")
}
