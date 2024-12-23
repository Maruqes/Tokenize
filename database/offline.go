package database

import (
	"fmt"
	"log"
	"time"
)

type Date struct {
	Day   int
	Month int
	Year  int
}

type DateInterface interface {
	String(date string) (Date, error)
}

func (d Date) String() string {
	return fmt.Sprintf("%02d/%02d/%04d", d.Day, d.Month, d.Year)
}

func StringToDate(date string) (Date, error) {
	var dateObj Date
	_, err := fmt.Sscanf(date, "%d/%d/%d", &dateObj.Day, &dateObj.Month, &dateObj.Year)
	if err != nil {
		log.Fatal(err)
	}

	if dateObj.Day < 1 || dateObj.Day > 31 {
		return dateObj, fmt.Errorf("invalid day")
	}
	if dateObj.Month < 1 || dateObj.Month > 12 {
		return dateObj, fmt.Errorf("invalid month")
	}
	if dateObj.Year < 1900 || dateObj.Year > 3000 {
		return dateObj, fmt.Errorf("invalid year")
	}

	return dateObj, nil
}

func DateFromUnix(unix int64) Date {
	t := time.Unix(unix, 0).UTC()
	return Date{Day: t.Day(), Month: int(t.Month()), Year: t.Year()}
}

type OfflinePayment struct {
	UserID        int
	DateOfPayment Date
	End_date      Date
}

func CreateOfflineTable() {
	query := `
    CREATE TABLE IF NOT EXISTS offline_payments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id TEXT,
		date_of_payment TEXT,
		end_date TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id)
    );
    `
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func AddOfflinePayment(user_id int, date_of_payment Date, end_date Date) error {
	date_string := fmt.Sprintf("%d-%02d-%02d", date_of_payment.Year, date_of_payment.Month, date_of_payment.Day)
	end_date_string := fmt.Sprintf("%d-%02d-%02d", end_date.Year, end_date.Month, end_date.Day)
	query := `INSERT INTO offline_payments (user_id, date_of_payment, end_date) VALUES (?, ?, ?);`
	_, err := db.Exec(query, user_id, date_string, end_date_string)
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
		var date_of_payment_string string
		var end_date_string string
		err = rows.Scan(&id, &user_id, &date_of_payment_string, &end_date_string)
		if err != nil {
			log.Fatal(err)
		}
		date := Date{}
		end_date := Date{}
		fmt.Sscanf(date_of_payment_string, "%d-%d-%d", &date.Year, &date.Month, &date.Day)
		fmt.Sscanf(end_date_string, "%d-%d-%d", &end_date.Year, &end_date.Month, &end_date.Day)

		offlinePayments = append(offlinePayments, OfflinePayment{UserID: user_id, DateOfPayment: date, End_date: end_date})
	}
	return offlinePayments, nil
}
