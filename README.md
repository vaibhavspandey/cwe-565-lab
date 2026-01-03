- i took 3 hours to design and publish this 9:30am to 12:30pm 
# Lab 1: Software or Data Integrity Failures (CWE-565)

## Overview & Intent

This repository contains a standalone, intentionally vulnerable web application designed for educational purposes. It targets **CWE-565: Reliance on Cookies without Validation and Integrity Checking**, which falls under the broader category of **OWASP Top 10 A08:2025: Software and Data Integrity Failures**.

**Intent of this Lab:**
This lab is designed for software developers. The provided code contains a critical security flaw common in legacy or optimized web applications. Your goal is not just to run the exploit, but to understand the architectural failure and implement a secure, industry-standard fix in the Go backend.

### Understanding the Vulnerability

**OWASP A08:2025 (Software and Data Integrity Failures)** focuses on making assumptions about software updates, critical data, and CI/CD pipelines without verifying their integrity.

**CWE-565** is a specific manifestation of this failure where an application relies on the existence or values of cookies to make security decisions (such as role authorization) but fails to verify that the cookie has not been modified by the client. Because cookies are stored client-side, they are fully under the user's control and must be treated as untrusted input.

---

## Lab Setup

To start the Gopher Financial banking environment, you must have Docker installed.

1. **Build the Container:**
```bash
docker build -t cwe-565-lab .

```


2. **Run the Application:**
```bash
docker run -p 8080:8080 cwe-565-lab

```


3. **Access the Application:**
Navigate to `http://localhost:8080` in your web browser.

---

## Objective

By the end of this lab, you will:

1. **Analyze** a Go web application to identify where trust boundaries are violated.
2. **Exploit** the application using Chrome DevTools to elevate privileges from a standard user to a system administrator.
3. **Patch** the vulnerability by implementing Server-Side Session Lookups.
4. **Verify** the fix using negative testing techniques.

---

## Phase 1: Understanding the Vulnerability

### Vulnerable Code Analysis

**Location:** `backend/main.go`

The application implements a "performance optimization" where the user's role is cached in a cookie upon login. This allows the dashboard to render UI elements without querying the database. However, the backend also relies on this cookie for access control to sensitive API endpoints.

**VULNERABLE CODE (Access Control Logic):**

```go
func adminHandler(c *gin.Context) {
    // VULNERABLE: The application trusts the client-provided cookie
    roleCookie, _ := c.Cookie("is_admin")

    // The integrity of 'roleCookie' is never verified
    if roleCookie != "true" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied"})
        return
    }

    // ... Sensitive Data Release ...
}

```

**Why It Is Vulnerable:**

1. **Untrusted Source:** The application treats the `is_admin` cookie as a trusted source of truth.
2. **Lack of Integrity:** There is no digital signature (HMAC) or encryption protecting the cookie value.
3. **Broken Trust Boundary:** The server assumes that because it set the cookie during login, the value remains unchanged.

---

## Phase 2: Exploitation Techniques

### The Attack

**Target:** Elevate privileges from "Standard User" (Alice) to "Admin" to access the confidential user database.

1. **Login as a Standard User:**
* **Username:** `alice`
* **Password:** `password123`
* **Observation:** The dashboard loads. Note that the "Admin Administration Console" is hidden.


2. **Inspect the Traffic:**
* Open Chrome DevTools (`F12` or Right Click > Inspect).
* Navigate to the **Application** tab.
* In the sidebar, expand **Cookies** > `http://localhost:8080`.
* Observe the following cookies:
* `session_id`: A UUID string (e.g., `550e8400-e29b...`).
* `is_admin`: Set to `false`.




3. **Modify the Integrity:**
* Double-click the value of the `is_admin` cookie.
* Change the value from `false` to `true`.
* Press Enter to save.


4. **Execute the Exploit:**
* Refresh the page.
* **Result:** The red "Admin Administration Console" box appears on the dashboard.
* Click the **"Access User Database"** button.
* **Impact:** The server returns a JSON dump containing sensitive PII (SSNs, Credit Scores, Addresses) for all users.



---

## Phase 3: Fixing The Vulnerable Code

The most robust fix is to stop relying on client-side state for authorization. We will switch to a **Server-Side Session Lookup**.

### The Fix

We will modify `backend/main.go`. Instead of checking the `is_admin` cookie, we will use the trusted `session_id` (validated by middleware) to look up the user's role directly from the database.

**BEFORE (Vulnerable):**

```go
// In adminHandler
roleCookie, _ := c.Cookie("is_admin")
if roleCookie != "true" {
    c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied"})
    return
}

```

**AFTER (Secure):**

```go
// In adminHandler

// 1. Retrieve the validated User ID from the context (set by authMiddleware)
userID, exists := c.Get("user_id")
if !exists {
    c.AbortWithStatus(http.StatusUnauthorized)
    return
}

// 2. Query the "Source of Truth" (The Database) for the role
var role string
err := DB.QueryRow("SELECT role FROM users WHERE id=?", userID).Scan(&role)

// 3. Check permissions based on the DB result, NOT the cookie
if err != nil || role != "admin" {
    c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied: Insufficient Privileges"})
    return
}

```

*Note: You should also update `loginHandler` to stop setting the `is_admin` cookie entirely, as it serves no valid security purpose.*

---

## Phase 4: Testing the Fix

### Positive Testing

1. Rebuild and run the container with the secure code.
2. Login as `admin` (Password: `adminpass`).
3. **Expected:** The Admin Console is visible, and the API returns data.

### Negative Testing (The "Zombie Cookie" Scenario)

A common oversight during remediation is assuming that fixing the code cleans up the client state. We must verify that the server ignores the malicious cookie even if it persists in the browser.

1. Login as `alice` (`alice` / `password123`).
2. Open DevTools and manually set `is_admin` to `true`.
3. Refresh the page.
4. **Expected Result:**
* The Admin Console should **NOT** appear.
* Accessing `/admin/users` directly should return `403 Forbidden`.
* This confirms the application is ignoring the client-side integrity violation and relying on the database.



---

## Prevention Best Practices

1. **Single Source of Truth:** Authorization decisions must always be made based on server-side state (Session Store, Database, or Redis), never client-side state.
2. **Integrity Checks:** If client-side state is unavoidable (e.g., stateless microservices), use **HMAC signatures** (e.g., JWTs or `gorilla/securecookie`) to sign the data. The server must verify the signature before using the data.
3. **Least Privilege:** Ensure API endpoints validate permissions independently of the UI. Hiding a button in the frontend is not a security control.
4. **Input Validation:** Treat all HTTP headers, cookies, and query parameters as untrusted user input.

## Additional Resources

* [OWASP Top 10: A08:2021 Software and Data Integrity Failures](https://owasp.org/Top10/A08_2021-Software_and_Data_Integrity_Failures/)
* [CWE-565: Reliance on Cookies without Validation and Integrity Checking](https://cwe.mitre.org/data/definitions/565.html)
* [Go Database/SQL Package Documentation](https://pkg.go.dev/database/sql)
