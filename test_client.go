package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CreateUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	fmt.Println("Testing gRPC-style HTTP API...")

	req := CreateUserRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(req)

	resp, err := http.Post("http://localhost:8080/users", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		var user UserResponse
		json.Unmarshal(body, &user)
		fmt.Printf("✅ User created successfully: %+v\n", user)

		getUserResp, err := http.Get(fmt.Sprintf("http://localhost:8080/users?id=%s", user.ID))
		if err != nil {
			fmt.Printf("Error getting user: %v\n", err)
			return
		}
		defer getUserResp.Body.Close()

		if getUserResp.StatusCode == 200 {
			getUserBody, _ := io.ReadAll(getUserResp.Body)
			var retrievedUser UserResponse
			json.Unmarshal(getUserBody, &retrievedUser)
			fmt.Printf("✅ User retrieved successfully: %+v\n", retrievedUser)
		} else {
			fmt.Printf("❌ Failed to retrieve user, status: %d\n", getUserResp.StatusCode)
		}
	} else {
		fmt.Printf("❌ Failed to create user, status: %d, body: %s\n", resp.StatusCode, string(body))
	}
}