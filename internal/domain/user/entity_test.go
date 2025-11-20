package user_test

import (
	"strings"
	"testing"

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