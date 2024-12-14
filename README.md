# Tokenize

**Tokenize** is a comprehensive payment and subscription management system powered by the Stripe API. It streamlines processes like account creation, user authentication, payment handling, and subscription management, offering an efficient solution for your business needs.

## Features

- **Authentication**: Secure user registration and login with validation.  
- **Payment Management**:  
  - Seamless Stripe subscription integration.  
  - Support for offline payments and Multibanco.  
- **Administrative Tools**:  
  - Role-based user permissions for better access control.  
- **Webhooks**:  
  - Automated handling of events such as customer creation, payment success, and subscription cancellations.  
- **System Health**: A dedicated endpoint to monitor system status.  

## Endpoints

### User Management

- **POST** `/create-user`  
  **Description**: Creates a new user account.  
  **Request**:  
  ```json
  {
    "username": "string",
    "password": "string",
    "email": "string"
  }
  ```
  **Responses**:  
  - **200 OK**: User created successfully, returns `{"id": <user_id>}`.  
  - **400 Bad Request**: Invalid payload or email already in use.  
  - **401 Unauthorized**: User is already logged in.  
  - **405 Method Not Allowed**: The method is not `POST`.  

- **POST** `/login-user`  
  **Description**: Authenticates a user and starts a session.  
  **Request**:  
  ```json
  {
    "email": "string",
    "password": "string"
  }
  ```
  **Responses**:  
  - **200 OK**: Successful login, sets authentication cookies.  
  - **400 Bad Request**: Invalid payload or email format.  
  - **401 Unauthorized**: Invalid credentials.  
  - **405 Method Not Allowed**: The method is not `POST`.  

- **GET** `/logout-user`  
  **Description**: Ends the user's session by clearing authentication cookies.  
  **Responses**:  
  - **200 OK**: Successfully logged out.  
  - **401 Unauthorized**: User is not logged in.  

- **GET** `/isActive`  
  **Description**: Checks if the current user session is active.  
  **Responses**:  
  - **200 OK**: Returns `{"active": true}` if the user is active, otherwise `{"active": false}`.  

- **GET** `/isActiveID`  
  **Description**: Checks if a user identified by their ID is active.  
  **Responses**:  
  - **200 OK**: Returns `{"active": true}` if the user is active, otherwise `{"active": false}`.  
  - **400 Bad Request**: Invalid user ID.  
  - **401 Unauthorized**: User is not logged in.  

### Payment Management

- **POST** `/create-checkout-session`  
  **Description**: Initiates a Stripe subscription checkout session.  
  **Responses**:  
  - **303 See Other**: Redirects to the Stripe Checkout session URL.  
  - **401 Unauthorized**: User is not logged in.  
  - **500 Internal Server Error**: Failed to create the session.  

- **GET** `/create-portal-session`  
  **Description**: Creates a Stripe Billing portal session for managing subscriptions.  
  **Responses**:  
  - **303 See Other**: Redirects to the Stripe Billing portal.  
  - **401 Unauthorized**: User is not logged in.  
  - **400 Bad Request**: Invalid customer ID.  

- **POST** `/multibanco`  
  **Description**: Creates a payment session using the Multibanco method for subscription payments.  
  **Responses**:  
  - **303 See Other**: Redirects to the Stripe Checkout session URL.  
  - **400 Bad Request**: Invalid payload or user information.  
  - **401 Unauthorized**: User is not logged in.  
  - **500 Internal Server Error**: Failed to create the session.  

### Webhooks

- **POST** `/webhook`  
  **Description**: Processes Stripe webhook events related to subscriptions, payments, and customer management.  
  **Responses**:  
  - **200 OK**: Webhook processed successfully.  
  - **400 Bad Request**: Invalid webhook payload.  

### System Health

- **GET** `/health`  
  **Description**: Checks the operational status of the system.  
  **Responses**:  
  - **200 OK**: System is operational.  

### System

- **GET** `/getPrecoSub`  
  **Description**: Retrieves the price of the subscription.  
  **Responses**:  
  - **200 OK**: Returns the subscription price.  
  - **500 Internal Server Error**: Failed to retrieve the price.  

## .env Configuration

| Variable                      | Description                                             | Example                       |
|-------------------------------|---------------------------------------------------------|-------------------------------|
| `SECRET_KEY`                  | Your Stripe secret key                                  | `sk_test_9SN7...`            |
| `PUBLISHABLE_KEY`             | Your Stripe publishable key                            | `pk_test_9SN7...`            |
| `SUBSCRIPTION_PRICE_ID`       | The Stripe subscription price ID                       | `price_0DEJ47...`            |
| `ENDPOINT_SECRET`             | Your Stripe webhook secret key                         | `whsec_9274...`              |
| `DOMAIN`                      | Your application domain                                | `http://localhost:4242`      |
| `SECRET_ADMIN`                | Secret key for offline payment authorization           | `sk_test_9SN7...`            |
| `LOGS_FILE`                   | Path to your log file                                  | `logs.txt`                   |
| `NUMBER_OF_SUBSCRIPTIONS_MONTHS` | Number of months per subscription cycle              | `12`                         |
| `STARTING_DATE`               | Specific subscription start date (or `0/0` for normal) | `1/12`                      |

### Mouros Specific
| Variable                      | Description                                             | Example                       |
|-------------------------------|---------------------------------------------------------|-------------------------------|
| `MOUROS_STARTING_DATE`               | Mouros start date (or `0/0` for normal) | `1/12`                      |
| `MOUROS_ENDING_DATE`               | Mouros end date (or `0/0` for normal) | `30/12`                      |
| `COUPON_ID`               | coupon for mouros (or `0/0` for normal) | `bsguqRh0`                      |

## Types of Subscriptions

- **Normal**  
  Standard subscription where users pay, gain access, and renew after a fixed period.  
  `STARTING_DATE` should be `0/0`.

- **OnlyStartOnDayX**  
  The subscription cost is prorated based on the months remaining in the year. For example, with `STARTING_DATE` set to `1/1` (January 1st), subscribing in June costs half the yearly price (6 months).

- **OnlyStartOnDayXNoSubscription**  
  Similar to `OnlyStartOnDayX`, but access is only granted once the `STARTING_DATE` is reached.

- **Mouros**  
  Mouros is a specific type of subscription that the idea is you pay a full year even if you just started it or finishing, first year is a defined price but next ones have a coupon added, the motivation behind this project.

## Setup Instructions

### 1. Install Dependencies
Ensure you have [Go](https://golang.org/) installed. Use the following command to install all required dependencies:
```bash
go mod tidy
