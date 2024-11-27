# Tokenize

**Tokenize** is a comprehensive payment and subscription management system powered by the Stripe API. It streamlines processes like account creation, user authentication, payment handling, and subscription management, offering an efficient solution for your business needs.

## Features

- **Authentication**: Secure user registration and login with validation.
- **Payment Management**:
  - Seamless Stripe subscription integration.
  - Support for offline payments.
- **Administrative Tools**:
  - Role-based user permissions for better access control.
- **Webhooks**:
  - Automated handling of events such as customer creation, payment success, and subscription cancellations.
- **System Health**: A dedicated endpoint to monitor system status.

## Endpoints

### User Management
- **POST** `/create-user`  
  Create a new user account.

- **POST** `/login-user`  
  Authenticate a user and initiate a session.

- **GET** `/logout-user`  
  End the session for the logged-in user.

### Payment Management
> These endpoints require the user to be logged in.

- **POST** `/create-checkout-session`  
  Initiate a Stripe subscription checkout session.

- **GET** `/create-portal-session`  
  Create a Stripe Billing portal session for subscription management.

### Webhooks
- **POST** `/webhook`  
  Process Stripe webhook events related to subscriptions, payments, and customer management.

### System Health
- **GET** `/health`  
  Verify the operational status of the system.

### Offline Payments
- **GET** `/pay-offline`  
  Facilitate manual payments for offline subscriptions. A superuser with the `SECRET_ADMIN` key can create accounts that simulate Stripe subscriptions for manual payment methods like cash.

- **GET** `/get-offline-last-time`  
  Retrieve the expiration date of an offline subscription.

- **GET** `/get-offline-id`  
  Fetch payment records by user ID for offline transactions processed via `/pay-offline`.

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
| `STARTING_DATE`               | Specific subscription start date (or `0/0` for normal) | `12/20`                      |

## Types of Subscriptions

- **Normal**  
  Standard subscription where users pay, gain access, and renew after a fixed period.  
  `STARTING_DATE` should be `0/0`.

- **OnlyStartOnDayX**  
  The subscription cost is prorated based on the months remaining in the year. For example, with `STARTING_DATE` set to `1/1` (January 1st), subscribing in June costs half the yearly price (6 months).

- **OnlyStartOnDayXNoSubscription**  
  Similar to `OnlyStartOnDayX`, but access is only granted once the `STARTING_DATE` is reached.

## Setup Instructions

### 1. Install Dependencies
Ensure you have [Go](https://golang.org/) installed. Use the following command to install all required dependencies:
```bash
go mod tidy
