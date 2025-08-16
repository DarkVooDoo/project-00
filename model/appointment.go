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
	Price float64 `json:"price,string"`
}

type CardPosition struct{
	Height float64
	Position float64
}

type Appointment struct{
    Id string
    EmployeeName string
    CustomerName string
    Adresse string
    Category string
    Services []ServicePayload
	ServiceList []string
    Service string
	ServiceTook string
    Price string
    Contact string
    Status string
	Position CardPosition
    Description string
    UserId int64
    Date string `json:"date"`
    Timeframe string
    FormatDate string
    EmployeeId int64 `json:"employee,string"`
    EtablishmentId int64
}

type DayAppointmentRecap struct{
	Total int
	Accepted int
	Cancelled int
	Finish int
}

type EmployeeWithAppointment struct{
	Employee Employe
	Appointment []Appointment
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

func (a *Appointment) EmployeeAppointmentDayInNumbers(conn *sql.Conn) DayAppointmentRecap{
	var dayRecap DayAppointmentRecap
	dayNumbersRow := conn.QueryRowContext(context.Background(), `
		SELECT 
		(SELECT COUNT(a.id) FROM appointment AS a WHERE LOWER(a.date)::DATE = $2 AND a.employee_id=$1 OR a.etablishment_id=$3 AND LOWER(a.date)::DATE = $2),
		(SELECT COUNT(a.id) FROM appointment AS a WHERE LOWER(a.date)::DATE = $2 AND a.employee_id=$1 AND a.status='Confirmé' OR a.etablishment_id=$3 AND a.status='Confirmé' AND LOWER(a.date)::DATE = $2),
		(SELECT COUNT(a.id) FROM appointment AS a WHERE LOWER(a.date)::DATE = $2 AND a.employee_id=$1 AND a.status='Terminé' OR a.etablishment_id=$3 AND a.status='Terminé' AND LOWER(a.date)::DATE = $2),
		(SELECT COUNT(a.id) FROM appointment AS a WHERE LOWER(a.date)::DATE = $2 AND a.employee_id=$1 AND a.status='Annulé' OR a.etablishment_id=$3 AND a.status='Annulé' AND LOWER(a.date)::DATE = $2)`, a.EmployeeId, a.Date, a.EtablishmentId)
	if err := dayNumbersRow.Scan(&dayRecap.Total, &dayRecap.Accepted, &dayRecap.Finish, &dayRecap.Cancelled); err != nil{
		log.Printf("error scanning the numbers for appointment: %s", err)
		return dayRecap
	}
	return dayRecap
}

func (a *Appointment) EmployeePlanning(conn *sql.Conn)(planningList []Appointment, shift []string){
	var shiftString string

	timestampRows, err := conn.QueryContext(context.Background(), `SELECT TO_CHAR(generate_series(CONCAT('2025-05-05 ', open_time)::timestamp, 
	CONCAT('2025-05-05 ', close_time)::timestamp, INTERVAL '1 hour')::TIMESTAMP, 'HH24:MI') FROM employee AS e 	
	LEFT JOIN schedule AS s ON e.etablishment_id=s.etablishment_id AND s.day = EXTRACT(ISODOW FROM $1::DATE) - 1 WHERE e.id=$2`, a.Date, a.EmployeeId)
	if err != nil{
		log.Printf("error getting timestamp for planning: %s", err)
		return
	}
	for timestampRows.Next(){
		if err = timestampRows.Scan(&shiftString); err != nil{
			log.Printf("error scanning the shift for the planning: %s", err)
			continue
		}
		shift = append(shift, shiftString)
	}

    appointmentRow, err := conn.QueryContext(context.Background(), `
	WITH appointment_cte AS(
  		SELECT a.id AS a_id, u.firstname || ' ' || u.lastname AS fullname, COALESCE(u.phone, 'N/A') AS phone, TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth à HH24:MI') AS fulldate,
  		  et.id AS et_id, (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id) AS total, 
  		  (SELECT array_to_string(array_agg(s.name), ' - ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id) AS service,
  		  (SELECT array_to_string(array_agg(s.id), ',') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id) AS s_ids, a.status,
		  LOWER(a.date) AS startAppointment, UPPER(a.date) AS endAppointment
  		FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id LEFT JOIN users AS u ON u.id=a.user_id 
  		WHERE a.employee_id=$1 AND TO_CHAR(LOWER(a.date), 'YYYY-MM-DD') = $2 AND a.status='Confirmé' ORDER BY LOWER(a.date) ASC 
		), calc_minute AS(
		  	SELECT *,
		  	TO_CHAR(startAppointment, 'HH24')::INT * 60 + TO_CHAR(startAppointment, 'MI')::INT AS startMinute, 
		  	TO_CHAR(endAppointment, 'HH24')::INT * 60 + TO_CHAR(endAppointment, 'MI')::INT AS endMinute,
			TO_CHAR(CONCAT('2025-05-05 ',$3::TEXT)::TIMESTAMP, 'HH24')::INT * 60 + TO_CHAR(CONCAT('2025-05-05 ', $3::TEXT)::TIMESTAMP, 'MI')::INT AS scheduleBase FROM appointment_cte
		), pos AS(
		  SELECT *,
		  ((startMinute - scheduleBase) / 60 ) * 100 AS pos, 
		  ((endMinute - startMinute) / 60) * 100 - 5 AS height
		  FROM calc_minute
		)
		SELECT a_id, fullname, phone, fulldate, et_id, total, service, s_ids, status, pos, height FROM pos`, a.EmployeeId, a.Date, shift[0])
    if err != nil{
        log.Printf("error in the query %s", err)
        return
    }
    for appointmentRow.Next(){
        if err := appointmentRow.Scan(&a.Id, &a.CustomerName, &a.Contact, &a.Timeframe, &a.EtablishmentId, &a.Price, &a.Service, &a.ServiceTook, &a.Status, 
		&a.Position.Position, &a.Position.Height); err != nil{
            log.Printf("error scanning the columns: %s", err)
            continue
        }
        planningList = append(planningList, *a)
    }

    dateRow := conn.QueryRowContext(context.Background(), `SELECT TO_CHAR($1::DATE, 'TMDay DD TMMonth YYYY')`, a.Date)
    if err := dateRow.Scan(&a.Date); err != nil{
        log.Printf("error getting the date")
    }
    return
}

func (a *Appointment) Create()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    var totalDuration int
	var totalPrice float64
    for _, v := range a.Services{
        if v.Id != ""{
            totalDuration += v.Duration
			totalPrice += v.Price
        }
    }
    tx, err := conn.BeginTx(context.Background(), &sql.TxOptions{})
    if err != nil{
        log.Printf("error creating transition: %s", err)
        tx.Rollback()
        return errors.New("error tx")
    }

    appointmentRow := tx.QueryRow(`INSERT INTO appointment("date", total, status, user_id, etablishment_id, employee_id) VALUES(TSRANGE($1::TIMESTAMP, $1::TIMESTAMP + $2::INTERVAL),$3,$4,$5,$6,$7) RETURNING id`, a.Date, fmt.Sprintf("%d minute", totalDuration), totalPrice, "Confirmé", a.UserId, a.EtablishmentId, a.EmployeeId)
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
	//review := Review{UserId: a.UserId, EtablishmentId: a.EtablishmentId, EmployeeId: a.EmployeeId}
	//if err := review.Create(conn); err != nil{
	//	log.Printf("error creating the pending review: %s", err)
	//}
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
    TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY - HH24:MI'), a.total,
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
    TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY - HH24:MI'), a.total,
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
    TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY - HH24:MI'), a.total,
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
    a.status FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id LEFT JOIN employee AS e ON e.id=a.employee_id LEFT JOIN users AS u ON u.id=e.user_id 
	WHERE a.user_id=$1 AND a.status = 'Confirmé' AND LOWER(a.date) > NOW() ORDER BY LOWER(a.date) ASC LIMIT 5 `, a.UserId)
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

func (a *Appointment) Complete(conn *sql.Conn) error{
	result, err := conn.ExecContext(context.Background(), `UPDATE appointment SET status='Terminé' WHERE id=$1 AND employee_id=$2 OR id=$1 AND etablishment_id=$3`, a.Id, a.EmployeeId, a.EtablishmentId)
	if err != nil{
		log.Printf("error executing query: %s", err)
		return errors.New("error in the query")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0{
		log.Printf("error no row affected: %s", err)
		return errors.New("no rows affected")
	}
	return nil
}

func (a *Appointment) GetFull (conn *sql.Conn)(customer User, allEmployee []Employe, allService []Service, availebleDates []string,  err error){
    var serviceTaken []string
    appointmentRow := conn.QueryRowContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, eu.firstname || ' ' || eu.lastname, COALESCE(u.phone, 'N/A'), u.email, 
    LOWER(a.date), TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY'), a.total,
    (SELECT array_to_string(array_agg(s.name), ' - ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    (SELECT array_agg(s.id) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
	(SELECT array_to_string(array_agg(s.id), ',') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    a.user_id, a.employee_id, a.etablishment_id, a.status FROM appointment AS a LEFT JOIN users AS u ON u.id=a.user_id 
    LEFT JOIN employee AS e ON e.id=a.employee_id LEFT JOIN users AS eu ON eu.id=e.user_id WHERE a.id=$1 AND a.user_id=$2 OR a.id=$1 AND a.employee_id=$3`, a.Id, a.UserId, a.EmployeeId)

    if err := appointmentRow.Scan(&a.Id, &a.CustomerName, &a.EmployeeName, &customer.Phone, &customer.Email, &a.Date, &a.FormatDate, &a.Price, &a.Service, pq.Array(&serviceTaken), 
    &a.ServiceTook, &a.UserId, &a.EmployeeId, &a.EtablishmentId, &a.Status); err != nil{
        log.Printf("error scanning the row Appointment: %s", err)
        return customer, allEmployee, allService, availebleDates, errors.New("error getting the appointment")
    }

    service := Service{EtablishmentId: a.EtablishmentId}
    allService, err = service.GetList(conn)
    for i, s := range allService{
        for _, st := range serviceTaken{
			id, _ := strconv.Atoi(st)
            if s.Id == int64(id){
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

    result, err := conn.ExecContext(context.Background(), `UPDATE appointment SET status='Annulé' WHERE id=$2 AND user_id=$1 OR id=$2 AND employee_id=$3 OR etablishment_id=$4 AND id=$2`, 
	a.UserId, a.Id, a.EmployeeId, a.EtablishmentId)
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

func (a *Appointment) EtablishmentTodayAppointments(conn *sql.Conn)[]Appointment{

    var list []Appointment
    aList, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, c.name,
    TO_CHAR(LOWER(a.date), 'HH24:MI'), (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT array_to_string(array_agg(s.name), ' - ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id)
    FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id LEFT JOIN category AS c ON c.id=et.category_id 
	LEFT JOIN users AS u ON u.id=a.user_id WHERE a.etablishment_id=$1 AND LOWER(a.date)::DATE = NOW()::DATE ORDER BY LOWER(a.date) ASC`, a.EtablishmentId)

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

func (a Appointment) EtablishmentAppointments(conn *sql.Conn)(ewa []EmployeeWithAppointment){
	employee := Employe{EtablishmentId: a.EtablishmentId}

	date := a.Date
	var appointmentList []Appointment
	employeeList := employee.AppointmentEmployee(conn, date, a.Status)

	for _, v := range employeeList{
		appointmentList = []Appointment{}
    	aList, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, c.name,
    	TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth à HH24:MI'), a.total, a.status,
    	(SELECT array_agg(s.name) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id)
    	FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id LEFT JOIN category AS c ON c.id=et.category_id
		LEFT JOIN users AS u ON u.id=a.user_id WHERE a.etablishment_id=$1 AND LOWER(a.date)::DATE = $2 AND
		CASE $3
			WHEN 'Confirmé' THEN a.status = 'Confirmé'
			WHEN 'Terminé' THEN a.status = 'Terminé'
			WHEN 'Annulé' THEN a.status = 'Annulé'
			ELSE a.status IS NOT NULL
		END 
		AND a.employee_id=$4 ORDER BY LOWER(a.date) DESC`, a.EtablishmentId, date, a.Status, v.Id)
    	if err != nil{
    	    log.Printf("Error in the query: %s", err)
    	    return
    	}
    	for aList.Next(){
    	    if err = aList.Scan(&a.Id, &a.CustomerName, &a.Adresse, &a.Category, &a.Date, &a.Price, &a.Status, pq.Array(&a.ServiceList)); err != nil{
    	        log.Printf("error scanning the columns: %s", err)
    	    }
    	    appointmentList = append(appointmentList, a)
    	}
		ewa = append(ewa, EmployeeWithAppointment{
			v,
			appointmentList,
		})
	}

    return
}
