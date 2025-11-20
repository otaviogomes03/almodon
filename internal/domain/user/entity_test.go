package user_test

import (
	"strings"
	"testing"

	"github.com/alan-b-lima/almodon/internal/auth"
	. "github.com/alan-b-lima/almodon/internal/domain/user"
)

func TestProcessEmail(t *testing.T) {
	type Tests struct {
		input      string
		shouldFail bool
	}

	tests := []Tests{
		{"usuario@almodon.com", false},
		{"nome.sobrenome@dominio.br", false},
		{"", true},                   
		{"usuarioalmodon.com", true}, 
		{"usuario@", true},           
		{"@almodon.com", true},      
		{"usuario@.com", true},       
	}

	for _, test := range tests {
		_, err := ProcessEmail(test.input)

		if (err != nil) != test.shouldFail {
			if test.shouldFail {
				t.Errorf("Email '%v': expected error, but got nil", test.input)
			} else {
				t.Errorf("Email '%v': did not expect error, but got: %v", test.input, err)
			}
		}
	}
}

func TestProcessPassword(t *testing.T) {
	longPass := strings.Repeat("a", 65)

	type Tests struct {
		input      string
		shouldFail bool
	}

	tests := []Tests{
		{"SenhaForte123!", false},
		{"12345678", false}, 
		{"1234567", true},   
		{longPass, true},    
		{" 12345678", true}, 
		{"12345678 ", true}, 
		{"", true},          
		{"Senha" + string(rune(0)) + "123", true}, 
	}

	for _, test := range tests {
		_, err := ProcessPassword(test.input)

		if (err != nil) != test.shouldFail {
			if test.shouldFail {
				t.Errorf("Password '%v': expected error, but got nil", test.input)
			} else {
				t.Errorf("Password '%v': did not expect error, but got: %v", test.input, err)
			}
		}
	}
}

func TestNew(t *testing.T) {
	type Tests struct {
		siape      int
		name       string
		email      string
		password   string
		role       auth.Role
		shouldFail bool
	}

	tests := []Tests{
		{123456, "Alan Lima", "alan@almodon.com", "StrongPass1!", auth.Admin, false},
		{123456, "Alan Lima", "invalid-email", "StrongPass1!", auth.User, true},
		{123456, "Alan Lima", "alan@almodon.com", "short", auth.User, true},
		{123456, "", "alan@almodon.com", "StrongPass1!", auth.User, true},
		{123456, "Alan Lima", "alan@almodon.com", "StrongPass1!", auth.Role(50), true},
	}

	for _, test := range tests {
		u, err := New(test.siape, test.name, test.email, test.password, test.role)

		if (err != nil) != test.shouldFail {
			if test.shouldFail {
				t.Errorf("New User (Integration): expected error, but got nil. Input: %+v", test)
			} else {
				t.Errorf("New User (Integration): did not expect error, but got: %v. Input: %+v", err, test)
			}
		}

		if !test.shouldFail && err == nil {
			if u.Name() != test.name {
				t.Errorf("New User: expected name '%s', got '%s'", test.name, u.Name())
			}
			if u.Role() != test.role {
				t.Errorf("New User: expected role '%v', got '%v'", test.role, u.Role())
			}
		}
	}
}