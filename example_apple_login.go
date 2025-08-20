// Example Apple ID Login Implementation
// This file demonstrates how to use the Apple ID token verification

package main

import (
	"encoding/json"
	"fmt"
	"what-to-eat/be/model"
	"what-to-eat/be/service"
)

func exampleAppleLogin() {
	// Example usage of Apple ID token verification
	authService := service.AuthService{}

	// Test LoginDto with Apple type
	loginDto := model.LoginDto{
		Token: "your-apple-id-token-here", // Replace with actual Apple ID token
		Type:  "apple",
	}

	fmt.Printf("Login request: %+v\n", loginDto)

	// Example of how the login would work
	// result, err := authService.Login(loginDto)
	// if err != nil {
	//     log.Printf("Login failed: %v", err)
	//     return
	// }

	// resultJSON, _ := json.MarshalIndent(result, "", "  ")
	// fmt.Printf("Login successful: %s\n", resultJSON)

	// Test with Google (default behavior)
	googleLoginDto := model.LoginDto{
		Token: "your-google-id-token-here", // Replace with actual Google ID token
		Type:  "",                          // Will default to "google"
	}

	if googleLoginDto.Type == "" {
		googleLoginDto.Type = "google" // Default value
	}

	fmt.Printf("Google login request: %+v\n", googleLoginDto)

	// Example Apple user info structure
	exampleAppleUser := model.AppleUserInfo{
		Sub:           "001234.567890abcdef.1234",
		Email:         "user@example.com",
		VerifiedEmail: true,
		Name:          "John Doe",
		GivenName:     "John",
		FamilyName:    "Doe",
	}

	appleUserJSON, _ := json.MarshalIndent(exampleAppleUser, "", "  ")
	fmt.Printf("Example Apple user info: %s\n", appleUserJSON)

	fmt.Println("Apple ID token verification setup complete!")

	_ = authService // Use the variable to avoid compiler warning
}
