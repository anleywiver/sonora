// Package passwordhash wraps bcrypt for credential-based login (Sprint
// 14 sisipan, ADR 0012). Cost 12 — used consistently by the seed-owner
// CLI, the "Add User" admin form, and login verification, so a hash
// created by one path always verifies via any other.
package passwordhash

import "golang.org/x/crypto/bcrypt"

const Cost = 12

func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), Cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Compare(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
