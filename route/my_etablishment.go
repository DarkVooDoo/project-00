package route

import (
	"html/template"
	"log"
	"net/http"
	"planify/model"
	"strconv"
)

type MyEtablishmentRoute struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Etablishment model.Etablishment
    TodayAppointment []model.Appointment
	TopEmployee []model.Employe
    Category []model.KeyValue
    CurrentEtablishment string
}

var MyEtablishmentHandler http.Handler = &MyEtablishmentRoute{}

func (me MyEtablishmentRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        case http.MethodPost:
            me.Post(w, r)
        default:
            me.Get(w, r)
    }
}

func (me MyEtablishmentRoute) Get(w http.ResponseWriter, r *http.Request){
    // Grab cookie of the last etablishment and fetch it
    if err := VerifyToken(r, w, &me.User); err != nil{
        w.Header().Add("Location", "/connexion")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    
	etablishentId, err  := strconv.Atoi(r.URL.Query().Get("etablishment"))
	var eId int64
	if err != nil{
		log.Printf("error converting the id to number: %s", err)
		eId = me.User.Etablishment
	}else{
		eId = int64(etablishentId)
	}

    conn := model.GetDBPoolConn()
    defer conn.Close()
    
	me.Navbar = model.GetNavbarFromCache(conn, me.User)
    me.Category = model.Categorys(conn)
	me.User.Etablishment = eId
    etablishment := model.Etablishment{UserId: me.User.Id, Id: me.User.Etablishment}

    etablishment.UserEtablishment(conn)
    me.Etablishment = etablishment
	if err := model.CreateAccessToken(me.User.Id, me.User.ShortName, me.User.Etablishment, me.User.Employee, w); err != nil{
		log.Printf("Error creating the token; %s", err)
		return
	}

	employee := model.Employe{EtablishmentId: me.User.Etablishment}
    appointment := model.Appointment{UserId: me.User.Id, EtablishmentId: me.User.Etablishment}
    me.TodayAppointment = appointment.EtablishmentTodayAppointments(conn)
	me.TopEmployee = employee.TopEmployees(conn, "Monthly")
    if err := CreatePage(me, w, "view/page.html", "view/my_etablishment.tmpl", "view/component/AppointmentCard.tmpl"); err != nil{
        return
    }
}

func (me MyEtablishmentRoute) Post(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &me.User); err != nil{
        log.Printf("error reading the token")
        return
    }
	conn := model.GetDBPoolConn()
	defer conn.Close()
	filter := r.FormValue("employeeFilter")
	employee := model.Employe{EtablishmentId: me.User.Etablishment}

	employeeList := employee.TopEmployees(conn, filter)

	temp, err := template.New("employee-perf").Parse(`
	{{range .}}
    	<div class="employee-perf">
    	    {{with .Picture}}
    	        <img src="{{.Picture}}" />
    	    {{else}}
    	        <div class="preview">{{.ShortName}}</div>
    	    {{end}}
    	    <div class="content">
    	        <h3 class="name">{{.Name}}</h3>
    	        <span class="title">Barber</span>
    	    </div>
    	    <div class="numbers">{{.TotalClient}} Clients</div>
    	</div>
	{{end}}
	`)
	if err != nil{
		log.Printf("error parsing the template: %s", err)
		return
	}
	if err := temp.Execute(w, employeeList); err != nil{
		log.Printf("error executing the template: %s", err)
		return
	}
}

