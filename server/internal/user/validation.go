package user

import (
	"errors"
	"fmt"
	"regexp"
)

var alphabeticRegexp = regexp.MustCompile("^[\\p{L}\\_\\-\\. ]+$")
var groupCodeRegexp = regexp.MustCompile("^\\p{L}{2,3}\\-[0-9]{1,2}\\-[0-9]{1,2}$")
var phoneNumberRegexp = regexp.MustCompile("^\\+[1-9]\\d{1,14}$")

func validateUserName(name string, t string) error {
	if len(name) > 100 {
		return fmt.Errorf("user %s too long", t)
	}

	if !alphabeticRegexp.MatchString(name) {
		return fmt.Errorf("user %s contains invalid characters", t)
	}

	return nil
}

func validateGroupCode(groupCode string) error {
	if len(groupCode) > 100 {
		return errors.New("group code too long")
	}

	if !groupCodeRegexp.MatchString(groupCode) {
		return errors.New("group code is not valid")
	}

	return nil
}

func validatePhoneNumber(phoneNumber string) error {
	if !phoneNumberRegexp.MatchString(phoneNumber) {
		return errors.New("invalid phone number")
	}
	return nil
}
