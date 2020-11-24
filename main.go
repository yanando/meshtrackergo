package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

type OrderStatusResponse struct {
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

type OrderInfo struct {
	placed     int
	processing int
	shipped    int
	delivered  int
	cancelled  int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fascia := getFascia(scanner)
	locale := getLocale(scanner)
	zip := getZip(scanner)
	ordernumbers := getOrderNumbers(scanner)

	var wg sync.WaitGroup
	wg.Add(len(ordernumbers))
	orderInfo := OrderInfo{
		placed:     0,
		processing: 0,
		shipped:    0,
		delivered:  0,
		cancelled:  0,
	}

	for _, ordernumber := range ordernumbers {
		go func(ordernumber string) {
			status, err := getOrderStatus(fascia, locale, zip, ordernumber)

			if err != nil {
				fmt.Println(ordernumber, "error")
				wg.Done()
				return
			}

			fmt.Println(ordernumber, status)

			switch status {
			case "Your order has been placed.":
				orderInfo.placed++
			case "Your order is currently being processed.":
				orderInfo.processing++
			case "Your order has been despatched.":
				orderInfo.shipped++
			case "Your order has been delivered.":
				orderInfo.delivered++
			case "It looks like your order has been cancelled.":
				orderInfo.cancelled++
			}

			wg.Done()
		}(ordernumber)
	}
	wg.Wait()

	fmt.Printf(`
placed     %d
processing %d
shipped    %d
delivered  %d
cancelled  %d`, orderInfo.placed, orderInfo.processing, orderInfo.shipped, orderInfo.delivered, orderInfo.cancelled)

	fmt.Scanln()
}

func getFascia(scanner *bufio.Scanner) string {
	fmt.Println(`please put in your fascia:
[1] Footpatrol
[2] Size
[3] JDsports`)

	scanner.Scan()
	fasciaNum := scanner.Text()
	var fascia string

	switch fasciaNum {
	case "1":
		fascia = "footpatrol"
	case "2":
		fascia = "size"
	case "3":
		fascia = "jdsports"
	default:
		fmt.Println("invalid fascia chosen")
		return getFascia(scanner)
	}

	return fascia
}

func getLocale(scanner *bufio.Scanner) string {
	fmt.Println(`which locale:
[1] UK
[2] NL
[3] DE
[4] DK
[5] BE
[6] IT
[7] ES
[8] FR`)

	scanner.Scan()
	localeNum := scanner.Text()
	var locale string

	switch localeNum {
	case "1":
		locale = "uk"
	case "2":
		locale = "nl"
	case "3":
		locale = "de"
	case "4":
		locale = "dk"
	case "5":
		locale = "be"
	case "6":
		locale = "it"
	case "7":
		locale = "es"
	case "8":
		locale = "fr"
	default:
		fmt.Println("invalid locale chosen")
		return getLocale(scanner)
	}

	return locale
}

func getZip(scanner *bufio.Scanner) string {
	fmt.Println("please put in your zip")

	scanner.Scan()
	return scanner.Text()
}

func getOrderNumbers(scanner *bufio.Scanner) []string {
	fmt.Println(`how do you want to input your ordernumbers?
[1] ordernumbers.txt (seperatated by a newline)
[2] manual input`)

	scanner.Scan()
	text := scanner.Text()

	if text == "1" {
		data, err := ioutil.ReadFile("ordernumbers.txt")

		if err != nil {
			fmt.Println("error reading ordernumbers.txt, make sure the file exists and is in the same directory as the executable")
			return getOrderNumbers(scanner)
		}

		rawOrdernumbers := string(data)

		return strings.Split(rawOrdernumbers, "\r\n")

	} else {
		fmt.Println("please put in your order number (seperated by a space)")

		scanner.Scan()
		orderNumbers := scanner.Text()
		return strings.Split(orderNumbers, " ")
	}
}

func getOrderStatus(fascia string, locale string, zip string, ordernumber string) (status string, err error) {
	var editedFascia string
	if locale == "gb" {
		if fascia == "footpatrol" {
			editedFascia = fascia + "gb"
		} else if fascia == "size" {
			editedFascia = fascia
		} else {
			editedFascia = fascia + locale
		}
	} else {
		editedFascia = fascia + locale
	}

	url := fmt.Sprintf("https://data.smartagent.io/v1/jdsports/track-my-order?orderNumber=%s&fascia=%s&postcode=%s", ordernumber, editedFascia, zip)

	resp, err := http.DefaultClient.Get(url)

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	resp.Body.Close()

	var responseStruct OrderStatusResponse

	err = json.Unmarshal(body, &responseStruct)

	if err != nil {
		return "", err
	}

	return responseStruct.Message.Text, nil
}
