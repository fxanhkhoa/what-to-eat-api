# Apple ID Token Verification

This document describes the implementation of Apple ID token verification in the what-to-eat API.

## Overview

The API now supports both Google and Apple ID token authentication. The `LoginDto` has been extended with a `Type` field that can be either "google" or "apple", with "google" as the default value.

## Changes Made

### 1. Model Updates

**`model/auth.go`**:
- Added `Type` field to `LoginDto` with default value "google"
- Added `AppleClaims` struct for Apple ID token claims
- Added `AppleUserInfo` struct for processed Apple user information

**`model/user.go`**:
- Added `AppleID` field to `User` struct
- Added `AppleID` field to `CreateUserDto` struct

**`model/jwt.go`**:
- Added `AppleID` field to `JwtCustomClaims` struct

### 2. Service Updates

**`service/auth.go`**:
- Updated `Login()` method to handle both Google and Apple login types
- Added `verifyAppleIdToken()` method for Apple ID token verification
- Added `getAppleJWKS()` method to fetch Apple's public keys
- Added `createRSAPublicKey()` method to create RSA public key from JWK
- Updated token generation methods to include Apple ID in JWT claims

**`service/user.go`**:
- Added `FindUserByAppleID()` method to find users by Apple ID
- Added `CreateUserWithAppleFromOAuth()` method to create users from Apple OAuth

### 3. Apple ID Token Verification Process

The Apple ID token verification follows these steps:

1. **Parse JWT Header**: Extract the `kid` (key ID) from the token header
2. **Fetch Apple JWKS**: Get Apple's JSON Web Key Set from `https://appleid.apple.com/auth/keys`
3. **Find Matching Key**: Locate the public key with matching `kid`
4. **Create RSA Public Key**: Convert JWK to RSA public key using modulus (n) and exponent (e)
5. **Verify Token**: Use RSA public key to verify the JWT signature
6. **Validate Claims**: Ensure issuer is "https://appleid.apple.com"
7. **Extract User Info**: Create `AppleUserInfo` from validated claims

## Usage Examples

### Login with Apple ID Token

```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "type": "apple"
}
```

### Login with Google ID Token (Default)

```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "type": "google"
}
```

or simply:

```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Apple ID Token Structure

Apple ID tokens contain the following relevant claims:

```json
{
  "iss": "https://appleid.apple.com",
  "aud": "your.bundle.id",
  "exp": 1234567890,
  "iat": 1234567890,
  "sub": "001234.567890abcdef.1234",
  "email": "user@example.com",
  "email_verified": true,
  "name": {
    "firstName": "John",
    "lastName": "Doe"
  }
}
```

**Note**: Apple only provides name information in the first authentication. Subsequent logins may not include the name fields.

## Security Considerations

1. **Key Rotation**: Apple rotates their signing keys regularly. The implementation fetches keys dynamically from Apple's JWKS endpoint.

2. **Audience Validation**: In production, you should validate the `aud` (audience) claim matches your app's bundle identifier.

3. **Issuer Validation**: The implementation verifies that the issuer is "https://appleid.apple.com".

4. **Token Expiration**: Standard JWT expiration validation is handled by the jwt library.

## Dependencies

The Apple ID verification uses the following Go packages:
- `github.com/golang-jwt/jwt/v5` for JWT parsing and validation
- `crypto/rsa` for RSA public key operations
- `math/big` for big integer operations with RSA keys
- Standard library packages for HTTP requests and JSON processing

## Error Handling

The implementation handles various error scenarios:
- Invalid JWT format
- Missing or invalid key ID
- Key not found in Apple's JWKS
- RSA key creation errors
- Token signature verification failures
- Invalid issuer claims

All errors are logged and returned with descriptive messages for debugging.
