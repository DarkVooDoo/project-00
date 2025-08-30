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
	Recent []model.Etablishment
    NextAppointment model.Appointment
	Navbar model.CacheNavbar
	Review model.Review
}

type Notitification struct{
    Title string
    Message string
    Type string
}

var LandpageHandler http.Handler = &Landpage{}

func (l Landpage) ServeHTTP(w http.ResponseWriter, r *http.Request){
    if r.URL.Path != "/"{
		VerifyToken(r, w, &l.User)
		if err := CreatePage(l, w, "view/page.html", "view/notfound.tmpl"); err != nil{
			log.Printf("error creating the page: %s", err)
		}
        return
    }
    switch r.Method{
        default:
            l.Get(w, r)
    }
}

func (l Landpage) Get(w http.ResponseWriter, r *http.Request){
	err := VerifyToken(r, w, &l.User)
    conn := model.GetDBPoolConn()
    defer conn.Close()
	if err == nil{
		e := model.Etablishment{UserId: l.User.Id} 
		appointment := model.Appointment{UserId: l.User.Id}
		l.Etablishments = e.Latest(conn)
		l.Recent = e.Recent(conn)
		review := model.Review{UserId: l.User.Id}
		if err := review.Get(conn); err != nil{
			log.Printf("error getting the review: %s", err)
		}
		l.Review = review
        appointment.UserNextAppointment(conn)
        l.NextAppointment = appointment
		l.Navbar = model.GetNavbarFromCache(conn, l.User)
    }
    CreatePage(l, w, "view/page.html", "view/landpage.tmpl", "view/component/AppointmentCard.tmpl", "view/component/EtablishmentCard.tmpl", "view/component/review_modal.tmpl")
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
		  	{{if eq .Type "success"}}
		  	  	<svg class="icon" viewBox="0 0 96 96" fill="none">
		  	  		<g clip-path="url(#clip0_487_429)">
		  	  			<path d="M66.36 30.32L40 56.68L25.64 42.36L20 48L40 68L72 36L66.36 30.32ZM48 8C25.92 8 8 25.92 8 48C8 70.08 25.92 88 48 88C70.08 88 88 70.08 88 48C88 25.92 70.08 8 48 8ZM48 80C30.32 80 16 65.68 16 48C16 30.32 30.32 16 48 16C65.68 16 80 30.32 80 48C80 65.68 65.68 80 48 80Z" fill="#356A36"/>
		  	  		</g>
		  	  		<defs>
		  	  			<clipPath id="clip0_487_429">
		  	  				<rect width="96" height="96" fill="white"/>
		  	  			</clipPath>
		  	  		</defs>
		  	  	</svg>
		  	{{else if eq .Type "warning"}}
		  	  	<svg class="icon" viewBox="0 0 96 96" fill="none">
		  	  		<g clip-path="url(#clip0_490_434)">
		  	  			<path d="M48 23.96L78.12 76H17.88L48 23.96ZM48 8L4 84H92L48 8Z" fill="#D1834A"/>
		  	  			<path d="M52 64H44V72H52V64Z" fill="#D1834A"/>
		  	  			<path d="M52 40H44V60H52V40Z" fill="#D1834A"/>
		  	  		</g>
		  	  		<defs>
		  	  			<clipPath id="clip0_490_434">
		  	  				<rect width="96" height="96" fill="white"/>
		  	  			</clipPath>
		  	  		</defs>
		  	  	</svg>
		  	{{else if eq .Type "error"}}
				<svg class="icon" viewBox="0 0 96 96" fill="none" xmlns="http://www.w3.org/2000/svg">
					<g clip-path="url(#clip0_491_443)">
						<path d="M62.92 12H33.08L12 33.08V62.92L33.08 84H62.92L84 62.92V33.08L62.92 12ZM76 59.6L59.6 76H36.4L20 59.6V36.4L36.4 20H59.6L76 36.4V59.6Z" fill="#9B2127"/>
						<path d="M52 28H44V52H52V28Z" fill="#9B2127"/>
						<path d="M52 60H44V68H52V60Z" fill="#9B2127"/>
					</g>
					<defs>
						<clipPath id="clip0_491_443">
							<rect width="96" height="96" fill="white"/>
						</clipPath>
					</defs>
				</svg>
			{{end}}
          	<div>
          	  <h1 class="n-title">{{.Title}}</h1>
          	  <p class="n-msg">{{.Message}}</p>
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
