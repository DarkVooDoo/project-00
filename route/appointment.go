package route

import (
	"log"
	"net/http"
	"planify/model"
)

type AppointmentRoute struct{
    User model.UserClaim
    Appointment []model.Appointment
    Type string
}

var AppointmentHandler http.Handler = AppointmentRoute{}

func (a AppointmentRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        default:
            a.Get(w, r)
    }
}

func (a AppointmentRoute) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &a.User); err != nil{
        w.Header().Add("Location", "/")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    appointment := model.Appointment{UserId: a.User.Id}
    a.Type = r.URL.Query().Get("type")
    switch a.Type{
        case "upcomming":
            a.Appointment = appointment.UserAllUpcommingAppointment()
        case "foregoing":
            a.Appointment = appointment.UserAllForegoingAppointment()
        case "cancelled":
            a.Appointment = appointment.UserAllCancelledAppointment()
        default:
            a.Appointment = appointment.UserAllAppointment()
    }
    if err := CreatePage(a, w, "view/page.html", "view/appointment.tmpl"); err != nil{
        log.Printf("error creating the page")
    }
}
