package route

import (
	"log"
	"net/http"
	"planify/model"
	"strconv"
	"text/template"
	"time"
)

type PlanningPayload struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Employee model.Employe
    Schedule []model.Appointment
	Shift []string
    Today string
	Recap model.DayAppointmentRecap
}

var PlanningHandler http.Handler = &PlanningPayload{}

func (p PlanningPayload) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        case http.MethodPost:
            p.Post(w, r)
		case http.MethodPut:
			p.Put(w, r)
        default:
            p.Get(w, r)
    }
}

func (p PlanningPayload) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &p.User); err != nil{
        w.Header().Add("Location", "/connexion")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
	conn := model.GetDBPoolConn()
	defer conn.Close()
	p.Navbar = model.GetNavbarFromCache(conn, p.User)
    planning := model.Appointment{Date: time.Now().Format(time.DateOnly), EmployeeId: p.User.Employee}
	p.Recap = planning.EmployeeAppointmentDayInNumbers(conn)
	schedule, shift := planning.EmployeePlanning(conn)
	p.Schedule = schedule
	p.Shift = shift
    p.Today = planning.Date
    if err := CreatePage(p, w, "view/page.html", "view/planning.tmpl", "view/component/day-planning.tmpl"); err != nil{
        return
    }
}

func (p PlanningPayload) Put(w http.ResponseWriter, r *http.Request){
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		log.Printf("error verifying token: %s", err)
		return
	}
	employeeId, err := strconv.Atoi(r.FormValue("employee"))
	if err != nil{
		log.Printf("error converting employee id to integer: %s", err)
		return
	}
	employee := model.Employe{UserId: user.Id, Id: int64(employeeId)}
	if err := employee.VerifyUserEmployee(); err != nil{
		log.Printf("error you are no employee here")
		return
	}
	if err := model.CreateAccessToken(user.Id, user.ShortName, user.Picture, user.Etablishment, int64(employeeId), w); err != nil{
		log.Printf("error creating the token: %s", err)
		return
	}
}

func (p PlanningPayload) Post(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &p.User); err != nil{
        w.Header().Add("HX-Redirect", "/connexion")
        return
    }
	planning := model.Appointment{EmployeeId: p.User.Employee}
    if err := ReadJsonBody(r.Body, &planning); err != nil{
        log.Printf("error reading the json")
        return
    }
	conn := model.GetDBPoolConn()
	defer conn.Close()
	p.Recap = planning.EmployeeAppointmentDayInNumbers(conn)
	schedule, shift := planning.EmployeePlanning(conn)
    p.Schedule = schedule
	p.Shift = shift
    p.Today = planning.Date

	temp, err := template.ParseGlob("view/component/day-planning.tmpl")
    if err != nil{
        log.Printf("error parsing the template: %s", err)
        return
    }
    if err = temp.ExecuteTemplate(w, "Planning", p); err != nil{
        log.Printf("error executing the template: %s", err)
    }
}
