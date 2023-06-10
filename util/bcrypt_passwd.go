package util

import "golang.org/x/crypto/bcrypt"

// Password encryption using adaptive hash algorithm, irreversible
func GenPasswd(passwd string) string {
	hashPasswd, _ := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	return string(hashPasswd)
}

// Determine if two strings of hash are from the same plaintext by comparing them
// hashPasswd The ciphertext to be compared
// passwd plaintext
func ComparePasswd(hashPasswd string, passwd string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashPasswd), []byte(passwd)); err != nil {
		return err
	}
	return nil
}
