package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	intro()

	doneChan := make(chan bool)

	go readUserInput(os.Stdin, doneChan)

	<-doneChan
	close(doneChan)

	fmt.Println("GoodBye.")
}

func readUserInput(in io.Reader, doneChan chan bool) {
	scanner := bufio.NewScanner(in)

	for {
		result, done := checkNumbers(scanner)
		if done {
			doneChan <- true
			return
		}

		fmt.Println(result)
		prompt()
	}
}

func checkNumbers(scanner *bufio.Scanner) (string, bool) {
	scanner.Scan()

	if strings.EqualFold(scanner.Text(), "q") {
		return "", true
	}

	numToCheck, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return "Please enter a whole number!", false
	}

	_, msg := isPrime(numToCheck)

	return msg, false
}

func intro() {
	fmt.Println("Welcome to the prime number checker!")
	fmt.Println("Enter a whole number, and we'll tell you if it's prime or not. Enter q to quit.")
	prompt()
}

func prompt() {
	fmt.Print("-> ")

}

func isPrime(value int) (bool, string) {
	if value == 0 || value == 1 {
		return false, fmt.Sprintf("%d is not prime, by definition!", value)
	}

	if value < 0 {
		return false, "negative numbers are not prime, by definition!"
	}

	for i := 2; i <= int(math.Sqrt(float64(value))); i++ {
		if value%i == 0 {
			return false, fmt.Sprintf("%d is not prime,are divisible by: %d", value, i)
		}
	}
	return true, fmt.Sprintf("%d is a prime number!", value)
}
