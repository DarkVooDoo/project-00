package route

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"planify/model"
	"strconv"
	"strings"
	"time"
)


type NewAppointment struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Etablishment model.Etablishment
    EmployeeId int `json:"employee,string"`
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
	conn := model.GetDBPoolConn()
	defer conn.Close()
	n.Navbar = model.GetNavbarFromCache(conn, n.User)
	id, _ := strconv.Atoi(r.PathValue("id"))
    etablishment := model.Etablishment{Id: id}
    etablishment.GetEmployeeAndService(conn)
	autofill := r.URL.Query().Get("s")
	if autofill != ""{
		serviceIds := strings.Split(autofill, ",")
		for index, v := range etablishment.Service{
			for _, cId := range serviceIds{
				idChecked, _ := strconv.Atoi(cId)
				if v.Id == idChecked{
					etablishment.Service[index].Checked = true
				}
			}
		}
	}
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
	_, err := time.Parse(time.DateTime, fmt.Sprintf("%s %s:00", n.Date, n.Time))
	if n.EmployeeId == 0 || len(n.Service) == 0 || err != nil{
		DisplayNotification(Notitification{"Info", "Formulaire incomplet", "warning"}, w)
		return
	}
	etablishmentId, _ := strconv.Atoi(r.PathValue("id"))
    var newAppointment model.Appointment = model.Appointment{EmployeeId: n.EmployeeId, Date: fmt.Sprintf("%s %s", n.Date, n.Time), UserId: n.User.Id, Services: n.Service, EtablishmentId: etablishmentId}
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

	etablishmentId, _ := strconv.Atoi(r.PathValue("id"))
    appointment := model.Appointment{UserId: n.User.Id, EmployeeId: n.EmployeeId, Date: n.Date, EtablishmentId: etablishmentId}
    hours := appointment.AvaileblesDates(conn)
    temp, err := template.New("test").Parse(`
		{{range $index, $value := .}}
			<button type="button" name="time" class="btn btn-outline" onclick="onTimePick(this)" data-index="{{$index}}">{{$value}}</button>
		{{else}}
			<div class="notime">
				<h2 class="notime-header">Aucun créneau disponible</h2>
				<div class="notime-message">
					<svg class="icon" viewBox="0 0 24 24" fill="none">
						<g id="SVGRepo_bgCarrier" stroke-width="0"></g>
						<g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g>
						<g id="SVGRepo_iconCarrier"> 
							<circle cx="12" cy="17" r="1" fill="var(--warning-fg)"></circle>
							<path d="M12 10L12 14" stroke="var(--warning-fg)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> 
							<path d="M3.44722 18.1056L10.2111 4.57771C10.9482 3.10361 13.0518 3.10362 13.7889 4.57771L20.5528 18.1056C21.2177 19.4354 20.2507 21 18.7639 21H5.23607C3.7493 21 2.78231 19.4354 3.44722 18.1056Z" stroke="var(--warning-fg)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> 
						</g>
					</svg>
					<p class="info">Nous sommes désolés, mais aucun créneau n'est disponible pour la date sélectionnée.</p>
				</div>
			</div>
		{{end}}`)
    if err != nil{
        log.Printf("error parsing the template: %s", err)
        return
    }
    if err := temp.Execute(w, hours); err != nil{
        log.Printf("error executiong template: %s", err)
    }
}
