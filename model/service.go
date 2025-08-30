package model

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
)

type Service struct{
    Id int64 `json:"id,string"`
    Name string `json:"name"`
    Price string `json:"price"`
	DiscountPrice string
    Duration int `json:"duration,string"`
    Description string `json:"description"`
	Discount int `json:"discount,string"`
    Checked bool
    CurrencyPrice string
    EtablishmentId int64
	AppointmentId int64
}

func (s *Service) Create()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    serviceRow := conn.QueryRowContext(context.Background(), `INSERT INTO service (name, price, duration, description, discount, etablishment_id) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, s.Name, s.Price, s.Duration, s.Description, s.Discount, s.EtablishmentId)
    if err := serviceRow.Scan(&s.Id); err != nil{
        log.Printf("error query inserting service: %s", err)
        return errors.New("error in the query")
    }
    return nil
}

func (s *Service) GetList(conn *sql.Conn)([]Service, error){
    var list []Service

    serviceList, err := conn.QueryContext(context.Background(),`SELECT id, name, price::NUMERIC, price, duration, description, discount FROM service WHERE etablishment_id=$1`, s.EtablishmentId)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return list, errors.New("error in the query")
    }
    for serviceList.Next(){
        if err := serviceList.Scan(&s.Id, &s.Name, &s.Price, &s.CurrencyPrice, &s.Duration, &s.Description, &s.Discount); err != nil{
            log.Printf("error scanning service: %s", err)
        }
        list = append(list, *s)
    }
    return list, nil
}

func (s *Service) Update()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    result, err := conn.ExecContext(context.Background(), `UPDATE service SET name=$1, price=$2, discount=$3, duration=$4, description=$5 WHERE id=$6 AND etablishment_id=$7`, 
	s.Name, strings.ReplaceAll(s.Price, ".", ","), s.Discount, s.Duration, s.Description, s.Id, s.EtablishmentId)
    if err != nil{
        log.Printf("error in the updating query service: %s", err)
        return errors.New("error in the query")
    }
    affected, err := result.RowsAffected()
    if err != nil || affected == 0{
        log.Printf("error nothing happend: %s", err)
        return errors.New("error nothing deleted")
    }
    return nil
}

func (s *Service) Delete()error{
    conn := GetDBPoolConn()
    defer conn.Close()
 
    result, err := conn.ExecContext(context.Background(), `DELETE FROM service WHERE id=$1 AND etablishment_id=$2`, s.Id, s.EtablishmentId)
    if err != nil{
        log.Printf("error deleting service: %s", err)
        return errors.New("error in the query")
    }
    affected, err := result.RowsAffected()
    if err != nil || affected == 0{
        log.Printf("error nothing happend: %s", err)
        return errors.New("error nothing deleted")
    }
    return nil
}
