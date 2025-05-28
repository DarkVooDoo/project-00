package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/lib/pq"
)

type EtablishmentSchedule struct{
    From []string `json:"from"`
    To []string `json:"to"`
}

type SchedulePayload struct{
    EtablishmentSchedule
    Id string `json:"id"`
}

type DaySchedule struct{
	Day string
	Time string
}

type KeyValue struct{
    Id int
    Value string
}

type Etablishment struct{
    Id int
    Name string `json:"name"`
    Adresse string `json:"adresse"`
    Postal string `json:"postal"`
    Phone string `json:"phone"`
    Lat float64
    Lon float64
    Payment []string `json:"payment"`
    AllPayment []string 
    Employee []Employe
    Service []Service
    Schedule []DaySchedule
    TodaySchedule string
    IsOpen string
    Category string `json:"category"`
    UserId int
}

var Week = []string{"Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi", "Dimanche"}

func (e *Etablishment) Create()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    etablishmentRow:= conn.QueryRowContext(context.Background(), `INSERT INTO etablishment(name, adresse, postal, phone, payment, category_id, user_id) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, 
    e.Name, e.Adresse, e.Postal, e.Phone, pq.Array(e.Payment), e.Category, e.UserId)
    if err := etablishmentRow.Scan(&e.Id); err != nil{
        log.Printf("error scanning the query: %s", err)
        return errors.New("error nothing happend")
    }
    return nil
}

func(e *Etablishment) Latest(conn *sql.Conn)[]Etablishment{
    var list []Etablishment

    rows, err := conn.QueryContext(context.Background(), `SELECT e.id, e.name, adresse, postal, COALESCE(phone, 'N/A'), c.name, CONCAT(e.schedule->'from'->>EXTRACT(ISODOW FROM NOW())::int - 1, ' - ',
    e.schedule->'to'->>EXTRACT(ISODOW FROM NOW())::int - 1) AS schedule FROM etablishment AS e LEFT JOIN category AS c ON e.category_id=c.id LIMIT 5`)
    if err != nil{
        log.Printf("error getting etablishment: %s", err)
        return list
    }
    for rows.Next(){
        if err := rows.Scan(&e.Id, &e.Name, &e.Adresse, &e.Postal, &e.Phone, &e.Category, &e.TodaySchedule); err != nil{
            log.Printf("error scanning rows: %s", err)
        }
        if len(e.TodaySchedule) < 4 {e.TodaySchedule = "Fermé"}
        list = append(list, *e)
    }
    return list
}

func (e *Etablishment) Parametre(conn *sql.Conn)error{
    etablishmentRow := conn.QueryRowContext(context.Background(), `SELECT e.name, e.adresse, e.postal, COALESCE(e.phone, ''), c.id, e.payment, enum_range(NULL::payment_type) 
    FROM etablishment AS e LEFT JOIN category AS c ON c.id=e.category_id WHERE e.id=$1 AND e.user_id=$2`, e.Id, e.UserId)
    if err := etablishmentRow.Scan(&e.Name, &e.Adresse, &e.Postal, &e.Phone, &e.Category, pq.Array(&e.Payment), pq.Array(&e.AllPayment)); err != nil{
        log.Printf("error scanning the etablishment row: %s", err)
        return errors.New("error in the query etablishment")
    }
    return nil 
}

func (e *Etablishment) UserEtablishments(conn *sql.Conn)([]Etablishment, error){
    var list []Etablishment
    etablishmentList, err := conn.QueryContext(context.Background(), `SELECT id, name FROM etablishment WHERE user_id=$1`, e.UserId)
    if err != nil{
        log.Printf("error in query etablishments: %s", err)
        return list, errors.New("error getting the etablishments")
    }
    for etablishmentList.Next(){
        if err = etablishmentList.Scan(&e.Id, &e.Name); err != nil{
            continue
        }
        list = append(list, *e)
    }
    return list, nil
}

func (e *Etablishment) GetEmployeeAndService(conn *sql.Conn){
    var employee Employe
    var service Service
    var name  sql.NullString
	var id sql.NullInt64
    //TODO: optimizer les requetes a la DB on doit se connecter 2 fois
    employeeList, err := conn.QueryContext(context.Background(), `SELECT em.id, u.firstname || ' ' || u.lastname, e.name, e.id FROM etablishment AS e 
    LEFT JOIN employee AS em ON em.etablishment_id=e.id LEFT JOIN users AS u ON u.id=em.user_id WHERE e.id=$1`, e.Id)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return
    }
    for employeeList.Next(){
        if err = employeeList.Scan(&id, &name, &e.Name, &e.Id); err != nil{
            log.Printf("error scan: %s", err)
        }
        employee.Name = name.String
        employee.Id = int(id.Int64)
        e.Employee = append(e.Employee, employee)
    }

    serviceList, err := conn.QueryContext(context.Background(), `SELECT s.id, s.name, s.price, s.duration FROM service AS s WHERE s.etablishment_id=$1`, e.Id)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return
    }
    for serviceList.Next(){
        if err = serviceList.Scan(&service.Id, &service.Name, &service.Price, &service.Duration); err != nil{
            log.Printf("error scan: %s", err)
        }
        e.Service = append(e.Service, service)
    }
}

func (e *Etablishment) Public(conn *sql.Conn)(int, error){
    var schedule, phone  sql.NullString
    var service Service
    var jsonSchedule EtablishmentSchedule
    var weekDay int
    etablishmentRow, err := conn.QueryContext(context.Background(), `SELECT e.name, e.payment, e.phone, e.adresse, e.postal, e.schedule, s.name, s.price, s.description, s.duration, 
    c.name, EXTRACT(ISODOW FROM NOW()) - 1, CASE WHEN e.schedule->'from'->>EXTRACT(ISODOW FROM NOW())::INT - 1 = '' THEN 'Actuellement Fermé'
    WHEN TSRANGE(NOW()::DATE + (e.schedule->'from'->>EXTRACT(ISODOW FROM NOW())::INT - 1)::TIME, NOW()::DATE + (e.schedule->'to'->>EXTRACT(ISODOW FROM NOW())::INT - 1)::TIME) @> NOW()::TIMESTAMP  
    THEN 'Actuellement Ouvert' ELSE 'Actuellement Fermé' END
    FROM etablishment AS e LEFT JOIN service AS s ON s.etablishment_id=e.id LEFT JOIN category AS c ON c.id=e.category_id WHERE e.id=$1`, e.Id)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return weekDay, errors.New("error in the query")
    }
    for etablishmentRow.Next(){
        if err := etablishmentRow.Scan(&e.Name, pq.Array(&e.Payment), &phone, &e.Adresse, &e.Postal, &schedule, &service.Name, &service.Price, &service.Description, 
        &service.Duration,  &e.Category, &weekDay, &e.IsOpen); err != nil{
            log.Printf("error scanning the schedule: %s", err)
            continue
        }
        e.Service = append(e.Service, service)
    }

    if e.Name == ""{
        return weekDay, errors.New("error no data")
    }
    if err := json.Unmarshal([]byte(schedule.String), &jsonSchedule); err != nil{
        log.Printf("error unmashal data: %s", err)
    }
    for i, v := range jsonSchedule.From{
        if v == "" || jsonSchedule.To[i] == ""{
            e.Schedule = append(e.Schedule, DaySchedule{Week[i], "Fermé"})
            continue
        }
        e.Schedule = append(e.Schedule,  DaySchedule{Week[i], fmt.Sprintf("%s - %s", v, jsonSchedule.To[i])})
        e.Phone = phone.String
    }
    return weekDay, nil
}

func (e *Etablishment) GetSchedule()(EtablishmentSchedule){
    conn := GetDBPoolConn()
    defer conn.Close()

    var schedule sql.NullString
    var jsonSchedule EtablishmentSchedule
    scheduleRow := conn.QueryRowContext(context.Background(), `SELECT schedule FROM etablishment WHERE id=$1`, e.Id)
    if err := scheduleRow.Scan(&schedule); err != nil{
        log.Printf("error scanning the schedule: %s", err)
        return jsonSchedule
    }

    if err := json.Unmarshal([]byte(schedule.String), &jsonSchedule); err != nil{
        log.Printf("error unmashal data: %s", err)
        return jsonSchedule
    
    }
    return jsonSchedule
}

func (e *Etablishment) UpdateSchedule(schedule EtablishmentSchedule, id string)error{
    conn := GetDBPoolConn()
    defer conn.Close()

    sch, _ := json.Marshal(schedule)
    result, err := conn.ExecContext(context.Background(), `UPDATE etablishment SET schedule=$1 WHERE id=$3 AND user_id=$2`, string(sch), e.UserId, id)
    if err != nil{
        log.Printf("error updating the etablishment schedule: %s", err)
        return errors.New("error in the query")
    }
    affected, err := result.RowsAffected()
    if affected == 0 || err != nil{
        log.Printf("no update occured")
        return errors.New("zero update occured")
    }

    return nil
}

func Categorys(conn *sql.Conn)[]KeyValue{

    var categorys []KeyValue
    var keyValue KeyValue
    categoryRows, err := conn.QueryContext(context.Background(), `SELECT id, name FROM category`)
    if err != nil{
        log.Printf("error in the query category: %s", err)
        return categorys
    }
    for categoryRows.Next(){
        if err = categoryRows.Scan(&keyValue.Id, &keyValue.Value); err != nil{
            log.Printf("error scanning category: %s", err)
        }
        categorys = append(categorys, keyValue)
    }
    return categorys
}
