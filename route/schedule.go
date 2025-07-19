package route

import (
	"html/template"
	"log"
	"net/http"
	"planify/model"
)

type ScheduleRoute struct{
    EtablishmentId int64
    Week []string
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
        case http.MethodPost:
            s.Post(w, r)
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
    s.Week = []string{"Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi", "Dimanche"}
    
    temp, err := template.New("schedule").Parse(`
			{{$schedule := .Schedule}}
            <div class="schedule">
                {{range $index, $element := .Week}}
                    <div class="day">
                        <div class="description">
                            <h1 class="label">{{$element}}</h1>
                            <div style="display:flex;gap:.5rem;">
								<button type="button" class="newShiftBtn" hx-post="/schedule" hx-ext="json-enc-custom" hx-swap="beforeend" hx-target="closest .day" hx-vals='{"day": "{{$index}}"}'></button>
                            </div>
                        </div>
						{{range $i, $v := $schedule}}
							{{if eq $v.Day $index}}
							<div class="shift" data-open="{{slice $v.OpenTime 0 5}}" data-close="{{slice $v.CloseTime 0 5}}" >
                        	    <input type="time" name="open_time" class="input ot" value="{{$v.OpenTime}}" oninput="onTimeChange(this)" />
                        	    <p>Au</p>
                        	    <input type="time" name="close_time" class="input ct" value="{{$v.CloseTime}}" oninput="onTimeChange(this)" />
								<button type="button" class="close" hx-delete="/schedule" hx-ext="json-enc-custom" hx-vals='{"id": "{{$v.Id}}"}' hx-swap="delete" hx-target="closest .shift">
                        	        <svg class="icon" viewBox="0 0 96 96" fill="none">
                        	            <path d="M18.8281 13.1719L13.1719 18.8281L42.3438 48L13.1719 77.1719L18.8281 82.8281L48 53.6562L77.1719 82.8281L82.8281 77.1719L53.6562 48L82.8281 18.8281L77.1719 13.1719L48 42.3438L18.8281 13.1719Z" fill="var(--text-color)"/>
                        	        </svg>
                        	    </button>
                                <button type="button" class="close updateBtn hidden" hx-put="/schedule" hx-ext="json-enc-custom" hx-swap="none" hx-vals='{"id": "{{$v.Id}}"}' 
								hx-include="closest .shift">
									<svg class="icon" viewBox="0 0 24 24" fill="none"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M21 3V8M21 8H16M21 8L18 5.29168C16.4077 3.86656 14.3051 3 12 3C7.02944 3 3 7.02944 3 12C3 16.9706 7.02944 21 12 21C16.2832 21 19.8675 18.008 20.777 14" stroke="var(--text-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> </g></svg>
                                </button>
                        	</div>
							{{end}}
						{{end}}
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

func (s ScheduleRoute) Post(w http.ResponseWriter, r *http.Request){

    var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		w.Header().Add("HX-Redirect", "/connexion")
		return
	}
	newSchedule := model.DaySchedule{OpenTime: "06:00:00", CloseTime: "07:00:00", EtablishmentId: user.Etablishment}
    if err := ReadJsonBody(r.Body, &newSchedule); err != nil{
        DisplayNotification(Notitification{Title: "Error", Message: "Mise a jour echoué", Type: "error"}, w)
        return
    }
	conn := model.GetDBPoolConn()
	defer conn.Close()
	if err := newSchedule.Create(conn); err != nil{
		log.Printf("error creating the schedule: %s", err)
        DisplayNotification(Notitification{Title: "Error", Message: "Mise a jour imposible", Type: "error"}, w)
		return
	}
	tmpl, err := template.New("schedule").Parse(`
         <div class="shift" data-open="{{slice .OpenTime 0 5}}" data-close="{{slice .CloseTime 0 5}}">
             <input type="time" name="from" class="input" value="{{.OpenTime}}" />
             <p>Au</p>
             <input type="time" name="to" class="input" value="{{.CloseTime}}" />
		 	<button type="button" class="close" hx-delete="/schedule" hx-vals='{"id": "{{.Id}}"}'>
                 <svg class="icon" viewBox="0 0 96 96" fill="none">
                     <path d="M18.8281 13.1719L13.1719 18.8281L42.3438 48L13.1719 77.1719L18.8281 82.8281L48 53.6562L77.1719 82.8281L82.8281 77.1719L53.6562 48L82.8281 18.8281L77.1719 13.1719L48 42.3438L18.8281 13.1719Z" fill="var(--text-color)"/>
                 </svg>
             </button>
             <button type="button" class="close updateBtn hidden" hx-put="/schedule" hx-ext="json-enc-custom" hx-swap="none" hx-vals='{"id": "{{.Id}}"}' 
			 hx-include="closest .shift">
			 	<svg class="icon" viewBox="0 0 24 24" fill="none"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M21 3V8M21 8H16M21 8L18 5.29168C16.4077 3.86656 14.3051 3 12 3C7.02944 3 3 7.02944 3 12C3 16.9706 7.02944 21 12 21C16.2832 21 19.8675 18.008 20.777 14" stroke="var(--text-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> </g></svg>
             </button>
         </div>
	`)
	if err != nil{
		log.Printf("error parsing the template")
        DisplayNotification(Notitification{Title: "Error", Message: "Parsing error", Type: "error"}, w)
		return
	}
	if err = tmpl.Execute(w, newSchedule); err != nil{
		log.Printf("error executing the template")
	}
	DisplayNotification(Notitification{"Reussi", "nouveau horaire ajouté", "success"}, w)
	
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
