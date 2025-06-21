package route

import (
	"log"
	"net/http"
	"planify/model"
	"strconv"
)

type MyEtablishmentRoute struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Etablishment model.Etablishment
    Appointment []model.Appointment
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
	if err != nil{
		log.Printf("error converting the id to number: %s", err)
		etablishentId = me.User.Etablishment
	}

    conn := model.GetDBPoolConn()
    defer conn.Close()
    
	me.Navbar = model.GetNavbarFromCache(conn, me.User)
    me.Category = model.Categorys(conn)
	me.User.Etablishment = etablishentId
    etablishment := model.Etablishment{UserId: me.User.Id, Id: me.User.Etablishment}

    etablishment.UserEtablishment(conn)
    me.Etablishment = etablishment
	if err := model.CreateAccessToken(me.User.Id, me.User.ShortName, me.User.Picture, me.User.Etablishment, me.User.Employee, w); err != nil{
		log.Printf("Error creating the token; %s", err)
		return
	}

    //if len(list) == 0{
    //    w.Header().Add("Location", "/etablissement/creer")
    //    w.WriteHeader(http.StatusTemporaryRedirect)
    //    return
    //}
    appointment := model.Appointment{UserId: me.User.Id, EtablishmentId: me.User.Etablishment}
    me.Appointment = appointment.EtablishmentUpcomingAppointments(conn)
    if err := CreatePage(me, w, "view/page.html", "view/my_etablishment.tmpl", "view/component/AppointmentCard.tmpl"); err != nil{
        return
    }
}

func (me MyEtablishmentRoute) Post(w http.ResponseWriter, r *http.Request){
    //Fetch new Etablishment and put the id in cookie
    if err := VerifyToken(r, w, &me.User); err != nil{
        log.Printf("error reading the token")
        return
    }
    //w.Header().Add("HX-Redirect", "/etablissement")
    //w.WriteHeader(http.StatusTemporaryRedirect)
}

