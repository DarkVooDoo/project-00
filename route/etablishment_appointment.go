package route

import (
	"log"
	"net/http"
	"planify/model"
	"text/template"
)

type Appointment struct{
    Id string
    Employee string `json:"employee"`
    Service []string `json:"service"`
    Date string `json:"date"`
}

var EtablishmentAppointmentHandler http.Handler = &Appointment{}

func (a Appointment) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
        default:
            a.Get(w, r)
    }
}

func (a Appointment) Get(w http.ResponseWriter, r *http.Request){
    var user model.UserClaim
    if err := VerifyToken(r, w, &user); err != nil{
        w.Header().Add("Location", "/")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    conn := model.GetDBPoolConn()
    defer conn.Close()

    appointment := model.Appointment{EtablishmentId: r.PathValue("id")}
    var myAppointments []model.Appointment
    switch r.URL.Query().Get("type"){
        case "oldest":
            myAppointments = appointment.EtablishmentForegoingAppointments(conn)
        default:
            myAppointments = appointment.EtablishmentUpcomingAppointments(conn)
    }

    test, err := template.ParseGlob("view/component/AppointmentCard.tmpl")
    if err != nil{
        log.Printf("error creating the template: %s", err)
        return
    }
    temp, err := test.New("appointments").Parse(`
        {{range .}}
            {{template "AppointmentCard" .}}
        {{end}}
    `)
    if err != nil{
        log.Printf("error creating the template: %s", err)
        return
    }
    if err = temp.Execute(w, myAppointments); err != nil{
        log.Printf("error executing template: %s", err)
    }
}

