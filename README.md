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

### System
- **GET** `/getPrecoSub`  
  Get price of subscription



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
  Mouros is a specific type of subscription that the idea is you pay a full year even if you just started it or finishing, first year is a defined price but next ones have a coupon added, the motivation behing this project

## Setup Instructions

### 1. Install Dependencies
Ensure you have [Go](https://golang.org/) installed. Use the following command to install all required dependencies:
```bash
go mod tidy
