package route

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"planify/model"
)

type Landpage struct{
    User model.UserClaim
    Etablishments []model.Etablishment
    NextAppointment model.Appointment
}

type Notitification struct{
    Title string
    Message string
    Type string
}

var LandpageHandler http.Handler = &Landpage{}

func (l Landpage) ServeHTTP(w http.ResponseWriter, r *http.Request){
    if r.URL.Path != "/"{
        http.NotFound(w, r)
        return
    }
    switch r.Method{
        default:
            l.Get(w, r)
    }
}

func (l Landpage) Get(w http.ResponseWriter, r *http.Request){
    VerifyToken(r, w, &l.User)
    conn := model.GetDBPoolConn()
    defer conn.Close()
    var e model.Etablishment
    l.Etablishments = e.Latest(conn)
    if l.User.Id != ""{
        var appointment model.Appointment = model.Appointment{UserId: l.User.Id}
        appointment.UserNextAppointment(conn)
        l.NextAppointment = appointment
    }
    CreatePage(l, w, "view/page.html", "view/landpage.tmpl", "view/component/AppointmentCard.tmpl", "view/component/EtablishmentCard.tmpl")
}

func CreatePage(data any, w http.ResponseWriter, pattern ...string)error{
    temp, err := template.ParseFiles(pattern...)
    if err != nil{
        log.Printf("error loading template: %s", err)
        return errors.New("error parsing template")
    }
    w.Header().Add("Content-Encoding", "gzip")
    gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
    if err != nil{
        log.Printf("error compressing the file: %s", err)
        return nil
    }
    defer gz.Close()
    if err = temp.Execute(gz, data); err != nil{
        log.Printf("error executiong template: %s", err)
        return errors.New("error executing the template")
    }
    return nil
}

func VerifyToken(r *http.Request, w http.ResponseWriter, u *model.UserClaim)(error){
    token, err := r.Cookie("access-token")
    if err == nil{
        if err = u.VerifyAccessToken(token.Value, w); err != nil{
            log.Println("error verifying token")
            *u = model.UserClaim{}
            return errors.New("error verifying token")
        }
    } else{
        *u = model.UserClaim{}
        return errors.New("not token found")
    }
    return nil
}

func DisplayNotification(notif Notitification, w http.ResponseWriter){
    temp, err := template.New("Notiication").Parse(`
        <div id="notification" hx-swap-oob="true" style="background-color: var(--{{.Type}}-bg);">
          <img src="/static/{{.Type}}.svg" class="icon" />
          <div>
            <h1 class="title">{{.Title}}</h1>
            <p>{{.Message}}</p>
          </div>
          <svg class="icon" viewBox="0 0 96 96" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M18.8281 13.1719L13.1719 18.8281L42.3438 48L13.1719 77.1719L18.8281 82.8281L48 53.6562L77.1719 82.8281L82.8281 77.1719L53.6562 48L82.8281 18.8281L77.1719 13.1719L48 42.3438L18.8281 13.1719Z" fill="black"/>
          </svg>
        </div>
    `)
    if err != nil{
        log.Printf("error parsing the notification display: %s", err)
        return
    }
    temp.Execute(w, notif)
}

func ReadJsonBody(body io.ReadCloser, j interface{})error{
    dec := json.NewDecoder(body)
    if err := dec.Decode(j); err != nil{
        log.Printf("error decoding the json: %s", err)
        return errors.New("error decoding the json")
    }
    return nil
}
