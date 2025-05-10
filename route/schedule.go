package route

import (
	"html/template"
	"log"
	"net/http"
	"planify/model"
)

type ScheduleRoute struct{
    Id string
    Week []string
    Schedule model.EtablishmentSchedule
}

type Schift struct{
    
}


var ScheduleHandler http.Handler = ScheduleRoute{}

func (s ScheduleRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
        case http.MethodPost:
            s.Post(w, r)
        default:
            s.Get(w, r)

    }
}

func (s ScheduleRoute) Get(w http.ResponseWriter, r *http.Request){

    var user model.UserClaim
    VerifyToken(r, w, &user)

    //TODO: obtenir le etablissement via le cookie
    entreprise := model.Etablishment{Id: "1"}
    schedule := entreprise.GetSchedule()
    s.Id = entreprise.Id
    s.Schedule = schedule.EtablishmentSchedule
    s.Week = []string{"Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi", "Dimanche"}
    
    temp, err := template.New("schedule").Parse(`
            {{$to := .Schedule.To}}
            {{$from := .Schedule.From}}
            <form hx-post="/schedule" hx-ext="json-enc-custom" hx-swap="none" class="schedule" hx-vals='{"id": "{{.Id}}"}'>
                {{range $index, $element := .Week}}
                    <div class="day">
                        <div class="description">
                            <h1 class="label">{{$element}}</h1>
                            <div style="display:flex;gap:.5rem;">
                                <button type="button" class="close" onclick="onCopyShift(this)">
                                    <svg class="icon" viewBox="0 0 54 64" fill="none">
                                        <path d="M47.6471 0H19.0588C15.5647 0 12.7059 2.88 12.7059 6.4V44.8C12.7059 48.32 15.5647 51.2 19.0588 51.2H47.6471C51.1412 51.2 54 48.32 54 44.8V6.4C54 2.88 51.1412 0 47.6471 0ZM47.6471 44.8H19.0588V6.4H47.6471V44.8ZM0 41.6V35.2H6.35294V41.6H0ZM0 24H6.35294V30.4H0V24ZM22.2353 57.6H28.5882V64H22.2353V57.6ZM0 52.8V46.4H6.35294V52.8H0ZM6.35294 64C2.85882 64 0 61.12 0 57.6H6.35294V64ZM17.4706 64H11.1176V57.6H17.4706V64ZM33.3529 64V57.6H39.7059C39.7059 61.12 36.8471 64 33.3529 64ZM6.35294 12.8V19.2H0C0 15.68 2.85882 12.8 6.35294 12.8Z" fill="var(--text-color)"/>
                                    </svg>
                                </button>
                                <button type="button" class="close" onclick="onDeleteShift(this)">
                                    <svg class="icon" viewBox="0 0 96 96" fill="none" xmlns="http://www.w3.org/2000/svg">
                                        <path d="M18.8281 13.1719L13.1719 18.8281L42.3438 48L13.1719 77.1719L18.8281 82.8281L48 53.6562L77.1719 82.8281L82.8281 77.1719L53.6562 48L82.8281 18.8281L77.1719 13.1719L48 42.3438L18.8281 13.1719Z" fill="var(--text-color)"/>
                                    </svg>
                                </button>
                            </div>
                        </div>
                        <div class="shift">
                            <input type="time" name="from" class="input" value="{{if $from}}{{index $from $index}}{{end}}" />
                            <p>Au</p>
                            <input type="time" name="to" class="input" value="{{if $to}}{{index $to $index}}{{end}}" />
                        </div>
                    </div>
                {{end}}
                <button type="submit" class="btn btn-primary">Enregistrer</button>
            </form>
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
    VerifyToken(r, w, &user)
    var payload model.SchedulePayload
    var schedule model.Etablishment = model.Etablishment{UserId: user.Id}
    if err := ReadJsonBody(r.Body, &payload); err != nil{
        DisplayNotification(Notitification{Title: "Error", Message: "Mise a jour echoué", Type: "error"}, w)
        return
    }
    if err := schedule.UpdateSchedule(payload.EtablishmentSchedule, payload.Id); err != nil{
        DisplayNotification(Notitification{Title: "Error", Message: "Mise a jour imposible", Type: "error"}, w)
        return
    }
    DisplayNotification(Notitification{Title: "Reussi", Message: "Mise a jour reussi", Type: "success"}, w)
}
