package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_isPrime(t *testing.T) {
	testCases := []struct {
		name     string
		testNum  int
		expected bool
		msg      string
	}{
		{
			"Valid prime number",
			7,
			true,
			"7 is a prime number!",
		}, {
			"Valid prime number",
			8,
			false,
			"8 is not prime,are divisible by: 2",
		},
		{
			"Invalid prime number by definition",
			0,
			false,
			"0 is not prime, by definition!",
		},
		{
			"Negative number",
			-1,
			false,
			"negative numbers are not prime, by definition!",
		},
	}

	for _, tc := range testCases {
		result, msg := isPrime(tc.testNum)

		if tc.expected != result {
			t.Errorf("%s: expected %t, got %t", tc.name, tc.expected, result)
		}
		if tc.msg != msg {
			t.Errorf("%s: expected %s, got %s", tc.name, tc.msg, msg)
		}

	}
}

func Test_prompt(t *testing.T) {
	oldOut := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("error when create pipe:%v", err)
	}
	os.Stdout = w

	prompt()

	err = w.Close()
	if err != nil {
		t.Errorf("error when close write pipe:%v", err)
	}
	os.Stdout = oldOut

	out, err := io.ReadAll(r)
	if err != nil {
		t.Errorf("error when read buffer:%v", err)
	}
	if string(out) != "-> " {
		t.Errorf("expected: -> , but got %s", string(out))
	}
}

func Test_intro(t *testing.T) {
	oldOut := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("error when create pipe:%v", err)
	}
	os.Stdout = w

	intro()

	err = w.Close()
	if err != nil {
		t.Errorf("error when close write pipe:%v", err)
	}
	os.Stdout = oldOut

	out, err := io.ReadAll(r)
	if err != nil {
		t.Errorf("error when read buffer:%v", err)
	}
	if !strings.Contains(string(out), "Enter a whole number") {
		t.Errorf("intro text not correct: %s", string(out))
	}
}

func Test_checkNumbers(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Empty",
			"",
			"Please enter a whole number!",
		},
		{
			"ZERO",
			"0",
			"0 is not prime, by definition!",
		},
		{
			"ONE",
			"1",
			"1 is not prime, by definition!",
		},
		{
			"Valid prime",
			"2",
			"2 is a prime number!",
		},
		{
			"Negative number",
			"-1",
			"Negative numbers are not prime, by definition!",
		},
		{
			"Typed",
			"three",
			"Please enter a whole number!",
		},
		{
			"Quit",
			"q",
			"",
		},
		{
			"Quit capitalize",
			"Q",
			"",
		},
		{
			"Japanese alphabet",
			"すき",
			"Please enter a whole number!",
		},
	}

	for _, tc := range testCases {
		input := strings.NewReader(tc.input)
		reader := bufio.NewScanner(input)

		result, _ := checkNumbers(reader)

		if !strings.EqualFold(result, tc.expected) {
			t.Errorf("%s: expected %s, but got %s", tc.name, tc.expected, result)
		}
	}
}

func Test_ReadUserInput(t *testing.T) {
	doneChan := make(chan bool)

	var stdin bytes.Buffer

	stdin.Write([]byte("1\nq\n"))

	go readUserInput(&stdin, doneChan)
	<-doneChan
	close(doneChan)
}
