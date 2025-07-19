package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"
)

type Employe struct{
    Id int64
    Email string
    Name string `json:"name"`
    ShortName string
    Schedule EtablishmentSchedule
    Picture string
	Joined string
	TotalClient int
    EtablishmentId int64
    UserId int64
}

var logg *slog.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
func (e *Employe) New()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    employeRow := conn.QueryRowContext(context.Background(), `WITH newEmployee AS (
            INSERT INTO employee (etablishment_id, user_id) VALUES($1, $2) ON CONFLICT (etablishment_id, user_id) DO NOTHING  RETURNING id, user_id
        )SELECT e.id, u.firstname || ' ' || u.lastname FROM newEmployee AS e JOIN users AS u ON u.id=e.user_id`, e.EtablishmentId, e.UserId)
    if err := employeRow.Scan(&e.Id, &e.Name); err != nil{
        logg.Error("error creating the employee", "type", err)
        return errors.New("errors creating the employee")
    }
    return nil
}

func (e Employe) VerifyUserEmployee() error{
	conn := GetDBPoolConn()
	defer conn.Close()

	var exist bool
	employee := conn.QueryRowContext(context.Background(), `SELECT user_id=$1 FROM employee WHERE id=$2`, e.UserId, e.Id)
	if err := employee.Scan(&exist);  err != nil{
		log.Printf("error executing the query: %s", err)
		return errors.New("error executing the query")
	}
	if !exist{
		return errors.New("error no exist")
	}
	return nil
}

func (e *Employe) GetEtablishmentEmployees(conn *sql.Conn)[]Employe{
    var allEmployee []Employe
    var schedule sql.NullString
    etablishmentEmployee, err := conn.QueryContext(context.Background(), `SELECT em.id, em.schedule, u.firstname || ' ' || u.lastname, LEFT(u.firstname, 1) || LEFT(u.lastname, 1),
	AGE(TIMESTAMP em.joined) FROM etablishment AS e RIGHT JOIN employee AS em ON em.etablishment_id=e.id LEFT JOIN users AS u ON u.id=em.user_id WHERE e.id=$1`, e.EtablishmentId)
    if err != nil{
        logg.Error("error getting employees", "description", err)
        return allEmployee
    }
    for etablishmentEmployee.Next(){
        if err = etablishmentEmployee.Scan(&e.Id, &schedule, &e.Name, &e.ShortName, &e.Joined); err != nil{
            logg.Info("error scanning row", "error scanning", err)
        }
        if schedule.Valid{
            if err = json.Unmarshal([]byte(schedule.String), &e.Schedule); err != nil{
                log.Printf("error in the json: %s", err)
            }
        }
        allEmployee = append(allEmployee, *e)
    }
    return allEmployee
}

func (e Employe) UpdateSchedule(schedule SchedulePayload) error{
    conn := GetDBPoolConn()
    defer conn.Close()

    sch, err := json.Marshal(EtablishmentSchedule{From: schedule.From, To: schedule.To})
    if err != nil{
        log.Printf("error decoding json: %s", err)
        return errors.New("erro decoding json")
    }
    result, err := conn.ExecContext(context.Background(), `UPDATE employee SET schedule=$1 WHERE id=$2`, string(sch), schedule.Id)
    if err != nil{
        logg.Error("error executing the query", "description", err)
        return errors.New("error in the query")
    }
    affected, err := result.RowsAffected()
    if affected == 0 || err != nil{
        logg.Info("Nothing happend")
        return errors.New("nothing happend")
    }
    return nil
}

func (e *Employe) TopEmployees(conn *sql.Conn)[]Employe{
	var listEmployee []Employe

	employeeRows, err := conn.QueryContext(context.Background(), `SELECT e.id, u.firstname || ' ' || u.lastname, COALESCE(u.picture, ''), UPPER(LEFT(u.firstname, 1)) || UPPER(LEFT(u.lastname, 1)),
	(SELECT COUNT(a.id) FROM appointment AS a WHERE a.employee_id=e.id AND a.status='Confirmé') FROM employee AS e LEFT JOIN users AS u ON u.id=e.user_id WHERE e.etablishment_id=$1`, e.EtablishmentId)
	if err != nil{
		log.Printf("error in the query top employees: %s", err)
		return listEmployee
	}
	for employeeRows.Next(){
		if err := employeeRows.Scan(&e.Id, &e.Name, &e.Picture, &e.ShortName, &e.TotalClient); err != nil{
			log.Printf("error scanning the top employees: %s", err)
			continue
		}
		listEmployee = append(listEmployee, *e)
	}
	return listEmployee
}

func (e *Employe) Delete()error{
    conn := GetDBPoolConn()
    defer conn.Close()

    result, err := conn.ExecContext(context.Background(), `DELETE FROM employee WHERE id=$1`, e.Id)
    if err != nil{
        logg.Error("Error", "description", err)
        return errors.New("error deleting employee")
    }
    if affected, err := result.RowsAffected(); err != nil || affected == 0{
        logg.Error("Error", "description", err)
        return errors.New("error impossible de supprimer ce employee")
    }
    return nil
}

func (e *Employe) SuggestEmployee()[]Employe{
    conn := GetDBPoolConn()
    defer conn.Close()

    var list []Employe
    employeList, err := conn.QueryContext(context.Background(), `SELECT u.id, u.email FROM users AS u WHERE u.email LIKE $1 AND 
	NOT EXISTS(SELECT 1 FROM employee AS e WHERE e.user_id = u.id AND e.etablishment_id=$2) LIMIT 5`, e.Email + "%", e.EtablishmentId)
    if err != nil{
        logg.Error("error searching user", "description", err)
        return list
    }

    for employeList.Next(){
        if err = employeList.Scan(&e.UserId, &e.Email); err != nil{
            logg.Error("error scanning the table")
        }
        list = append(list, *e)
    }
    return list
}
