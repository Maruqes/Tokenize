# Tokenize

**Tokenize** is a payment and subscription management system integrated with the Stripe API. It provides features for account creation, user management, authentication, payment processing, and more.

## Features

- **Authentication**: User registration and login with validation.
- **Payment Management**:
  - Subscriptions via Stripe.
  - Offline payments.
- **Admin Functions**:
  - Configurable permissions for different user roles.
- **Webhooks**:
  - Automated responses to events like customer creation, payment success, and subscription cancellations.
- **System Health**: Endpoint to check the system's status.

## Endpoints

### User Management
- **POST** `/create-user`  
  Create a new user account.

- **POST** `/login-user`  
  Authenticate a user and start a session.

- **GET** `/logout-user`  
  Log out the current user.

### Payment Management
Only work if user is logged in
- **POST** `/create-checkout-session`  
  Initiate a Stripe subscription checkout session.

- **GET** `/create-portal-session`  
  Create a Stripe Billing portal session for managing subscriptions.

### Webhooks
- **POST** `/webhook`  
  Handle Stripe webhook events for subscriptions, payments, and customer management.

### System Health
- **GET** `/health`  
  Check the status of the system.

### Offline payments
- **GET** `/pay-offline`  
  Pay offline, a superuser with a "SECRET_ADMIN" can create a account as it is being payed with stripe but manual, for example if you want to have the option to pay in cash 

- **GET** `/get-offline-last-time`  
  Get when someone subscription is ending (offline mode)

- **GET** `/get-offline-id`  
  Get the payments by user_id payd with "/pay-offline"


## .env File
- **"SECRET_KEY"**
Your secret stripe key example -> "sk_test_9SN7..."

- **"PUBLISHABLE_KEY"**
Your publishable stripe key example -> "pk_test_9SN7..."

- **"SUBSCRIPTION_PRICE_ID"**
The price id for the subscription stripe key example -> "price_0  DEJ47..."

- **"ENDPOINT_SECRET"**
Your endpoint secret stripe key example -> "whsec_9274..."

- **"DOMAIN"**
Your domain example -> "http://localhost:4242"

- **"SECRET_ADMIN"**
Your secret admin key (the one used for offline payments) example -> "sk_test_9SN7..."

- **"LOGS_FILE"**
Your logs file path key example -> "logs.txt"

- **"NUMBER_OF_SUBSCRIPTIONS_MONTHS"**
The number of months your subscription is -> "12"



## Setup Instructions

### 1. Install dependencies
Make sure you have [Go](https://golang.org/) installed. Run the following command to install all necessary dependencies:
```bash
go mod tidy
