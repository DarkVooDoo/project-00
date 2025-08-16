package route

import (
	"compress/gzip"
	"html/template"
	"log"
	"net/http"
	"planify/model"
	"strconv"
	"time"
)

type Appointment struct{
    Id string
    Employee string `json:"employee"`
    Service []string `json:"service"`
    Date string `json:"date"`
	Status string `json:"status"`
	AppointmentNumbers model.DayAppointmentRecap
	EtablishmentId int64
	List []model.EmployeeWithAppointment
}

var EtablishmentAppointmentHandler http.Handler = &Appointment{}

func (a Appointment) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
		case http.MethodPost:
			a.Post(w, r)
        default:
            a.Get(w, r)
    }
}

func (a Appointment) Get(w http.ResponseWriter, r *http.Request){
    var user model.UserClaim
    if err := VerifyToken(r, w, &user); err != nil{
        w.Header().Add("HX-Redirect", "/connexion")
        return
    }
    conn := model.GetDBPoolConn()
    defer conn.Close()

	etablishmentId, _ := strconv.Atoi(r.PathValue("id"))
	date := time.Now().Format(time.DateOnly)
    appointment := model.Appointment{EtablishmentId: int64(etablishmentId), Date: date}
	a.AppointmentNumbers = appointment.EmployeeAppointmentDayInNumbers(conn)
	a.List =  appointment.EtablishmentAppointments(conn)
	a.EtablishmentId = int64(etablishmentId)

	if err := CreatePage(a, w, "view/my_etablishment_appointment.tmpl", "view/component/etablishment_appointment.tmpl"); err != nil{
		log.Printf("error creating the page: %s", err)
		return
	}
}

func (a Appointment) Post(w http.ResponseWriter, r *http.Request){
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		log.Printf("unauthorized user: %s", err)
		w.Header().Add("HX-Redirect", "/connexion")
		return
	}

	if err := ReadJsonBody(r.Body, &a); err != nil{
		log.Printf("error reading the json payload: %s", err)
		return
	}
	conn := model.GetDBPoolConn()
	defer conn.Close()
	id, _ := strconv.Atoi(r.PathValue("id"))
	appointment := model.Appointment{Status: a.Status, Date: a.Date, EtablishmentId: int64(id)}
	a.AppointmentNumbers = appointment.EmployeeAppointmentDayInNumbers(conn)
	a.List =  appointment.EtablishmentAppointments(conn)
	a.EtablishmentId = int64(id)

	temp, err := template.New("Appointment").ParseFiles("view/component/etablishment_appointment.tmpl")
	if err != nil{
		log.Printf("error creating the template: %s", err)
		return
	}
    w.Header().Add("Content-Encoding", "gzip")
    gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
    if err != nil{
        log.Printf("error compressing the file: %s", err)
    }
    defer gz.Close()
	if err := temp.Execute(gz, a); err != nil{
		log.Printf("error executing the template: %s", err)
		return
	}
}
