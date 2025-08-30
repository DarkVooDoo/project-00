package model

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type Review struct{
	Id int64 `json:"id"`
	UserName string
	Comment string `json:"comment"`
	Rating float32 `json:"rating,string"`
	Key string
	Date string
	StarCount []int
	EtablishmentId int64 `json:"etablishmentId"`
	EmployeeId int64 `json:"employeeId"`
	UserId int64 `json:"userId"`
}

func (r *Review) Get(conn *sql.Conn)error{
	reviewGetRow := conn.QueryRowContext(context.Background(), `SELECT r.id, TO_CHAR(LOWER(a.date), 'DD TMMonth YYYY à HH24:MI'), r.review_key, r.employee_id, r.etablishment_id, 
	u.firstname || ' ' || u.lastname FROM review AS r LEFT JOIN employee AS e ON e.id=r.employee_id LEFT JOIN users AS u ON u.id=e.user_id 
	LEFT JOIN appointment AS a ON a.id=r.appointment_id WHERE r.user_id=$1 AND r.rating IS  NULL ORDER BY r.created_at DESC`, r.UserId)
	if err := reviewGetRow.Scan(&r.Id, &r.Date, &r.Key, &r.EmployeeId, &r.EtablishmentId, &r.UserName); err != nil{
		log.Printf("error scanning the review: %s", err)
		return errors.New("error getting the review")
	}
	return nil
}

func (r *Review) EtablishmentReview(conn *sql.Conn)float64{
	var rating float64
	reviewRows := conn.QueryRowContext(context.Background(), `SELECT AVG(rating) FROM review WHERE etablishment_id=$1`, r.EtablishmentId)
	if err := reviewRows.Scan(&rating); err != nil{
		log.Printf("error scanning review: %s", err)
		return rating
	}
	return rating
}

func(r *Review) Update(conn *sql.Conn)error{
	reviewUpdateRow, err := conn.ExecContext(context.Background(), `UPDATE review SET comment=$1, rating=$2 WHERE id=$3 AND user_id=$4`, r.Comment, r.Rating, r.Id, r.UserId)
	if err != nil{
		log.Printf("error executing the update query: %s", err)
		return errors.New("error in the query")
	}
	aff, err := reviewUpdateRow.RowsAffected()
	if aff == 0 || err != nil{
		log.Printf("error no rows affected: %d error: %s", aff, err)
		return errors.New("error no rows has been affected")
	}
	return nil
}

func (r *Review) Delete(conn *sql.Conn)error{
	reviewDeleteRow, err := conn.ExecContext(context.Background(), `DELETE FROM review WHERE id=$1 AND user_id=$2`, r.Id, r.UserId)
	if err != nil{
		log.Printf("error deleting the review: %s", err)
		return errors.New("error deleting the row")
	}
	aff, err := reviewDeleteRow.RowsAffected()
	if err != nil || aff == 0{
		log.Printf("error nothing deleted: %s", err)
		return errors.New("error nothing deleted")
	}
	return nil
}
