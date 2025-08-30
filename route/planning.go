package route

import (
	"html/template"
	"log"
	"net/http"
	"planify/model"
	"strconv"
	"time"
)

type PlanningPayload struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Employee model.Employe
    Schedule []model.Appointment
	Shift []string
    Today string
	Recap model.DayAppointmentRecap
}

var PlanningHandler http.Handler = &PlanningPayload{}

func (p PlanningPayload) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        case http.MethodPost:
            p.Post(w, r)
		case http.MethodPatch:
			p.Patch(w, r)
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
	p.Recap = planning.EmployeeAppointmentDayInNumbers(conn)
	schedule, shift := planning.EmployeePlanning(conn)
	p.Schedule = schedule
	p.Shift = shift
	p.Today = planning.Date
	if err := CreatePage(p, w, "view/page.html", "view/planning.tmpl", "view/component/day-planning.tmpl"); err != nil{
		return
	}
}

func (p PlanningPayload) Patch(w http.ResponseWriter, r *http.Request){
	var appointment model.Appointment
	if err := ReadJsonBody(r.Body, &appointment); err != nil{
		log.Printf("error reading the payload: %s", err)
		return
	}
	temp, err := template.New("appointment-modal").Parse(`
	{{$total := .Price}}
	<div class="modal-overlay active" id="appointmentModal">
    	<div class="modal">
    	  <div class="modal-header">
    	    <button type="button" class="modal-close" onclick='this.closest("#appointmentModal").classList.remove("active")
			body.style.overflow=""'>
    	      <svg class="modal-close-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    	        <line x1="18" y1="6" x2="6" y2="18"></line>
    	        <line x1="6" y1="6" x2="18" y2="18"></line>
    	      </svg>
    	    </button>
    	    
    	    <h2 class="modal-title" id="modalTitle">Détails du rendez-vous</h2>
    	    <p class="modal-subtitle" id="modalSubtitle">Informations complètes</p>
    	  </div>
    	  
    	  <div class="modal-body">
    	    <div class="appointment-detail">
    	      <div class="detail-row">
    	        <svg class="detail-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    	          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
    	          <circle cx="12" cy="7" r="4"></circle>
    	        </svg>
    	        <div class="detail-content">
    	          <div class="detail-label">Client</div>
    	          <div class="detail-value" id="modalClient">{{.CustomerName}}</div>
    	        </div>
    	      </div>
    	      
    	      <div class="detail-row">
    	        <svg class="detail-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    	          <circle cx="12" cy="12" r="10"></circle>
    	          <polyline points="12 6 12 12 16 14"></polyline>
    	        </svg>
    	        <div class="detail-content">
    	          <div class="detail-label">Horaire</div>
    	          <div class="detail-value" id="modalTime">{{.Timeframe}}</div>
    	        </div>
    	      </div>
    	      
    	      <div class="detail-row">
    	        <svg class="detail-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    	          <circle cx="12" cy="12" r="3"></circle>
    	          <path d="M12 1v6m0 6v6"></path>
    	        </svg>
    	        <div class="detail-content">
    	          <div class="detail-label">Statut</div>
    	          <div class="detail-value" id="modalStatus">{{.Status}}</div>
    	        </div>
    	      </div>
    	      
    	      <div class="detail-row">
    	        <svg class="detail-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    	          <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"></path>
    	        </svg>
    	        <div class="detail-content">
    	          <div class="detail-label">Téléphone</div>
    	          <div class="detail-value" id="modalPhone">{{.Contact}}</div>
    	        </div>
    	      </div>
    	    </div>
    	    
    	    <div class="services-list">
    	      <h4 class="services-title">Services demandés</h4>
    	      <div id="modalServices">
			  	{{range .Services}}
			  		<div class="service-item">
						<span class="service-name">{{.Name}}</span>
          				<span class="service-price">{{.CurrencyPrice}}</span>
					</div>
				{{end}}
			  	<div class="service-item">
					<span class="service-name">Total</span>
          			<span class="service-price">{{$total}}</span>
				</div>
    	      </div>
    	    </div>
    	    
    	    <div class="modal-actions">
    	      <a href="/rendez-vous/{{.Id}}" class="btn btn-outline">
    	        <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    	          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
    	          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
    	        </svg>
    	        Modifier
    	      </a>
    	      <button class="btn btn-primary" hx-post="/rendez-vous/{{.Id}}" hx-swap="none" hx-vals='{"status": "Annulé"}'>
    	        <svg class="btn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    	          <path d="M9 11l3 3l8-8"></path>
    	          <path d="M21 12c0 4.97-4.03 9-9 9s-9-4.03-9-9 4.03-9 9-9c1.51 0 2.93.37 4.18 1.03"></path>
    	        </svg>
    	        Marquer terminé
    	      </button>
    	    </div>
    	  </div>
    	</div>
	</div>
	`)
	if err != nil{
		log.Printf("error parsing the template: %s", err)
		return
	}
	if err = temp.Execute(w, appointment); err != nil{
		log.Printf("error executing the template: %s", err)
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
	employee := model.Employe{UserId: user.Id, Id: int64(employeeId)}
	if err := employee.VerifyUserEmployee(); err != nil{
		log.Printf("error you are no employee here")
		return
	}
	if err := model.CreateAccessToken(user.Id, user.ShortName, user.Picture, user.Etablishment, int64(employeeId), w); err != nil{
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
	p.Recap = planning.EmployeeAppointmentDayInNumbers(conn)
	schedule, shift := planning.EmployeePlanning(conn)
    p.Schedule = schedule
	p.Shift = shift
    p.Today = planning.Date

	temp, err := template.ParseGlob("view/component/day-planning.tmpl")
    if err != nil{
        log.Printf("error parsing the template: %s", err)
        return
    }
    if err = temp.ExecuteTemplate(w, "Planning", p); err != nil{
        log.Printf("error executing the template: %s", err)
    }
}
