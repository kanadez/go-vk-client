package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const vk_auth_url string = "https://api.vk.com/method"
const vk_auth_token string = "vk1.a.fkgV1-PwSZi8nXnZRdjAr65TGmcuxFjMykgu8nUbUKlqk_ZjZMce1p8Q_yR49Uaog3g80Ey8lKmpMfyqM-J4ROC-DJcgCcI9u2F9l8CPhdkFMokupcCGu2Y62g0ofpGspRoVIKnrtS2yQpb4v_ZSQGOvoq4WyBzaCZFnbANKe8o6vrkf8BVlWBT59P7CdxCgxGOXBWbY5nfFcDVjBb4mWQ"

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		printHelp()
		os.Exit(0)
	}

	vk_id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   20 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	processVkRequest(client, vk_id)
}

func printHelp() {
	fmt.Println("Usage: program <vk_user_id>")
}

func processVkRequest(client *http.Client, vk_id int64) (string, error) {
	method := "users.get"

	type VkQueryOptions struct {
		UserIDs     int64  `url:"user_ids,omitempty"`
		Fields      string `url:"fields,omitempty"`
		AccessToken string `url:"access_token,omitempty"`
		Version     string `url:"v,omitempty"`
	}

	var vk_payload = VkQueryOptions{
		UserIDs:     vk_id,
		Fields:      "bdate",
		AccessToken: vk_auth_token,
		Version:     "5.131",
	}

	vk_payload_query_string, err := query.Values(vk_payload)
	fmt.Printf("Sending request to %s, method %s, with params:\n%+v\n", vk_auth_url, method, vk_payload)

	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s?%s", vk_auth_url, method, vk_payload_query_string.Encode()), nil)

	if err != nil {
		fmt.Println(err.Error())
		return "", errors.Wrap(err, "account NewRequest")
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+vk_auth_token)
	req.Header.Add("Host", "edge.vk.com")
	req.Header.Add("Content-Length", strconv.Itoa(len(vk_payload_query_string)))

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
		return "", errors.Wrap(err, "account Do")
	}

	fmt.Println(fmt.Sprintf("vk server response status: %s", resp.Status))

	defer resp.Body.Close() // defer очищает ресурс

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)

		type BodyJson struct {
			ID              int64  `json:"id"`
			Bdate           string `json:"bdate,omitempty"`
			FirstName       string `json:"first_name,omitempty"`
			LastName        string `json:"last_name,omitempty"`
			CanAccessClosed bool   `json:"can_access_closed"`
			IsClosed        bool   `json:"is_closed"`
		}

		type Response struct {
			Response []BodyJson `json:"response"`
		}

		type ErrorDetails struct {
			ErrorCode    int    `json:"error_code"`
			ErrorMessage string `json:"error_msg"`
		}

		type Error struct {
			Error ErrorDetails `json:"error"`
		}

		var response Response
		err = json.Unmarshal([]byte(bodyString), &response)
		if err != nil {
			log.Fatal(err)
		}

		var error Error
		err = json.Unmarshal([]byte(bodyString), &error)
		if err != nil {
			log.Fatal(err)
		}

		if error.Error.ErrorCode != 0 {
			fmt.Printf("Body : %+v\n", error)
		} else {
			fmt.Printf("Body : %+v\n", response)
		}

	}

	return "", nil
}
