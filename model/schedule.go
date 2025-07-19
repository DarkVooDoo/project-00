package model

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type DaySchedule struct{
	Id string `json:"id"`
	Day int `json:"day,string"`
	DayName string
	IsClose string
	OpenTime string `json:"open_time"`
	CloseTime string `json:"close_time"`
	EtablishmentId int64 
}

func (s DaySchedule) Create(conn *sql.Conn)error{
	
	scheduleRow := conn.QueryRowContext(context.Background(), `INSERT INTO schedule (day, open_time, close_time, etablishment_id) VALUES($1,$2,$3,$4) RETURNING id`, 
	s.Day, s.OpenTime, s.CloseTime, s.EtablishmentId)
	if err := scheduleRow.Scan(&s.Id); err != nil{
		log.Println("error scanning the inserted row: %s", err)
		return errors.New("error creating new schedule")
	}
	return nil
}

func (s DaySchedule) GetSchedule(conn *sql.Conn)[]DaySchedule{
    var schedule []DaySchedule
	//TODO: How to know if the etablishment is currently open if there is multiple shift in one day?
	scheduleRows, err := conn.QueryContext(context.Background(), `SELECT id, day, TO_CHAR(open_time, 'HH24:MI'), TO_CHAR(close_time, 'HH24:MI'),
	 CASE WHEN open_time > close_time AND EXTRACT(ISODOW FROM NOW()) - 1 = day THEN
  		CASE WHEN CONCAT(MAKE_DATE(2025,4,22), ' ', NOW()::TIME)::TIMESTAMP BETWEEN CONCAT(MAKE_DATE(2025,4,22),' ', open_time)::timestamp AND CONCAT(MAKE_DATE(2025,4,23),' ', close_time)::timestamp 
		THEN 'Actuellement ouvert' 
		WHEN CONCAT(MAKE_DATE(2025,4,22), ' ', NOW()::TIME)::TIMESTAMP 
    		NOT BETWEEN CONCAT(MAKE_DATE(2025,4,22), ' ', open_time)::timestamp AND CONCAT(MAKE_DATE(2025,4,23), ' ', close_time)::timestamp THEN 'Actuellement fermé'
    	END
  		WHEN EXTRACT(ISODOW FROM NOW()) - 1 = day AND CONCAT(MAKE_DATE(2025,4,22), ' ', NOW()::TIME)::TIMESTAMP 
		BETWEEN CONCAT(MAKE_DATE(2025,4,22), ' ', open_time)::TIMESTAMP AND CONCAT(MAKE_DATE(2025,4,22),' ', close_time)::TIMESTAMP THEN 'Actuellement ouvert' 
		WHEN EXTRACT(ISODOW FROM NOW()) - 1 = day AND CONCAT(MAKE_DATE(2025,4,22), ' ', NOW()::TIME)::TIMESTAMP 
		NOT BETWEEN CONCAT(MAKE_DATE(2025,4,22), ' ', open_time)::TIMESTAMP AND CONCAT(MAKE_DATE(2025,4,22), ' ', close_time)::TIMESTAMP THEN 'Actuellement fermé' 
		ELSE '' END FROM schedule WHERE etablishment_id=$1 ORDER by day`, s.EtablishmentId)
	if err != nil{
		log.Printf("error getting the schedule: %s", err)
		return schedule
	}
	for scheduleRows.Next(){
		if err := scheduleRows.Scan(&s.Id, &s.Day, &s.OpenTime, &s.CloseTime, &s.IsClose); err != nil{
			log.Printf("error scanning the table schedule: %s", err)
			continue
		}
		schedule = append(schedule, s)
	}
	return schedule
}

func (s *DaySchedule) Update()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    result, err := conn.ExecContext(context.Background(), `UPDATE schedule SET open_time=$1, close_time=$2 WHERE etablishment_id=$3 AND id=$4`, s.OpenTime, s.CloseTime, s.EtablishmentId, s.Id)
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

func(s DaySchedule) Delete()error{
	conn := GetDBPoolConn()
	defer conn.Close()

	result, err := conn.ExecContext(context.Background(), `DELETE FROM schedule WHERE id=$1 AND etablishment_id=$2`, s.Id, s.EtablishmentId)
	if err != nil{
		log.Printf("error executing the schedule query: %s", err)
		return errors.New("error deleting the schedule")
	}
	if aff, _ := result.RowsAffected(); aff == 0{
		log.Printf("no rows affected deleting the schedule: %s", err)
		return errors.New("error deleting the schedule")
	}
	return nil 
}
