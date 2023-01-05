package value

import (
	"fmt"
	"strings"
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
	PhoneNumber string
}

func (r ReferralCode) Encode() string {
	return r.encodePhoneNumber()
}

func (r *ReferralCode) Decode(encodedStr string) {
	r.PhoneNumber = r.decodePhoneNmber(encodedStr)
}

func (r ReferralCode) encodePhoneNumber() string {
	return digitToWord.Replace(strings.TrimPrefix(r.PhoneNumber, "010"))
}

func (r ReferralCode) decodePhoneNmber(encodedStr string) string {
	return fmt.Sprintf("010%s", wordToDigit.Replace(strings.ToUpper(encodedStr)))
}

func EncodeReferralCode(referralCode ReferralCode) string {
	return referralCode.Encode()
}

func DecodeReferralCode(encodedStr string) ReferralCode {
	r := ReferralCode{}

	r.Decode(encodedStr)
	return r
}
