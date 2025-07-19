package model

import (
	"context"
	"database/sql"
	"errors"
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

type KeyValue struct{
    Id int
    Value string
}

type Etablishment struct{
    Id int64
    Name string `json:"name"`
    Adresse string `json:"adresse"`
    Postal string `json:"postal"`
    Phone string `json:"phone"`
	Lat float64 `json:"lat,string"`
	Lon float64 `json:"lon,string"`
    Payment []string `json:"payment"`
    AllPayment []string 
    Employee []Employe
    Service []Service
    Schedule []DaySchedule
	Review []Review
    TodaySchedule string
    IsOpen string
	NextShift string
    Category string `json:"category"`
    UserId int64
}

var Week = []string{"Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi", "Dimanche"}

func (e *Etablishment) Create()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    etablishmentRow:= conn.QueryRowContext(context.Background(), `INSERT INTO etablishment(name, adresse, postal, geolocation, phone, payment, category_id, user_id) 
	VALUES($1,$2,$3,POINT($4, $5),$6,$7,$8,$9) RETURNING id`, e.Name, e.Adresse, e.Postal, e.Lat, e.Lon, e.Phone, pq.Array(e.Payment), e.Category, e.UserId)
    if err := etablishmentRow.Scan(&e.Id); err != nil{
        log.Printf("error scanning the query: %s", err)
        return errors.New("error nothing happend")
    }
    return nil
}

func(e *Etablishment) Latest(conn *sql.Conn)[]Etablishment{
    var list []Etablishment
	rows, err := conn.QueryContext(context.Background(), `
		WITH derniers_etabs AS (
		    SELECT * FROM etablishment
		),
		maintenant AS (
		    SELECT EXTRACT(ISODOW FROM NOW()) - 1 AS jour_semaine, NOW()::time AS heure_actuelle
		),
		horaires AS (
		    SELECT 
		        e.id,
		        e.name,
		        e.adresse,
				e.postal,
				e.phone,
				c.name AS category,
		        oh.day,
		        oh.open_time,
		        oh.close_time,
				-- détermine si l'établissement est ouvert maintenant
		        CASE 
		            WHEN oh.day = EXTRACT(ISODOW FROM NOW()) - 1
		                 AND NOW()::time BETWEEN oh.open_time AND oh.close_time 
		            THEN TRUE 
		            ELSE FALSE 
		        END AS est_ouvert,
		        -- calcule le prochain horaire d'ouverture après maintenant
		        ((oh.day - EXTRACT(DOW FROM NOW()) + 7) % 7) AS jours_d_attente
		    FROM derniers_etabs e
		    LEFT JOIN schedule oh ON e.id = oh.etablishment_id
			LEFT JOIN category AS c ON c.id = e.category_id
		),
		prochains_horaires AS (
		    SELECT *,
		        CASE 
					WHEN est_ouvert THEN CONCAT('Ferme à ', TO_CHAR(close_time, 'HH24:MI'))
					WHEN open_time IS NULL THEN 'Fermé Termporairement'
					ELSE CONCAT('Ouvre ', ('{Lun,Mar,Mer,Jeu,Ven,Sam,Dim}'::TEXT[])[day + 1], ' ' , TO_CHAR(open_time, 'HH24:MI'))
				END AS prochain_horaire,
		        ROW_NUMBER() OVER (PARTITION BY id ORDER BY jours_d_attente) AS rn
		    FROM horaires
		)
		SELECT id, name, adresse, postal, COALESCE(phone, ''), category, CASE WHEN est_ouvert THEN 'Ouvert' ELSE 'Fermé' END, prochain_horaire FROM prochains_horaires
		WHERE rn = 1
		GROUP BY id, name, adresse, postal, phone, category, prochain_horaire, est_ouvert
	`)
    if err != nil{
        log.Printf("error getting etablishment: %s", err)
        return list
    }
    for rows.Next(){
        if err := rows.Scan(&e.Id, &e.Name, &e.Adresse, &e.Postal, &e.Phone, &e.Category, &e.IsOpen, &e.NextShift); err != nil{
            log.Printf("error scanning rows: %s", err)
        }
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

func (e *Etablishment) UpdateParametre(conn * sql.Conn)error{
	result, err := conn.ExecContext(context.Background(), `UPDATE etablishment SET name=$1, phone=$2, adresse=$3, postal=$4, payment=$5, category_id=$6 WHERE id=$7`, 
	e.Name, e.Phone, e.Adresse, e.Postal, pq.Array(e.Payment), e.Category, e.Id)
	if err != nil{
		log.Printf("error executing the query: %s", err)
		return errors.New("error in the query")
	}
	aff, err := result.RowsAffected()
	if err != nil || aff == 0{
		log.Printf("error zero rows affected: %d", aff)
		return errors.New("error nothing happend")
	}
	return nil
}

func (e *Etablishment) UserEtablishment(conn *sql.Conn)(error){
    etablishmentList:= conn.QueryRowContext(context.Background(), `SELECT id, name FROM etablishment WHERE user_id=$1 AND id=$2`, e.UserId, e.Id)
	if err := etablishmentList.Scan(&e.Id, &e.Name); err != nil{
		log.Printf("error scanning the etablishment: %s", err)
    }
    return nil
}

func (e *Etablishment) GetEmployeeAndService(conn *sql.Conn){
    var employee Employe
    var service Service
    var name  sql.NullString
	var id sql.NullInt64
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
        employee.Id = id.Int64
        e.Employee = append(e.Employee, employee)
    }

	serviceList, err := conn.QueryContext(context.Background(), `SELECT s.id, s.name, s.price, s.price::NUMERIC, s.duration FROM service AS s WHERE s.etablishment_id=$1`, e.Id)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return
    }
    for serviceList.Next(){
        if err = serviceList.Scan(&service.Id, &service.Name, &service.CurrencyPrice, &service.Price, &service.Duration); err != nil{
            log.Printf("error scan: %s", err)
        }
        e.Service = append(e.Service, service)
    }
}

func (e *Etablishment) Public(conn *sql.Conn)(int, error){
    var service Service
	var review Review
	schedule := DaySchedule{EtablishmentId: e.Id}
    var weekDay int
	//(SELECT AVG(rating) FROM review WHERE etablishment_id=$1)
    etablishmentRow, err := conn.QueryContext(context.Background(), `SELECT e.name, e.payment, COALESCE(e.phone, ''), e.adresse, e.postal, s.name, s.price, s.description, s.duration, 
    c.name, EXTRACT(ISODOW FROM NOW()) - 1 
    FROM etablishment AS e LEFT JOIN service AS s ON s.etablishment_id=e.id LEFT JOIN category AS c ON c.id=e.category_id WHERE e.id=$1`, e.Id)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return weekDay, errors.New("error in the query")
    }
    for etablishmentRow.Next(){
        if err := etablishmentRow.Scan(&e.Name, pq.Array(&e.Payment), &e.Phone, &e.Adresse, &e.Postal, &service.Name, &service.Price, &service.Description, 
        &service.Duration,  &e.Category, &weekDay); err != nil{
            log.Printf("error scanning the schedule: %s", err)
            continue
        }
        e.Service = append(e.Service, service)
    }

    if e.Name == ""{
        return weekDay, errors.New("error no data")
    }

	e.Schedule = schedule.GetSchedule(conn)

	reviewRows, err := conn.QueryContext(context.Background(), `SELECT r.id, r.comment, r.rating, 
	r.created_at, u.firstname || ' ' || u.lastname FROM review AS r LEFT JOIN users AS u ON u.id=r.user_id WHERE etablishment_id=$1 LIMIT 3`, e.Id)
	if err != nil{
		log.Printf("error in the query for reviews: %s", err)
	}
	for reviewRows.Next(){
		if err := reviewRows.Scan(&review.Id, &review.Comment, &review.Rating, &review.Date, &review.UserName); err != nil{
			log.Printf("error scanning the review: %s", err)
			continue
		}
		review.StarCount = make([]int, 5)
		review.Rating = review.Rating - 1
		e.Review = append(e.Review, review)
	}
    return weekDay, nil
}

func Payments(conn *sql.Conn)[]string{
	var paymentList []string
	paymentRow := conn.QueryRowContext(context.Background(), `SELECT enum_range(NULL::payment_type)`)
	if err := paymentRow.Scan(pq.Array(&paymentList)); err != nil{
		log.Printf("error scanning the row: %s", err)
		return paymentList
	}
	return paymentList
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
