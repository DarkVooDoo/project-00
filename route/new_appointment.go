package route

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"planify/model"
)


type NewAppointment struct{
    User model.UserClaim
    Etablishment model.Etablishment
    EmployeeId string `json:"employee"`
    Service []model.ServicePayload `json:"service"`
    Time string `json:"time"`
    Date string `json:"date"`
}

var NewAppointmentHandler http.Handler = &NewAppointment{}

func (a NewAppointment) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        case http.MethodPut:
            a.Put(w, r)
        case http.MethodPost:
            a.Post(w, r)
        default:
            a.Get(w, r)
    }

}

func (n NewAppointment) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &n.User); err != nil{
        w.Header().Add("Location", "/connexion")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    etablishment := model.Etablishment{Id: r.PathValue("id")}
    etablishment.GetEmployeeAndService()
    n.Etablishment = etablishment
    if err := CreatePage(n, w, "view/page.html", "view/new_appointment.tmpl"); err != nil{
        return
    }
}

func (n NewAppointment) Post(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &n.User); err != nil{
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    if err := ReadJsonBody(r.Body, &n); err != nil{
        log.Printf("error reading the body: %s", err)
        return
    }
    var newAppointment model.Appointment = model.Appointment{EmployeeId: n.EmployeeId, Date: fmt.Sprintf("%s %s", n.Date, n.Time), UserId: n.User.Id, Services: n.Service, EtablishmentId: r.PathValue("id")}
    if err := newAppointment.Create(); err != nil{
        DisplayNotification(Notitification{"Error", "requete echoué", "error"}, w)
        return
    }
    w.Header().Add("HX-Redirect", "/")
    w.WriteHeader(http.StatusTemporaryRedirect)
}

func (n NewAppointment) Put(w http.ResponseWriter, r *http.Request){
    VerifyToken(r, w, &n.User)
    if err := ReadJsonBody(r.Body, &n); err != nil{
        log.Printf("error reading payload: %s", err)
        return
    }
    conn := model.GetDBPoolConn()
    defer conn.Close()

    appointment := model.Appointment{UserId: n.User.Id, EmployeeId: n.EmployeeId, Date: n.Date, EtablishmentId: r.PathValue("id")}
    hours := appointment.AvaileblesDates(conn)
    temp, err := template.New("test").Parse(`{{range $index, $value := .}}<button type="button" name="time" class="btn btn-outline" onclick="onTimePick(this)" data-index="{{$index}}">{{$value}}</button>{{end}}`)
    if err != nil{
        log.Printf("error parsing the template: %s", err)
        return
    }
    if err := temp.Execute(w, hours); err != nil{
        log.Printf("error executiong template: %s", err)
    }
}
