package smtpdetails

//PasswordGenerator Mockable interface to generate passwords
//go:generate moq -out password_moq.go . PasswordGenerator
type PasswordGenerator interface {
	Generate(length, numDigits, numSymbols int, noUpper, allowRepeat bool) (string, error)
}
