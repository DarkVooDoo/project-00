package route

import (
	"log"
	"net/http"
	"planify/model"
	"strconv"
	"text/template"
	"time"
)

type PlanningPayload struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Employee model.Employe
    Schedule []model.Appointment
    Today string
}

var PlanningHandler http.Handler = &PlanningPayload{}

func (p PlanningPayload) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        case http.MethodPost:
            p.Post(w, r)
		case http.MethodPut:
			p.Put(w, r)
        default:
            p.Get(w, r)
    }
}

func (p PlanningPayload) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &p.User); err != nil{
        w.Header().Add("Location", "/connexion")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
	conn := model.GetDBPoolConn()
	defer conn.Close()
	p.Navbar = model.GetNavbarFromCache(conn, p.User)
    planning := model.Appointment{Date: time.Now().Format(time.DateOnly), EmployeeId: p.User.Employee}
    p.Schedule = planning.EmployeePlanning(conn)
    p.Today = planning.Date
    if err := CreatePage(p, w, "view/page.html", "view/planning.tmpl"); err != nil{
        return
    }
}

func (p PlanningPayload) Put(w http.ResponseWriter, r *http.Request){
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		log.Printf("error verifying token: %s", err)
		return
	}
	employeeId, err := strconv.Atoi(r.FormValue("employee"))
	if err != nil{
		log.Printf("error converting employee id to integer: %s", err)
		return
	}
	employee := model.Employe{UserId: user.Id, Id: employeeId}
	if err := employee.VerifyUserEmployee(); err != nil{
		log.Printf("error you are no employee here")
		return
	}
	if err := model.CreateAccessToken(user.Id, user.ShortName, user.Picture, user.Etablishment, employeeId, w); err != nil{
		log.Printf("error creating the token: %s", err)
		return
	}
	
}

func (p PlanningPayload) Post(w http.ResponseWriter, r *http.Request){

    if err := VerifyToken(r, w, &p.User); err != nil{
        w.Header().Add("HX-Redirect", "/connexion")
        return
    }
	planning := model.Appointment{EmployeeId: p.User.Employee}
    if err := ReadJsonBody(r.Body, &planning); err != nil{
        log.Printf("error reading the json")
        return
    }
	conn := model.GetDBPoolConn()
	defer conn.Close()
    p.Schedule = planning.EmployeePlanning(conn)
    p.Today = planning.Date

    temp, err := template.New("schedule").Parse(`
        {{range .Schedule}}
            <div class="card {{if eq .Status "Happend"}}happend{{else if eq .Status "Now"}}now{{end}}" id="appointment{{.Id}}">
              <div class="client">
                <div class="client-group">
                    <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="var(--primary-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                        <circle cx="12" cy="7" r="4"></circle>
                    </svg>
                    <div>
                        <p class="label">Client</p>
                        <h1 class="info">{{.CustomerName}}</h1>
                    </div>
                </div>
                <div class="client-group">
                    <svg class="icon" viewBox="0 0 24 24" fill="none" >
                    <path d="M13.2596 1.88032C13.3258 1.47143 13.7124 1.19406 14.1212 1.26025C14.1466 1.2651 14.228 1.28032 14.2707 1.28982C14.356 1.30882 14.475 1.33808 14.6234 1.38131C14.9202 1.46775 15.3348 1.61015 15.8324 1.83829C16.8287 2.29505 18.1545 3.09405 19.5303 4.46985C20.9061 5.84565 21.7051 7.17146 22.1619 8.16774C22.39 8.66536 22.5324 9.07996 22.6188 9.37674C22.6621 9.52515 22.6913 9.64417 22.7103 9.7295C22.7198 9.77217 22.7268 9.80643 22.7316 9.83174L22.7374 9.86294C22.8036 10.2718 22.5287 10.6743 22.1198 10.7405C21.7121 10.8065 21.328 10.5305 21.2602 10.1235C21.2581 10.1126 21.2524 10.0833 21.2462 10.0556C21.2339 10.0002 21.2125 9.91236 21.1787 9.79621C21.111 9.56388 20.9935 9.21854 20.7983 8.79287C20.4085 7.94256 19.7075 6.76837 18.4696 5.53051C17.2318 4.29265 16.0576 3.59165 15.2073 3.20182C14.7816 3.00667 14.4363 2.88913 14.2039 2.82146C14.0878 2.78763 13.9418 2.75412 13.8864 2.74178C13.4794 2.67396 13.1936 2.28804 13.2596 1.88032Z" fill="var(--primary-color)"/>
                    <path fill-rule="evenodd" clip-rule="evenodd" d="M13.4859 5.32978C13.5997 4.93151 14.0148 4.70089 14.413 4.81468L14.207 5.53583C14.413 4.81468 14.413 4.81468 14.413 4.81468L14.4145 4.8151L14.416 4.81554L14.4194 4.81651L14.4271 4.81883L14.4469 4.82499C14.462 4.82981 14.4808 4.83609 14.5033 4.84406C14.5482 4.85999 14.6075 4.88266 14.6803 4.91386C14.826 4.9763 15.0251 5.07272 15.2696 5.21743C15.7591 5.50711 16.4272 5.98829 17.2122 6.77326C17.9972 7.55823 18.4784 8.22642 18.768 8.71589C18.9128 8.9604 19.0092 9.15946 19.0716 9.30515C19.1028 9.37795 19.1255 9.43731 19.1414 9.48222C19.1494 9.50467 19.1557 9.5235 19.1605 9.53858L19.1666 9.55837L19.169 9.56612L19.1699 9.56945L19.1704 9.57098C19.1704 9.57098 19.1708 9.57243 18.4496 9.77847L19.1708 9.57242C19.2846 9.9707 19.054 10.3858 18.6557 10.4996C18.2608 10.6124 17.8493 10.3867 17.7315 9.99462L17.7278 9.98384C17.7224 9.96881 17.7114 9.93923 17.6929 9.89602C17.6559 9.80969 17.5888 9.66846 17.4772 9.47987C17.2542 9.10312 16.8515 8.53388 16.1516 7.83392C15.4516 7.13397 14.8823 6.73126 14.5056 6.5083C14.317 6.39668 14.1758 6.32958 14.0894 6.29258C14.0462 6.27407 14.0167 6.26303 14.0016 6.2577L13.9909 6.254C13.5988 6.13613 13.373 5.72468 13.4859 5.32978Z" fill="var(--primary-color)"/>
                    <path fill-rule="evenodd" clip-rule="evenodd" d="M5.00745 4.40708C6.68752 2.72701 9.52266 2.85473 10.6925 4.95085L11.3415 6.11378C12.1054 7.4826 11.7799 9.20968 10.6616 10.3417C10.6467 10.3621 10.5677 10.477 10.5579 10.6778C10.5454 10.9341 10.6364 11.5269 11.5548 12.4453C12.4729 13.3635 13.0656 13.4547 13.3221 13.4422C13.5231 13.4325 13.6381 13.3535 13.6585 13.3386C14.7905 12.2203 16.5176 11.8947 17.8864 12.6587L19.0493 13.3077C21.1454 14.4775 21.2731 17.3126 19.5931 18.9927C18.6944 19.8914 17.4995 20.6899 16.0953 20.7431C14.0144 20.822 10.5591 20.2846 7.13735 16.8628C3.71556 13.441 3.17818 9.98579 3.25706 7.90486C3.3103 6.50066 4.10879 5.30574 5.00745 4.40708ZM9.38265 5.68185C8.78363 4.60851 7.17394 4.36191 6.06811 5.46774C5.29276 6.24309 4.7887 7.0989 4.75599 7.96168C4.6902 9.69702 5.11864 12.7228 8.19801 15.8021C11.2774 18.8815 14.3031 19.31 16.0385 19.2442C16.9013 19.2115 17.7571 18.7074 18.5324 17.932C19.6382 16.8262 19.3916 15.2165 18.3183 14.6175L17.1554 13.9685C16.432 13.5648 15.4158 13.7025 14.7025 14.4158C14.6325 14.4858 14.1864 14.902 13.395 14.9405C12.5847 14.9799 11.604 14.6158 10.4942 13.506C9.38395 12.3958 9.02003 11.4148 9.0597 10.6045C9.09846 9.81294 9.51468 9.36733 9.58432 9.29768C10.2976 8.58436 10.4354 7.56819 10.0317 6.84478L9.38265 5.68185Z" fill="var(--primary-color)"/>
                    </svg>
                    <div>
                        <p class="label">Contact</p>
                        <p class="info">{{.Contact}}</p>
                    </div>
                </div>
                <div class="client-group">
                    <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="var(--primary-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path>
                    </svg>
                    <div>
                        <p class="label">Service</p>
                        <p class="info">{{.Service}}</p>
                    </div>
                </div>
                <div class="client-group">
                    <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="var(--primary-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                      <line x1="16" y1="2" x2="16" y2="6"></line>
                      <line x1="8" y1="2" x2="8" y2="6"></line>
                      <line x1="3" y1="10" x2="21" y2="10"></line>
                    </svg>
                    <div>
                        <p class="label">Date et Heure</p>
                        <p class="info">{{.Timeframe}}</p>
                    </div>
                </div>
              </div>
			  <div class="footer">
                {{if eq .Status "Terminé"}}
                    <a href="/etablissement/{{.EtablishmentId}}/rendez-vous/nouveau?s={{.ServiceTook}}" class="footerBtn btn-outline">Reserve a nouveau</a>
                {{else if eq .Status "Confirmé"}}
                    <button type="button" class="footerBtn deleteBtn" onclick="onCancelAppointment({{.Id}})">Annuler</button>
                    <a href="/rendez-vous/{{.Id}}" class="footerBtn updateBtn">Modifier</a>
                {{end}}
              </div>
            </div>
        {{else}}
            <div class="offday">
                <div class="date-display" id="current-date">{{.Today}}</div>
                <svg class="empty-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="var(--primary-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                  <line x1="16" y1="2" x2="16" y2="6"></line>
                  <line x1="8" y1="2" x2="8" y2="6"></line>
                  <line x1="3" y1="10" x2="21" y2="10"></line>
                </svg>
                <h1 class="header">Vous avez aucun client aujourd'hui</h1>
                <p class="info">Vous n'avez aucun client programmé pour cette journée. Profitez de ce temps libre pour vous organiser ou ajouter de nouveaux rendez-vous.</p>
            </div>
        {{end}}
    `)
    if err != nil{
        log.Printf("error parsing the template: %s", err)
        return
    }

    if err = temp.Execute(w, p); err != nil{
        log.Printf("error executing the template: %s", err)
    }
}
