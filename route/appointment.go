package route

import (
	"log"
	"net/http"
	"planify/model"
	"strconv"
)

type AppointmentRoute struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Appointment []model.Appointment
	PageCount []string
	CurrentPage int
    Type string
}

type AppointmentPayload struct{
    EmployeeId int64 `json:"employee,string"`
    Service []model.ServicePayload `json:"service"`
    Time string `json:"time"`
    Date string `json:"date"`
	Phone string `json:"phone"`
	Name string `json:"name"`
}

var AppointmentHandler http.Handler = AppointmentRoute{}

func (a AppointmentRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
		case http.MethodPost:
			a.Post(w, r)
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
	conn := model.GetDBPoolConn()
	defer conn.Close()
	a.Navbar = model.GetNavbarFromCache(conn, a.User)
	appointment := model.Appointment{UserId: a.User.Id}
	a.Type = r.URL.Query().Get("type")
	switch a.Type{
	case "upcomming":
		appointment.Status = "Confirmé"
	case "foregoing":
		appointment.Status = "Terminé"
	case "cancelled":
		appointment.Status = "Annulé"
	default:
		appointment.Status = "All"
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil{
		a.CurrentPage = 1
	}else{
		a.CurrentPage = page
	}
	appointments, pageCount := appointment.UserAppointment(conn, a.CurrentPage)
	a.Appointment = appointments
	a.PageCount = make([]string, pageCount + 1)
	if err := CreatePage(a, w, "view/page.html", "view/appointment.tmpl"); err != nil{
		log.Printf("error creating the page")
	}
}

func (a AppointmentRoute) Post(w http.ResponseWriter, r *http.Request){
	if err := VerifyToken(r, w, &a.User); err != nil{
		log.Printf("unauthorized")
		return
	}
	//appointment := model.Appointment{UserId: a.User.Id}

}
