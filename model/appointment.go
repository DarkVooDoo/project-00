package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/lib/pq"
)

type ServicePayload struct{
    Id string `json:"id"`
    Duration int `json:"duration,string"`
}

type Appointment struct{
    Id string
    EmployeeName string
    CustomerName string
    Adresse string
    Category string
    Services []ServicePayload
    Service string
	ServiceTook string
    Price string
    Contact string
    Status string
    Description string
    UserId int
    Date string `json:"date"`
    Timeframe string
    FormatDate string
    EmployeeId int `json:"employee,string"`
    EtablishmentId int
}

type Appt struct{
    Id string
    Hour int
    Minute int
    Length int
    Name string
    Date string
    Price string
    Service string
}

func (a *Appointment) AvaileblesDates(conn *sql.Conn)[]string{
    var availeblesDates []string

    dates := conn.QueryRowContext(context.Background(), `SELECT GetAvaileblesDates($1, $2) FROM employee AS e WHERE id=$1`, a.EmployeeId, a.Date)
    if err := dates.Scan(pq.Array(&availeblesDates)); err != nil{
        log.Printf("error scanning the row: %s", err)
        return availeblesDates
    }
    return availeblesDates
}

func (a *Appointment) EmployeePlanning(conn *sql.Conn)[]Appointment{
    var planningList []Appointment

    appointmentRow, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, COALESCE(u.phone, 'N/A'), TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth à HH24:MI'), 
	et.id, (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    (SELECT array_to_string(array_agg(s.name), ' - ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
	(SELECT array_to_string(array_agg(s.id), ',') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), a.status 
	FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id LEFT JOIN users AS u ON u.id=a.user_id 
    WHERE a.employee_id=$1 AND TO_CHAR(LOWER(a.date), 'YYYY-MM-DD') = $2 ORDER BY LOWER(a.date) ASC `, a.EmployeeId, a.Date)
    if err != nil{
        log.Printf("error in the query %s", err)
        return planningList
    }
    for appointmentRow.Next(){
        if err := appointmentRow.Scan(&a.Id, &a.CustomerName, &a.Contact, &a.Timeframe, &a.EtablishmentId, &a.Price, &a.Service, &a.ServiceTook, &a.Status); err != nil{
            log.Printf("error scanning the columns: %s", err)
            continue
        }
        planningList = append(planningList, *a)
    }

    dateRow := conn.QueryRowContext(context.Background(), `SELECT TO_CHAR($1::DATE, 'TMDay DD TMMonth YYYY')`, a.Date)
    if err := dateRow.Scan(&a.Date); err != nil{
        log.Printf("error getting the date")
    }
    return planningList
}

func (a *Appointment) Create()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    var totalDuration int
    for _, v := range a.Services{
        if v.Id != ""{
            totalDuration += v.Duration
        }
    }
    tx, err := conn.BeginTx(context.Background(), &sql.TxOptions{})
    if err != nil{
        log.Printf("error creating transition: %s", err)
        tx.Rollback()
        return errors.New("error tx")
    }

    appointmentRow := tx.QueryRow(`INSERT INTO appointment("date", status, user_id, etablishment_id, employee_id) VALUES(TSRANGE($1::TIMESTAMP, $1::TIMESTAMP + $2::INTERVAL),$3,$4,$5,$6) RETURNING id`, a.Date, fmt.Sprintf("%d minute", totalDuration), "Confirmé", a.UserId, a.EtablishmentId, a.EmployeeId)
    if err = appointmentRow.Scan(&a.Id); err != nil{
        log.Printf("error scanning appointment: %s", err)
        tx.Rollback()
        return errors.New("error inserting to appointment")
    }

    for _,v := range a.Services{
        if v.Id == "" {continue}
        result, err := tx.Exec(`INSERT INTO appointment_service (appointment_id, service_id) VALUES($1,$2)`, a.Id, v.Id)
        if err != nil{
            log.Printf("error inserting in appointment_service: %s", err)
            tx.Rollback()
            return errors.New("error inserting to appointment_service")
        }
        if affected, err := result.RowsAffected(); err != nil || affected == 0{
            log.Printf("error happend row affected: %s", err)
            tx.Rollback()
            return errors.New("error row affected")
        }
    }
    if err = tx.Commit(); err != nil{
        log.Printf("error commiting the transction")
    }
    return nil
}

func (a *Appointment) UserNextAppointment(conn *sql.Conn) error{
    appointmentRow:= conn.QueryRowContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, c.name, 
    TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth à HH24:MI'), (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT array_to_string(array_agg(s.name), ' - ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id) 
    FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id LEFT JOIN employee AS e ON e.id=a.employee_id LEFT JOIN users AS u ON u.id=e.user_id
    LEFT JOIN category AS c ON c.id=a.etablishment_id WHERE a.user_id=$1 AND LOWER(a.date) > NOW() AND a.status = 'Confirmé' ORDER BY LOWER(a.date) ASC LIMIT 1 `, a.UserId)
    if err := appointmentRow.Scan(&a.Id, &a.EmployeeName, &a.Adresse, &a.Category, &a.Date, &a.Price, &a.Service); err != nil{
        log.Printf("error scanning the columns: %s", err)
        return errors.New("error appointment scanning")
    }
    return nil
}

func (a *Appointment) UserAllAppointment(conn *sql.Conn)(appointmentList []Appointment){
    appointmentRow, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, et.id,
    TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY - HH24:MI'), (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT array_to_string(array_agg(s.name), ', ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    (SELECT array_to_string(array_agg(s.id), ',') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    a.status FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id 
    LEFT JOIN employee AS e ON e.id=a.employee_id LEFT JOIN users AS u ON u.id=e.user_id WHERE a.user_id=$1 ORDER BY LOWER(a.date) DESC LIMIT 5 `, a.UserId)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return
    }
    for appointmentRow.Next(){
        if err := appointmentRow.Scan(&a.Id, &a.EmployeeName, &a.Adresse, &a.EtablishmentId, &a.Date, &a.Price, &a.Service, &a.ServiceTook, &a.Status); err != nil{
            continue
        }
        appointmentList = append(appointmentList, *a)
    }
    
    return 
}

func (a *Appointment) UserAllForegoingAppointment(conn *sql.Conn)(appointmentList []Appointment){
    appointmentRow, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, et.id,
    TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY - HH24:MI'), (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT array_to_string(array_agg(s.name), ', ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    (SELECT array_to_string(array_agg(s.id), ',') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    a.status FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id LEFT JOIN employee AS e ON e.id=a.employee_id 
	LEFT JOIN users AS u ON u.id=e.user_id WHERE a.user_id=$1 AND a.status = 'Terminé' ORDER BY LOWER(a.date) DESC LIMIT 5 `, a.UserId)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return
    }
    for appointmentRow.Next(){
        if err := appointmentRow.Scan(&a.Id, &a.EmployeeName, &a.Adresse, &a.EtablishmentId, &a.Date, &a.Price, &a.Service, &a.ServiceTook, &a.Status); err != nil{
            continue
        }
        appointmentList = append(appointmentList, *a)
    }
    return 
}

func (a *Appointment) UserAllCancelledAppointment(conn *sql.Conn)(appointmentList []Appointment){
    appointmentRow, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, et.id,
    TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY - HH24:MI'), (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT array_to_string(array_agg(s.name), ', ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    a.status FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id 
    LEFT JOIN employee AS e ON e.id=a.employee_id LEFT JOIN users AS u ON u.id=e.user_id WHERE a.user_id=$1 AND a.status = 'Annulé' ORDER BY LOWER(a.date) DESC LIMIT 5 `, a.UserId)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return
    }
    for appointmentRow.Next(){
        if err := appointmentRow.Scan(&a.Id, &a.EmployeeName, &a.Adresse, &a.EtablishmentId, &a.Date, &a.Price, &a.Service, &a.Status); err != nil{
            continue
        }
        appointmentList = append(appointmentList, *a)
    }
    return 
}

func (a *Appointment) UserAllUpcommingAppointment(conn *sql.Conn)(appointmentList []Appointment){
    appointmentRow, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, et.id,
    TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY - HH24:MI'), (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT array_to_string(array_agg(s.name), ', ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    a.status FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id 
    LEFT JOIN employee AS e ON e.id=a.employee_id LEFT JOIN users AS u ON u.id=e.user_id WHERE a.user_id=$1 AND a.status = 'Confirmé' ORDER BY LOWER(a.date) ASC LIMIT 5 `, a.UserId)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return
    }
    for appointmentRow.Next(){
        if err := appointmentRow.Scan(&a.Id, &a.EmployeeName, &a.Adresse, &a.EtablishmentId, &a.Date, &a.Price, &a.Service, &a.Status); err != nil{
            continue
        }
        appointmentList = append(appointmentList, *a)
    }
    return 
}

func (a *Appointment) GetFull (conn *sql.Conn)(customer User, allEmployee []Employe, allService []Service, availebleDates []string,  err error){
    var serviceTaken []string
    //TODO: Marqué tout les services déjà pris
    appointmentRow := conn.QueryRowContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, eu.firstname || ' ' || eu.lastname, COALESCE(u.phone, 'N/A'), u.email, 
    LOWER(a.date), TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY'), (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT array_to_string(array_agg(s.name), ' - ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    (SELECT array_agg(s.id) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    a.user_id, a.employee_id, a.etablishment_id, a.status FROM appointment AS a LEFT JOIN users AS u ON u.id=a.user_id 
    LEFT JOIN employee AS e ON e.id=a.employee_id LEFT JOIN users AS eu ON eu.id=e.user_id WHERE a.id=$1 AND a.user_id=$2 OR a.id=$1 AND a.employee_id=$3`, a.Id, a.UserId, a.EmployeeId)

    if err := appointmentRow.Scan(&a.Id, &a.CustomerName, &a.EmployeeName, &customer.Phone, &customer.Email, &a.Date, &a.FormatDate, &a.Price, &a.Service, pq.Array(&serviceTaken), 
    &a.UserId, &a.EmployeeId, &a.EtablishmentId, &a.Status); err != nil{
        log.Printf("error scanning the row Appointment: %s", err)
        return customer, allEmployee, allService, availebleDates, errors.New("error getting the appointment")
    }

    service := Service{EtablishmentId: a.EtablishmentId}
    allService, err = service.GetList(conn)
    for i, s := range allService{
        for _, st := range serviceTaken{
			id, _ := strconv.Atoi(st)
            if s.Id == id{
                allService[i].Checked = true
                break
            }
        }
    }
    employe := Employe{EtablishmentId: a.EtablishmentId}
    allEmployee = employe.GetEtablishmentEmployees(conn)
    availebleDates = a.AvaileblesDates(conn)
    if err != nil{
        log.Printf("error getting the services: %s", err)
        return customer, allEmployee, allService, availebleDates, errors.New("error getting all services")
    }
    return customer, allEmployee, allService, availebleDates, nil
}

func (a *Appointment) Delete()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    result, err := conn.ExecContext(context.Background(), `UPDATE appointment SET status='Annulé' WHERE id=$2 AND user_id=$1`, a.UserId, a.Id)
    if err != nil{
        log.Printf("error in the query delete appointment: %s", err)
        return errors.New("error deleteting appointment")
    }
    affected, err := result.RowsAffected()
    if err != nil || affected == 0{
        return errors.New("error row introuvable")
    }
    return nil
}

func (a *Appointment) EtablishmentUpcomingAppointments(conn *sql.Conn)[]Appointment{

    var list []Appointment
    aList, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, c.name,
    TO_CHAR(LOWER(a.date), 'TMDay DD à HH24:MI'), (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT array_to_string(array_agg(s.name), ' - ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id)
    FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id LEFT JOIN category AS c ON c.id=et.category_id 
    LEFT JOIN users AS u ON u.id=a.user_id WHERE a.etablishment_id=$1 AND a.status = 'Confirmé' ORDER BY LOWER(a.date) ASC`, a.EtablishmentId)

    if err != nil{
        log.Printf("Error in the query: %s", err)
        return list
    }
    for aList.Next(){
        if err = aList.Scan(&a.Id, &a.CustomerName, &a.Adresse, &a.Category, &a.Date, &a.Price, &a.Service); err != nil{
            log.Printf("error scanning the columns: %s", err)
        }
        list = append(list, *a)
    }
    return list
}

func (a *Appointment) EtablishmentForegoingAppointments(conn *sql.Conn)[]Appointment{
    var list []Appointment

    aList, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, c.name,
    TO_CHAR(LOWER(a.date), 'TMDay DD à HH24:MI'), (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT array_to_string(array_agg(s.name), ' - ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id)
    FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id LEFT JOIN category AS c ON c.id=et.category_id
    LEFT JOIN users AS u ON u.id=a.user_id WHERE a.etablishment_id=$1 AND a.status = 'Terminé'  ORDER BY LOWER(a.date) DESC`, a.EtablishmentId)
    if err != nil{
        log.Printf("Error in the query: %s", err)
        return list
    }
    for aList.Next(){
        if err = aList.Scan(&a.Id, &a.CustomerName, &a.Adresse, &a.Category, &a.Date, &a.Service, &a.Price); err != nil{
            log.Printf("error scanning the columns: %s", err)
        }
        list = append(list, *a)
    }
    return list

}
