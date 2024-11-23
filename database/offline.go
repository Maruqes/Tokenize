package database

import (
	"fmt"
	"log"
)

type Date struct {
	Day   int
	Month int
	Year  int
}

type OfflinePayment struct {
	UserID        int
	DateOfPayment Date
	Quantity      int
}

func CreateOfflineTable() {
	query := `
    CREATE TABLE IF NOT EXISTS offline_payments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id TEXT,
		date_of_payment TEXT,
		quantity INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id)
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func AddOfflinePayment(user_id int, date_of_payment Date, quantity int) error {
	date_string := fmt.Sprintf("%d-%02d-%02d", date_of_payment.Year, date_of_payment.Month, date_of_payment.Day)
	query := `INSERT INTO offline_payments (user_id, date_of_payment, quantity) VALUES (?, ?, ?);`
	_, err := db.Exec(query, user_id, date_string, quantity)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func DoesUserHasOfflinePayments(user_id int) (bool, error) {
	query := `SELECT * FROM offline_payments WHERE user_id = ?;`
	rows, err := db.Query(query, user_id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func GetOfflinePaymentByID(user_id int) ([]OfflinePayment, error) {
	query := `SELECT * FROM offline_payments WHERE user_id = ?;`
	rows, err := db.Query(query, user_id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var offlinePayments []OfflinePayment

	for rows.Next() {
		var id int
		var user_id int
		var date_of_payment string
		var quantity int
		err = rows.Scan(&id, &user_id, &date_of_payment, &quantity)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, user_id, date_of_payment, quantity)
		date := Date{}
		fmt.Sscanf(date_of_payment, "%d-%d-%d", &date.Year, &date.Month, &date.Day)
		offlinePayments = append(offlinePayments, OfflinePayment{UserID: user_id, DateOfPayment: date, Quantity: quantity})
	}
	return offlinePayments, nil
}
