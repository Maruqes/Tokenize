# Tokenize

Tokenize is an authentication system that allows you to create accounts, perform login, and manage account activation through Stripe. You can use the following endpoints to handle users, and you also have specific functions for managing subscriptions and payments in a flexible way.

## Endpoints

### Create User

**Route:** `/create-user`  
**Method:** `POST`

#### Description
Creates a user account if there is no active session.

#### Request Body (JSON)
- **username**: Username (string)
- **password**: Password (string)
- **email**: Email address (string)

#### Restrictions
- If the user is already authenticated, the endpoint returns `401 Unauthorized`.
- If the HTTP method is not `POST`, it returns `405 Method Not Allowed`.

---

### User Login

**Route:** `/login-user`  
**Method:** `POST`

#### Description
Allows a user with valid credentials to log in.

#### Request Body (JSON)
- **email**: Email address (string)
- **password**: Password (string)

#### Validation
- The request body must be valid JSON.
- If the JSON is malformed, it returns `400 Bad Request`.

---

### User Logout

**Route:** `/logout-user`  
**Method:** `POST`

#### Description
Logs out an authenticated user, removing their access.

---

### Create Portal Session

**Route:** `/create-portal-session`  
**Method:** `POST`

#### Description
Generates a Stripe Billing portal session, allowing the user to manage their subscriptions (e.g., update plan, change payment method, cancel, etc.).

---

## Stripe Integration

The system integrates with Stripe to allow account activation, subscription management, payments, and other billing functionalities. Below are functions you can define or call to handle subscription and payment creation and management.

### Subscription and Payment Functions

These functions give you the flexibility to create additional logic, such as trials, future scheduling, or payment/subscription pages.

#### Subscription Functions

1. **CreateSubscription**
   ```go
   CreateSubscription(
       userID int,
       trial_duration time.Duration,
       PriceID string,
       extraMetadata map[string]string,
   )
   ```
   - Creates a new subscription in Stripe, optionally including a trial period (`trial_duration`).
   - `extraMetadata` allows you to pass additional information to Stripe.

2. **CreateScheduledSubscription**
   ```go
   CreateScheduledSubscription(
       userID int,
       start time.Time,
       trial_duration time.Duration,
       PriceID string,
       extraMetadata map[string]string,
   )
   ```
   - Similar to `CreateSubscription` but allows scheduling the subscription to begin on a specific date (`start`).
   - `extraMetadata` work in the same way.

3. **CreateFreeTrial**
   ```go
   CreateFreeTrial(
       userID int,
       start time.Time,
       duration time.Duration,
       PriceID string,
       extraMetadata map[string]string,
   ) (*stripe.Subscription, error)
   ```
   - Creates a subscription with a free trial period starting at a given date and lasting for a specified `duration`.
   - Returns the Stripe subscription object or an error.

#### Payment Functions

1. **CreatePayment**
   ```go
   CreatePayment(
       userID int,
       amount float64,
       extraMetadata map[string]string,
   )
   ```
   - Performs a one-time payment for the specified `amount`.
   - Use `extraMetadata` to send additional info to Stripe (order ID, customer data, etc.).

2. **CreatePaymentPage**
   ```go
   CreatePaymentPage(
       userID int,
       amount float64,
       imageURL string,
       description string,
       extraMetadata map[string]string,
       success_url string,
       cancel_url string,
   )
   ```
   - Generates a payment page for a one-time payment where the user can enter card details and confirm the payment.
   - `success_url` and `cancel_url` define where the user is redirected after completing or canceling the payment.

3. **CreateSubscriptionPage**
   ```go
   CreateSubscriptionPage(
       userID int,
       priceID string,
       extraMetadata map[string]string,
       success_url string,
       cancel_url string,
   )
   ```
   - Creates a payment page for a subscription, ideal for monthly/annual plans.
   - As with one-time payments, `success_url` and `cancel_url` handle user redirection after the process.
