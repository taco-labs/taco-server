package value

import (
	"fmt"
	"strings"

	"github.com/taco-labs/taco/go/domain/value/enum"
)

var (
	digitToWord = strings.NewReplacer(
		"0", "A",
		"1", "B",
		"2", "C",
		"3", "D",
		"4", "E",
		"5", "F",
		"6", "G",
		"7", "H",
		"8", "I",
		"9", "J",
	)
	wordToDigit = strings.NewReplacer(
		"A", "0",
		"B", "1",
		"C", "2",
		"D", "3",
		"E", "4",
		"F", "5",
		"G", "6",
		"H", "7",
		"I", "8",
		"J", "9",
	)
)

type ReferralCode struct {
	ReferralType enum.ReferralType
	PhoneNumber  string
}

func (r ReferralCode) Encode() (string, error) {
	referralType, err := r.encodeReferralType()
	if err != nil {
		return "", err
	}
	phone := r.encodePhoneNumber()

	return fmt.Sprintf("%s-%s", referralType, phone), nil
}

func (r *ReferralCode) Decode(encodedStr string) error {
	parts := strings.Split(encodedStr, "-")
	if len(parts) != 2 {
		return fmt.Errorf("%w: invalid referral code: %s", ErrInvalidOperation, encodedStr)
	}
	referralType, err := r.decodeReferralType(parts[0])
	if err != nil {
		return err
	}
	phoneNumber := r.decodePhoneNmber(parts[1])

	r.ReferralType = referralType
	r.PhoneNumber = phoneNumber

	return nil
}

func (r ReferralCode) encodeReferralType() (string, error) {
	var encodedString string
	switch r.ReferralType {
	case enum.ReferralType_Driver:
		encodedString = "D"
	case enum.ReferralType_User:
		encodedString = "U"
	default:
		return "", fmt.Errorf("%w: invalid referral type: %v", ErrInvalidOperation, r.ReferralType)
	}

	return encodedString, nil
}

func (r ReferralCode) encodePhoneNumber() string {
	return digitToWord.Replace(strings.TrimPrefix(r.PhoneNumber, "010"))
}

func (r ReferralCode) decodeReferralType(encodedStr string) (enum.ReferralType, error) {
	var referralType enum.ReferralType
	switch encodedStr {
	case "D":
		referralType = enum.ReferralType_Driver
	case "U":
		referralType = enum.ReferralType_User
	default:
		return enum.ReferralType_Unknown, fmt.Errorf("%w: invalid encoded referral type: %v", ErrInvalidOperation, encodedStr)
	}

	return referralType, nil
}

func (r ReferralCode) decodePhoneNmber(encodedStr string) string {
	return fmt.Sprintf("010%s", wordToDigit.Replace(encodedStr))
}

func EncodeReferralCode(referralCode ReferralCode) (string, error) {
	return referralCode.Encode()
}

func DecodeReferralCode(encodedStr string) (ReferralCode, error) {
	r := ReferralCode{}

	err := r.Decode(encodedStr)
	return r, err
}
