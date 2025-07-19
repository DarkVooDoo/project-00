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
	Date string
	StarCount []int
	EtablishmentId int64 `json:"etablishmentId"`
	EmployeeId int64 `json:"employeeId"`
	UserId int64 `json:"userId"`
}

func (r *Review) Init(conn *sql.Conn)error{
	reviewRow, err := conn.ExecContext(context.Background(), `INSERT INTO review(etablishment_id, employee_id, user_id) VALUES($1,$2,$3)`, r.EtablishmentId, r.EmployeeId, r.UserId)
	if err != nil{
		log.Printf("error executing the query: %s", err)
		return errors.New("error in the query")
	}
	aff, err := reviewRow.RowsAffected()
	if aff == 0 || err != nil{
		log.Printf("error no rows affected: %d error: %s", aff, err)
		return errors.New("error no rows has been affected")
	}
	return nil
}

func(r *Review) Update(conn *sql.Conn)error{
	reviewUpdateRow, err := conn.ExecContext(context.Background(), `UPDATE review SET comment=$1, rating=$2" WHERE id=$2 AND user_id=$4`, r.Comment, r.Rating, r.Id, r.UserId)
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
	reviewDeleteRow, err := conn.ExecContext(context.Background(), `DELETE FROM review WHERE id=$1 AND user_id=$3`, r.Id, r.UserId)
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
