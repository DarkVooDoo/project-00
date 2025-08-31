package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/lib/pq"
)

type ServicePayload struct{
    Id string `json:"id"`
	Name string `json:"name"`
    Duration int `json:"duration,string"`
	Price float64 `json:"price,string"`
	CurrencyPrice string `json:"currencyPrice"`
}

type CardPosition struct{
	Height float64
	Position float64
}

type Appointment struct{
	Id string `json:"id"`
    EmployeeName string
	CustomerName string `json:"customerName"`
    Adresse string
    Category string
    Services []ServicePayload `json:"serviceList"`
	ServiceList []string 
    Service string
	ServiceTook string
	Price string `json:"price"`
	Contact string `json:"contact"`
	Status string `json:"status"`
	Position CardPosition
    Description string
    UserId int64
    Date string `json:"date"`
	Time string `json:"time"`
	Timeframe string `json:"timeframe"`
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
	var serviceList []ServicePayload

	timestampRows, err := conn.QueryContext(context.Background(), `SELECT TO_CHAR(generate_series(CONCAT('2025-05-05 ', open_time)::timestamp, 
	CONCAT('2025-05-05 ', close_time)::timestamp, INTERVAL '1 hour')::TIMESTAMP, 'HH24:MI') FROM employee AS e LEFT JOIN schedule AS s ON e.etablishment_id=s.etablishment_id 
	AND s.day = EXTRACT(ISODOW FROM $1::DATE) - 1 WHERE e.id=$2 AND open_time IS NOT NULL`, a.Date, a.EmployeeId)
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

	if len(shift) == 0{return}
    appointmentRow, err := conn.QueryContext(context.Background(), `
	WITH appointment_cte AS(
  		SELECT a.id AS a_id, COALESCE(a.name, u.firstname || ' ' || u.lastname) AS fullname, COALESCE(a.phone, u.phone, 'N/A') AS phone, TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth à HH24:MI') AS fulldate,
  		  et.id AS et_id, (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id) AS total, 
		  a.status, LOWER(a.date) AS startAppointment, UPPER(a.date) AS endAppointment
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
		SELECT a_id, fullname, phone, fulldate, et_id, total, status, pos, height FROM pos`, a.EmployeeId, a.Date, shift[0])
    if err != nil{
        log.Printf("error in the query %s", err)
        return
    }
    for appointmentRow.Next(){
        if err := appointmentRow.Scan(&a.Id, &a.CustomerName, &a.Contact, &a.Timeframe, &a.EtablishmentId, &a.Price, &a.Status, 
		&a.Position.Position, &a.Position.Height); err != nil{
            log.Printf("error scanning the columns: %s", err)
            continue
        }
        planningList = append(planningList, *a)
    }

	for i, v := range planningList{
		var service ServicePayload
		servicesRows, err := conn.QueryContext(context.Background(), `SELECT s.name, s.price FROM appointment_service AS app_s LEFT JOIN service AS s ON s.id=app_s.service_id
		WHERE app_s.appointment_id=$1`, v.Id)
		if err != nil{
			log.Printf("error in the query of the sevices: %s", err)
		}

		for servicesRows.Next(){
			if err := servicesRows.Scan(&service.Name, &service.CurrencyPrice); err != nil{
				log.Printf("error scanning the services: %s", err)
				continue
			}
			serviceList = append(serviceList, service)
		}
		serviceBytes, _ := json.Marshal(serviceList)
		planningList[i].ServiceTook = string(serviceBytes)
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

	if a.CustomerName != ""{
		appointmentRow := tx.QueryRow(`INSERT INTO appointment("date", status, name, phone, etablishment_id, employee_id) 
		VALUES(TSRANGE($1::TIMESTAMP, $1::TIMESTAMP + $2::INTERVAL),$3,$4,$5,$6,$7,$8) RETURNING id`, 
		a.Date, fmt.Sprintf("%d minute", totalDuration), "Confirmé", a.CustomerName, a.Contact, a.EtablishmentId, a.EmployeeId)
		if err = appointmentRow.Scan(&a.Id); err != nil{
			log.Printf("error scanning appointment: %s", err)
			tx.Rollback()
			return errors.New("error inserting to appointment")
		}
	}else{
		appointmentRow := tx.QueryRow(`INSERT INTO appointment("date", status, user_id, etablishment_id, employee_id) VALUES(TSRANGE($1::TIMESTAMP, $1::TIMESTAMP + $2::INTERVAL),$3,$4,$5,$6,$7) RETURNING id`, a.Date, fmt.Sprintf("%d minute", totalDuration), "Confirmé", a.UserId, a.EtablishmentId, a.EmployeeId)
		if err = appointmentRow.Scan(&a.Id); err != nil{
			log.Printf("error scanning appointment: %s", err)
			tx.Rollback()
			return errors.New("error inserting to appointment")
		}
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

func (a Appointment) UpdateAppointment(conn *sql.Conn, employeeId int64)error{
    var totalDuration int
    for _, v := range a.Services{
        if v.Id != ""{
            totalDuration += v.Duration
        }
    }
	tx, _ := conn.BeginTx(context.Background(), &sql.TxOptions{})
	a.Date = fmt.Sprintf("%s %s", a.Date, a.Time)
	result, err := tx.Exec(`UPDATE appointment AS a SET employee_id=$4, date=TSRANGE($5::TIMESTAMP, $5::TIMESTAMP + $6::INTERVAL)
	WHERE a.id=$1 AND a.etablishment_id=$2 OR a.id=$1 AND a.user_id=$3 OR a.id=$1 AND a.employee_id=$7`, 
	a.Id, a.EtablishmentId, a.UserId, a.EmployeeId, a.Date, fmt.Sprintf("%d minute", totalDuration), employeeId)
	if err != nil{
		log.Printf("error in the update query: %s", err)
		tx.Rollback()
		return errors.New("error in the query")
	}
	if rowAffected, _ := result.RowsAffected(); rowAffected == 0{
		log.Printf("error update zero rows affected: %d", rowAffected)
		tx.Rollback()
		return errors.New("error zero row affected")
	}

	tx.Exec(`DELETE FROM appointment_service WHERE appointment_id=$1`, a.Id)
	for _, v := range a.Services{
		if v.Id == "" {continue}
		if _, err := tx.Exec(`INSERT INTO appointment_service(service_id, appointment_id) VALUES($1,$2)`, v.Id, a.Id); err != nil{
			log.Printf("error inserting to appointment_service: %s", err)
			tx.Rollback()
			return errors.New("error inserting to appointment_service table")
		}
	}
	tx.Commit()
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

func (a *Appointment) UserAppointment(conn *sql.Conn, page int)(appointmentList []Appointment, pageCount int){
	limit := 5
	pageCountRow := conn.QueryRowContext(context.Background(), `
	SELECT CEIL(COUNT(*)::FLOAT / $3) FROM appointment AS a WHERE a.user_id=$1 AND 
	CASE $2 
		WHEN 'Confirmé' THEN a.status = 'Confirmé' 
		WHEN 'Terminé' THEN a.status = 'Terminé'
		WHEN 'Annulé' THEN a.status = 'Annulé'
		ELSE a.status IS NOT NULL 
	END`, a.UserId, a.Status, limit)

	if err := pageCountRow.Scan(&pageCount); err != nil{
		log.Printf("error scanning the page count: %s", err)
	}

    appointmentRow, err := conn.QueryContext(context.Background(), `SELECT a.id, u.firstname || ' ' || u.lastname, et.adresse || ', ' || et.postal, et.id,
    TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY - HH24:MI'), a.status,
    (SELECT array_to_string(array_agg(s.name), ', ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
    (SELECT array_to_string(array_agg(s.id), ',') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id),
    (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id) 
    FROM appointment AS a LEFT JOIN etablishment AS et ON et.id=a.etablishment_id 
    LEFT JOIN employee AS e ON e.id=a.employee_id LEFT JOIN users AS u ON u.id=e.user_id WHERE a.user_id=$1 AND 
	CASE $2 
		WHEN 'Confirmé' THEN a.status = 'Confirmé' 
		WHEN 'Terminé' THEN a.status = 'Terminé'
		WHEN 'Annulé' THEN a.status = 'Annulé'
		ELSE a.status IS NOT NULL END
	ORDER BY LOWER(a.date) DESC LIMIT $4 OFFSET $3 `, a.UserId, a.Status, page * 2 - 2, limit)

    if err != nil{
        log.Printf("error in the query: %s", err)
        return
    }
    for appointmentRow.Next(){
        if err := appointmentRow.Scan(&a.Id, &a.EmployeeName, &a.Adresse, &a.EtablishmentId, &a.Date, &a.Status, &a.Service, &a.ServiceTook, &a.Price); err != nil{
            continue
        }
        appointmentList = append(appointmentList, *a)
    }
    return 
}

func (a *Appointment) UpdateStatus(conn *sql.Conn) error{
	result, err := conn.ExecContext(context.Background(), `UPDATE appointment SET status=CASE $4 
	WHEN 'Termine' THEN 'Terminé'
	WHEN 'Annulé' THEN 'Annulé' END::appointment_status WHERE id=$1 AND employee_id=$2 OR id=$1 AND etablishment_id=$3`, a.Id, a.EmployeeId, a.EtablishmentId, a.Status)
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
	query := `
		WITH appointment_cte AS(
			SELECT a.id AS a_id, COALESCE(a.name, u.firstname || ' ' || u.lastname), COALESCE(a.phone, u.phone, 'N/A'), COALESCE(u.email, 'N/A'), LOWER(a.date), 
			TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth YYYY') AS formatDate, (SELECT SUM(s.price) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a.id), 
			COALESCE(a.user_id, 0), a.employee_id, a.etablishment_id, a.status, COALESCE(employee_user.firstname || ' ' || employee_user.lastname) AS employee_name
			FROM appointment AS a LEFT JOIN users AS u ON u.id=a.user_id LEFT JOIN employee AS em ON em.id=a.employee_id LEFT JOIN users AS employee_user ON employee_user.id=em.user_id 
			WHERE a.id=$1 AND a.user_id=$2 OR a.id=$1 AND a.employee_id=$3
		), service_cte AS(
			SELECT *, 
			(SELECT array_to_string(array_agg(s.name), ' - ') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a_id),
			(SELECT array_agg(s.id) FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a_id),
			(SELECT array_to_string(array_agg(s.id), ',') FROM appointment_service AS az LEFT JOIN service AS s ON s.id=az.service_id WHERE az.appointment_id=a_id) FROM appointment_cte
			
		)
		SELECT * FROM service_cte
	`
    appointmentRow := conn.QueryRowContext(context.Background(), query, a.Id, a.UserId, a.EmployeeId)

    if err := appointmentRow.Scan(&a.Id, &a.CustomerName, &customer.Phone, &customer.Email, &a.Date, &a.FormatDate, &a.Price, &a.UserId, &a.EmployeeId, &a.EtablishmentId, &a.Status, &a.EmployeeName,
	&a.Service, pq.Array(&serviceTaken), &a.ServiceTook); err != nil{
        log.Printf("error scanning the row Appointment: %s", err)
        return customer, allEmployee, allService, availebleDates, errors.New("error getting the appointment")
    }

    service := Service{EtablishmentId: a.EtablishmentId}
	allService, _ = service.GetList(conn)
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
    return customer, allEmployee, allService, availebleDates, nil
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
    	aList, err := conn.QueryContext(context.Background(), `SELECT a.id, COALESCE(a.name, u.firstname || ' ' || u.lastname), et.adresse || ', ' || et.postal, c.name,
    	TO_CHAR(LOWER(a.date), 'TMDay DD TMMonth à HH24:MI'), a.status,
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
    	    if err = aList.Scan(&a.Id, &a.CustomerName, &a.Adresse, &a.Category, &a.Date, &a.Status, pq.Array(&a.ServiceList)); err != nil{
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
