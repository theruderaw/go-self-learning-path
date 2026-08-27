# 15. Authentication & Authorization (OAuth2)

This section covers OAuth2 as a protocol and the general concepts of authentication (proving who a user is) and authorization (determining what they're allowed to do) — not tied to any specific third-party provider. The same mechanics apply whether a backend acts as its own authorization server or delegates to an external one.

## Authentication vs authorization

These two terms are often used together but mean different things:

- **Authentication** answers "who is this?" — verifying an identity, typically via a password, a token, or a cryptographic signature.
- **Authorization** answers "what are they allowed to do?" — checking permissions once identity is established.

A login endpoint performs authentication. A middleware check like "does this user own this resource, or have the admin role?" performs authorization. OAuth2 is fundamentally an **authorization** protocol — it governs delegated access to resources, not identity verification. (OpenID Connect, a layer built on top of OAuth2, is what adds standardized authentication back in — covered near the end of this section.)

## The four roles in OAuth2

The OAuth2 specification defines four parties, and understanding what each one is responsible for makes every flow easier to follow:

- **Resource Owner** — the user, who owns the data being accessed.
- **Client** — the application requesting access (e.g. the backend being built).
- **Authorization Server** — issues tokens after authenticating the resource owner and obtaining their consent.
- **Resource Server** — hosts the protected data and accepts tokens as proof of authorized access.

In a system with a third-party login provider, the provider runs both the authorization server and resource server. In a system that manages its own users, the backend plays *all four roles at once* — it authenticates its own users, issues its own tokens, and protects its own resources with them. The mechanics are identical either way.

## Grant types (flows)

A "grant type" is the specific sequence of steps used to obtain a token. The choice depends on the client's nature (a server can keep a secret; a browser or mobile app generally cannot).

**Authorization Code** — the standard flow for a server-side application:

1. The resource owner is redirected to the authorization server to authenticate and consent.
2. The authorization server redirects back with a short-lived authorization code.
3. The client exchanges that code for an access token via a direct, server-to-server call — one that includes a client secret, since this step never happens in the browser.

**Authorization Code with PKCE** — the same flow, extended with a cryptographic proof (a "code verifier" and "code challenge") so that public clients without a stored secret — mobile apps, single-page apps — can use it safely. PKCE prevents a stolen authorization code from being redeemed by anyone other than the client that originally requested it.

**Client Credentials** — used for machine-to-machine access with no resource owner involved at all (e.g. one backend service calling another). The client authenticates directly with its own credentials and receives a token representing itself, not a user.

**Refresh Token** — not an initial grant, but a way to obtain a new access token once the current one expires, without repeating the full flow.

**Implicit** — an older flow that returned tokens directly in a redirect URL fragment, skipping the code-exchange step. It is now considered insecure (tokens are exposed in browser history and referrer headers) and has been formally deprecated in favor of Authorization Code with PKCE.

## Tokens

**Access token** — a short-lived credential (commonly valid for minutes to about an hour) presented with each request to prove authorization. Its short lifetime limits the damage if it's ever leaked.

**Refresh token** — a long-lived credential, exchanged for a new access token when the old one expires, without requiring the resource owner to authenticate again. It is issued once and stored securely; a compromised refresh token is a much more serious problem than a compromised access token, since it can be used to mint new access tokens indefinitely until revoked.

**Bearer token** — describes *how* a token is used, not what it contains: whoever holds ("bears") it can use it, with no additional proof of identity required. This is why access tokens must always be transmitted over HTTPS and never logged or embedded in a URL.

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Opaque vs structured tokens** — a token can be an opaque random string, meaningless outside the authorization server (which must be queried to check validity), or a structured, self-contained token such as a JWT, which can be verified locally without a network call.

## JWT (JSON Web Token)

A JWT is the most common structured token format, made up of three base64url-encoded parts separated by dots: `header.payload.signature`.

```
eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI0MiIsImV4cCI6MTcwMDAwMDAwMH0.4f3c8a...
```

- **Header** — declares the signing algorithm (e.g. `HS256`, `RS256`).
- **Payload** — the claims: arbitrary data such as a user ID, an expiration time (`exp`), and an issuer (`iss`). The payload is only encoded, not encrypted — it is readable by anyone who intercepts the token, so sensitive data (passwords, secrets) is never placed inside it.
- **Signature** — proves the token hasn't been tampered with. Computed by signing the header and payload with a secret (HMAC) or a private key (RSA/ECDSA).

Issuing a JWT in Go, using the widely used `github.com/golang-jwt/jwt/v5` package:

```go
import "github.com/golang-jwt/jwt/v5"

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func generateToken(userID uint) (string, error) {
    claims := jwt.MapClaims{
        "sub": userID,
        "exp": time.Now().Add(15 * time.Minute).Unix(),
        "iat": time.Now().Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
```

Validating one:

```go
func parseToken(tokenString string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return jwtSecret, nil
    })
    if err != nil {
        return nil, err
    }
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    return claims, nil
}
```

Checking the signing method explicitly inside the key function is a deliberate safeguard — without it, a token forged with a different algorithm (such as `none`) could bypass verification entirely, a well-known class of JWT vulnerability.

## Password-based authentication (issuing the first token)

Before any token can be issued, the resource owner's identity is established — typically with a password, hashed and compared, never stored or compared in plaintext:

```go
import "golang.org/x/crypto/bcrypt"

func hashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}

func checkPassword(hash, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

A login handler combining password verification with token issuance:

```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    user, err := findUserByEmail(req.Email)
    if err != nil || !checkPassword(user.PasswordHash, req.Password) {
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }

    accessToken, _ := generateToken(user.ID)
    refreshToken, _ := generateRefreshToken(user.ID) // longer expiry, stored server-side

    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
    })
}
```

## Protecting routes with middleware

Once tokens are being issued, incoming requests are checked for a valid one before reaching a handler — the same middleware pattern introduced in section 6:

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        header := r.Header.Get("Authorization")
        if !strings.HasPrefix(header, "Bearer ") {
            http.Error(w, "missing token", http.StatusUnauthorized)
            return
        }
        tokenString := strings.TrimPrefix(header, "Bearer ")

        claims, err := parseToken(tokenString)
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), userIDKey, claims["sub"])
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

This is the same request-scoped context pattern from section 8 — the authenticated user's identity, once established, is attached to the context so downstream handlers can read it without re-parsing the token.

## Authorization — scopes and roles

Once a request is authenticated, a separate check decides whether it's *permitted*. Two common models:

**Scopes** — a token carries a list of permissions it was granted (e.g. `posts:read`, `posts:write`), checked against what an endpoint requires:

```go
func requireScope(scope string, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        scopes := r.Context().Value(scopesKey).([]string)
        if !slices.Contains(scopes, scope) {
            http.Error(w, "insufficient permissions", http.StatusForbidden)
            return
        }
        next(w, r)
    }
}
```

**Roles** — a simpler model where a user (rather than a token) has a role such as `admin` or `member`, checked directly:

```go
func requireAdmin(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        role := r.Context().Value(roleKey).(string)
        if role != "admin" {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }
        next(w, r)
    }
}
```

Authentication answers whether a request comes from a known identity at all; authorization — via scopes, roles, or ownership checks — answers whether that identity is allowed to perform this specific action.

## Refresh token rotation

Because a refresh token is long-lived, reusing the same one indefinitely increases the impact of a leak. A common mitigation is **rotation**: each time a refresh token is used, it is invalidated and replaced with a new one, and if an already-used (and therefore now-invalid) refresh token is presented again, every token in that chain is revoked — since reuse strongly suggests the token was stolen and used by two different parties.

```go
func handleRefresh(w http.ResponseWriter, r *http.Request) {
    var req struct{ RefreshToken string `json:"refresh_token"` }
    json.NewDecoder(r.Body).Decode(&req)

    stored, err := findRefreshToken(req.RefreshToken)
    if err != nil || stored.Revoked {
        revokeAllTokensForUser(stored.UserID) // reuse detected — treat as compromised
        http.Error(w, "invalid refresh token", http.StatusUnauthorized)
        return
    }

    revokeRefreshToken(stored.ID)
    newAccessToken, _ := generateToken(stored.UserID)
    newRefreshToken, _ := generateRefreshToken(stored.UserID)

    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  newAccessToken,
        "refresh_token": newRefreshToken,
    })
}
```

## Session-based auth, for comparison

Token-based auth (as above) is one model; session-based auth is the older alternative, still common and often simpler for a traditional server-rendered or same-origin application. Instead of a self-contained token, the server stores session state and gives the client only an opaque session ID, usually in a cookie:

```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
    // ... verify credentials ...
    sessionID := generateRandomID()
    saveSession(sessionID, user.ID, time.Now().Add(24*time.Hour))

    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    sessionID,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    })
}
```

The core tradeoff: a session can be invalidated instantly server-side (useful for "log out everywhere"), while a JWT-style access token remains valid until it expires, since nothing forces the resource server to check back with an authority before honoring it. Many systems combine both — a short-lived stateless access token for regular requests, and a server-tracked, revocable refresh token underneath it, which is exactly the pattern shown above.

## OpenID Connect — adding authentication back in

OAuth2 alone only proves that a client holds a valid access token for some resource — it does not by itself standardize *who the resource owner is*. **OpenID Connect (OIDC)** is a thin identity layer built on top of OAuth2 that adds a third token type, the **ID token** — a JWT containing standardized identity claims (`sub`, `email`, `name`) about the resource owner, issued alongside the access token from the same flow. This is the piece that turns a plain OAuth2 exchange into an actual "log in with X" experience, and it's why real-world third-party login integrations are typically described as OIDC rather than "raw" OAuth2, even though they reuse the exact same authorization code flow described above.
