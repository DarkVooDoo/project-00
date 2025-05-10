package route

import (
	"log"
	"net/http"
	"planify/model"
	"time"
)

type MyEtablishmentRoute struct{
    User model.UserClaim
    Etablishment []model.Etablishment
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
    
    etablishment := model.Etablishment{UserId: me.User.Id}
    conn := model.GetDBPoolConn()
    defer conn.Close()
    
    list, _ := etablishment.UserEtablishments(conn)
    me.Etablishment = list
    me.Category = model.Categorys(conn)
    etablishmentCookie, err := r.Cookie("eid")
    if len(list) > 0 && err != nil{
        me.CurrentEtablishment = list[0].Id
    }else if err == nil{
        etablishentQuery := r.URL.Query().Get("etablishment")
        if  etablishentQuery == ""{
            me.CurrentEtablishment = etablishmentCookie.Value
        }else{
            me.CurrentEtablishment = etablishentQuery
        }
    }else if len(list) == 0{
        w.Header().Add("Location", "/etablissement/creer")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    appointment := model.Appointment{UserId: me.User.Id, EtablishmentId: me.CurrentEtablishment}
    me.Appointment = appointment.EtablishmentUpcomingAppointments(conn)
    cookie := http.Cookie{
        Name: "eid",
        Value: me.CurrentEtablishment,
        HttpOnly: true,
        Secure: true,
        Expires: time.Now().Add(time.Hour * 3),
        SameSite: http.SameSiteStrictMode,
        Path: "/etablissement",
    }
    http.SetCookie(w, &cookie)
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
    cookie := http.Cookie{
        Name: "eid",
        Value: r.FormValue("etablishment"),
        HttpOnly: true,
        Secure: true,
        Expires: time.Now().Add(time.Hour * 3),
        SameSite: http.SameSiteStrictMode,
        Path: "/etablishment",
    }
    log.Println(cookie.Value)
    http.SetCookie(w, &cookie)
    //w.Header().Add("HX-Redirect", "/etablissement")
    //w.WriteHeader(http.StatusTemporaryRedirect)
}

