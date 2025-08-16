package route

import (
	"html/template"
	"log"
	"net/http"
	"planify/model"
)

type ScheduleRoute struct{
    EtablishmentId int64
    Schedule []model.DaySchedule
}

type Schift struct{
    
}


var ScheduleHandler http.Handler = ScheduleRoute{}

func (s ScheduleRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
		case http.MethodDelete:
			s.Delete(w, r)
		case http.MethodPut:
			s.Put(w, r)
        default:
            s.Get(w, r)

    }
}

func (s ScheduleRoute) Get(w http.ResponseWriter, r *http.Request){

    var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		w.Header().Add("HX-Redirect", "/")
		return
	}

	conn := model.GetDBPoolConn()
	defer conn.Close()
    schedule := model.DaySchedule{EtablishmentId: user.Etablishment}
	s.Schedule = schedule.GetSchedule(conn)

    temp, err := template.New("schedule").Parse(`
        <div class="schedule">
			{{range .Schedule}}
            	<div class="day">
                   	<div class="description">
                   	    <h1 class="label">{{.DayName}}</h1>
			   			<button type="button" class="close updateBtn hidden" hx-put="/schedule" hx-ext="json-enc-custom" hx-swap="none" hx-vals='{"id": "{{.Id}}"}' 
			   			hx-include="closest .day">
			   				<svg class="icon" viewBox="0 0 24 24" fill="none"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M21 3V8M21 8H16M21 8L18 5.29168C16.4077 3.86656 14.3051 3 12 3C7.02944 3 3 7.02944 3 12C3 16.9706 7.02944 21 12 21C16.2832 21 19.8675 18.008 20.777 14" stroke="var(--text-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> </g></svg>
                   	    </button>
                   	</div>
			   		<div class="shift" data-open="{{slice .OpenTime 0 5}}" data-close="{{if .CloseTime}}{{slice .CloseTime 0 5}}{{end}}" >
                   	    <input type="time" name="open_time" class="input ot" value="{{.OpenTime}}" oninput="onTimeChange(this)" />
                   	    <p>Au</p>
                   	    <input type="time" name="close_time" class="input ct" value="{{.CloseTime}}" oninput="onTimeChange(this)" />
			   			<button type="button" class="close" hx-delete="/schedule" hx-ext="json-enc-custom" hx-vals='{"id": "{{.Id}}"}' hx-swap="delete" hx-target="closest .shift">
                   	        <svg class="icon" viewBox="0 0 96 96" fill="none">
                   	            <path d="M18.8281 13.1719L13.1719 18.8281L42.3438 48L13.1719 77.1719L18.8281 82.8281L48 53.6562L77.1719 82.8281L82.8281 77.1719L53.6562 48L82.8281 18.8281L77.1719 13.1719L48 42.3438L18.8281 13.1719Z" fill="var(--text-color)"/>
                   	        </svg>
                   	    </button>
                   	</div>
               </div>
			{{end}}
        </div>
    `)
    if err != nil{
        log.Printf("error parsing the template: %s", err)
        return
    }
    if err := temp.Execute(w, s); err != nil{
        log.Printf("error executing: %s", err)
    }
}

func(s ScheduleRoute) Put(w http.ResponseWriter, r *http.Request){
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		w.Header().Add("HX-Redirect", "/connexion")
		return
	}
	schedule := model.DaySchedule{EtablishmentId: user.Etablishment}
    if err := ReadJsonBody(r.Body, &schedule); err != nil{
        DisplayNotification(Notitification{Title: "Error", Message: "Mise a jour echoué", Type: "error"}, w)
        return
    }
	if err := schedule.Update(); err != nil{
		log.Printf("error updating the etablishment schedule")
		DisplayNotification(Notitification{"Error", "Mise a jour du horaire impossible", "error"}, w)
	}
	DisplayNotification(Notitification{"Reussi", "Mise a jour reussi", "success"}, w)
}

func (s ScheduleRoute) Delete(w http.ResponseWriter, r *http.Request){
    var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		w.Header().Add("HX-Redirect", "/connexion")
		return
	}
	schedule := model.DaySchedule{EtablishmentId: user.Etablishment}
    if err := ReadJsonBody(r.Body, &schedule); err != nil{
        DisplayNotification(Notitification{Title: "Error", Message: "Mise a jour echoué", Type: "error"}, w)
        return
    }
	if err := schedule.Delete(); err != nil{
		DisplayNotification(Notitification{Title: "Error", Message: "Impossible de supprimer l'horaire", Type: "error"}, w)
		return
	}
	DisplayNotification(Notitification{"Reussi", "l'horaire supprimé", "success"}, w)
}
