package route

import (
	"html/template"
	"log"
	"net/http"
	"planify/model"
	"regexp"
)

type ViewAppointmentRoute struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Service []model.Service
    Appointment model.Appointment
    Customer model.User
    Employee []model.Employe
    AvailebleDates []string
}

var ViewAppointmentHandler http.Handler = ViewAppointmentRoute{}

func (va ViewAppointmentRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        case http.MethodPost:
            va.Post(w,r)
		case http.MethodPut:
			va.Put(w,r)
        default:
            va.Get(w, r)
    }
}

func (va ViewAppointmentRoute) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &va.User); err != nil{
        w.Header().Add("Location", "/connexion")
        w.WriteHeader(http.StatusTemporaryRedirect)
    }
	conn := model.GetDBPoolConn()
	defer conn.Close()
	va.Navbar = model.GetNavbarFromCache(conn, va.User)
    appointment := model.Appointment{Id: r.PathValue("id"), UserId: va.User.Id, EmployeeId: va.User.Employee}
    customer, employee, services, availebleDates, err := appointment.GetFull(conn)
    if err != nil{
        w.Header().Add("Location", "/")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    va.AvailebleDates = availebleDates
    va.Employee = employee
    va.Appointment = appointment
    va.Service = services
    va.Customer = customer
    CreatePage(va, w, "view/page.html", "view/view_appointment.tmpl")
}

func (va ViewAppointmentRoute) Post(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &va.User); err != nil{
        w.Header().Add("HX-Redirect", "/connexion")
		return
    }
	conn := model.GetDBPoolConn()
	defer conn.Close()
	appointment := model.Appointment{Id: r.PathValue("id"), UserId: va.User.Id, EmployeeId: va.User.Employee, Status: r.FormValue("status")}
	if err := appointment.UpdateStatus(conn); err != nil{
		log.Printf("error we cant complete the appointment")
		DisplayNotification(Notitification{"Echoué", "requete impossible", "error"}, w)
		return
	}
	updatePage(appointment.Status, r.Referer(), w)
	DisplayNotification(Notitification{"Reussi", "Rendez-vous terminé", "success"}, w)
}

func (va ViewAppointmentRoute) Put(w http.ResponseWriter, r *http.Request){
	if err := VerifyToken(r, w, &va.User); err != nil{
		log.Printf("unauthorized")
		return
	}
	appointment := model.Appointment{Id: r.PathValue("id"), EtablishmentId: va.User.Etablishment, UserId: va.User.Id}
	if err := ReadJsonBody(r.Body, &appointment); err != nil{
		log.Println("error reading the payload")
		return
	}
	conn := model.GetDBPoolConn()
	defer conn.Close()
	if err := appointment.UpdateAppointment(conn, va.User.Employee); err != nil{
		DisplayNotification(Notitification{"Error", "requête échoué", "error"}, w)
		return
	}
	w.Header().Add("HX-Redirect", "/planning")
}

func updatePage(t string, referer string, w http.ResponseWriter){
	
	isFullAppointment, _ := regexp.Match("/rendez-vous/[0-9]+$", []byte(referer))
	if isFullAppointment{
		var updateHtml string
		if t == "Annulé"{
			updateHtml = `<div class="status-badge cancelled" id="status-badge" hx-swap-oob="true">Annulé</div>`
		}else{
			updateHtml = `<div class="status-badge completed" id="status-badge" hx-swap-oob="true">Terminé</div>`
		}
		temp, err := template.New("update appointment").Parse(updateHtml)
		if err != nil{
			log.Printf("error parsing the template: %s", err)
			return
		}
		temp.Execute(w, nil)
	}
}
