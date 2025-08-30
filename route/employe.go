package route

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"planify/model"
	"strconv"
	"strings"
)

type EtablishmentEmployeeRoute struct{
    Week []string
    EmployeeList  []model.Employe
}

var EtablishmentEmployeHandler http.Handler = EtablishmentEmployeeRoute{}

func (e EtablishmentEmployeeRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        case http.MethodDelete:
            e.Delete(w, r)
        case http.MethodPatch:
            e.Patch(w, r)
        case http.MethodPut:
            e.Put(w, r)
        case http.MethodPost:
            e.Post(w, r)
        default:
            e.Get(w, r)
    }
}

func (e EtablishmentEmployeeRoute) Get(w http.ResponseWriter, r *http.Request){

    var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		w.Header().Add("HX-Redirect", "/")
		return
	}
    conn := model.GetDBPoolConn()
    defer conn.Close()

    employe := model.Employe{EtablishmentId: user.Etablishment}
    e.Week = []string{"Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi", "Dimanche"}
    e.EmployeeList = employe.GetEtablishmentEmployees(conn)
    temp, err := template.New("employe").Parse(`
        {{$week := .Week}}
        <form id="new-employe" hx-target=".employee-list" hx-post="/etablissement/employee" hx-swap="beforeend" hx-trigger="keyup[keyCode==13] throttle:5000ms" 
        hx-on::after-request="if(event.detail.successful && event.detail.elt.id != 'email')this.reset()"  >
            <input type="email" class="input" id="email" name="email" autocomplete="off" placeholder="Email" hx-put="/etablissement/employee" hx-trigger="keyup changed delay:1000ms" 
            hx-target="#employe-sugg" hx-swap="outerHTML" />
            <input type="text" id="id" name="id" class="hidden" />
            <div id="employe-sugg" class="hidden"></div>
        </form>
		<div class="employee-list">
            {{range .EmployeeList}}
                {{$from := .Schedule.From}}
                {{$to := .Schedule.To}}
                <div class="employee-card">
                    {{if .Picture}}
                        <img src="/static/clock.svg" class="picture" />
                    {{else}}
                        <div class="picture" style="border:1px solid var(--border-color);">{{.ShortName}}</div>
                    {{end}}
                    <h1 class="name">{{.Name}}</h1>
					<h2 class="more">Barber</h2>
					<span class="more">Ancienneté: {{.Joined}}</span>
                    <div class="actions">
                        <button type="button" class="actionBtn btn-outline" popovertarget="popover{{.Id}}">Horaire</button>
                        <button type="button" class="actionBtn btn-danger" popovertarget="confirmation{{.Id}}">Supprimer</button>
						<div id="confirmation{{.Id}}" class="confirmation" popover>
						    <div class="confirmation-header">
						        <div class="header-icon">
						            <svg viewBox="0 0 24 24" fill="none" class="icon">
						                <g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M12 15H12.01M12 12V9M4.98207 19H19.0179C20.5615 19 21.5233 17.3256 20.7455 15.9923L13.7276 3.96153C12.9558 2.63852 11.0442 2.63852 10.2724 3.96153L3.25452 15.9923C2.47675 17.3256 3.43849 19 4.98207 19Z" stroke="rgb(220 38 38)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> </g>
						            </svg>
						        </div>
						        <h3 class="header-title">Confirmation de suppression</h3>
						        <button type="button" class="header-icon closeBtn btn-outline" popovertarget="confirmation{{.Id}}" popovertargetaction="hide">
						            <svg viewBox="0 0 24 24" fill="none" class="icon"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <g id="Menu / Close_SM"> <path id="Vector" d="M16 16L12 12M12 12L8 8M12 12L16 8M12 12L8 16" stroke="var(--text-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> </g> </g></svg>
						
						        </button>
						    </div>
						    <div class="confirmation-body">
						        <p>Vous êtes sur le point de supprimer définitivement l'employé suivant :</p>
						        <div class="confirmation-employee">
						            <div class="shortname">{{.ShortName}}</div>
						            <div class="employee-data">
						                <h4>{{.Name}}</h4>
						                <p>Barber</p>
						                <p>Ancienneté: {{.Joined}}</p>
						            </div>
						        </div>
						    </div>
						    <div class="confirmation-footer">
						        <button type="button" class="btn btn-danger" popovertarget="confirmation{{.Id}}" popovertargetaction="hide">Cancel</button>
						        <button type="button" class="btn btn-primary" hx-delete="/etablissement/employee" hx-vals='{"id": "{{.Id}}"}' hx-swap="delete" 
								hx-target="closest .employee-card">Confirmer</button>
						    </div>
						</div>
                    </div>
                    <form class="employee-schedule" id="popover{{.Id}}" hx-patch="/etablissement/employee" hx-swap="none" hx-ext="json-enc-custom" hx-vals='{"id": "{{.Id}}"}' popover>
                        {{range $index, $element := $week}}
                            <div class="day">
                                <h1 class="label">{{$element}}</h1>
                                <input type="time" name="from" class="input" value="{{if $from}}{{index $from $index}}{{end}}" />
                                <p>Au</p>
                                <input type="time" name="to" class="input" value="{{if $to}}{{index $to $index}}{{end}}" />
                                <button type="button" class="resetBtn" onclick="onDeleteShift(this)">
                                    <svg class="icon" viewBox="0 0 96 96" fill="none">
                                        <path d="M18.8281 13.1719L13.1719 18.8281L42.3438 48L13.1719 77.1719L18.8281 82.8281L48 53.6562L77.1719 82.8281L82.8281 77.1719L53.6562 48L82.8281 18.8281L77.1719 13.1719L48 42.3438L18.8281 13.1719Z" fill="var(--text-color)"/>
                                    </svg>
                                </button>
                            </div>
                        {{end}}
                        <button type="submit" class="btn">Enregistrer</button>
                    </form>
                </div>
            {{end}}
		</div>
    `)
    if err != nil{
        log.Printf("error creating the template:  %s", err)
        return
    }
    if err := temp.Execute(w, e); err != nil{
        log.Printf("error executing template: %s", err)
    }
}

func (e EtablishmentEmployeeRoute) Put(w http.ResponseWriter, r *http.Request){
    if r.FormValue("email") == ""{
        return
    }
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		log.Printf("error verifying token: %s", err)
		return
	}
    employee := model.Employe{Email: r.FormValue("email"), EtablishmentId: user.Etablishment}
    employeeList := employee.SuggestEmployee()
    if len(employeeList) > 0{
        temp, err := template.New("suggest").Parse(`
        <div id="employe-sugg">
            {{range .}}
                <button type="button" data-id="{{.UserId}}" onclick="onEmployeSelected(this)" class="proposal">{{.Email}}</button>
            {{end}}
        </div>
        `)
        if err != nil{
            log.Printf("error parsing the template: %s", err)
            return
        }
        if err = temp.Execute(w, employeeList); err != nil{
            log.Printf("error executing template: %s", err)
            return
        }
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (e EtablishmentEmployeeRoute) Patch(w http.ResponseWriter, r *http.Request){

    var schedule model.SchedulePayload
    var employe model.Employe
    if err := ReadJsonBody(r.Body, &schedule); err != nil{
        log.Printf("error reading the json: %s", err)
        return
    }
    if err := employe.UpdateSchedule(schedule); err != nil{
        DisplayNotification(Notitification{"Error", "requete echoué", "error"}, w)
        return
    }
    DisplayNotification(Notitification{"Success", "requete reussi", "success"}, w)
}

func (e EtablishmentEmployeeRoute) Post(w http.ResponseWriter, r *http.Request){
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
        DisplayNotification(Notitification{"Error", "Requete echoué", "error"}, w)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil{
		log.Printf("error converting the id to integer: %s", err)
		return
	}
    employee := model.Employe{UserId: int64(id), EtablishmentId: user.Etablishment}
	if err := employee.New(); err != nil{
        DisplayNotification(Notitification{"Error", "Requete echoué", "error"}, w)
        return
    }
    var newEmployee struct{
        Employee model.Employe 
        Week []string
    } = struct{Employee model.Employe; Week []string}{employee, model.Week}
    temp, err := template.New("new-employe").Parse(`
        {{$week := .Week}}
        <div class="employee-card">
            {{if .Employee.Picture}}
                <img src="/static/clock.svg" class="picture" />
            {{else}}
                <div class="picture" style="border:1px solid var(--border-color);">{{.Employee.ShortName}}</div>
            {{end}}
            <h1 class="name">{{.Employee.Name}}</h1>
			<h2 class="more">Barber</h2>
			<span class="more">Ancienneté: {{.Employee.Joined}}</span>
            <div class="actions">
                <button type="button" class="actionBtn btn-outline" popovertarget="popover{{.Employee.Id}}">Horaire</button>
                <button type="button" class="actionBtn btn-danger" popovertarget="confirmation{{.Employee.Id}}">Supprimer</button>
                <div id="confirmation{{.Employee.Id}}" class="confirmation" popover>
				    <div class="confirmation-header">
				        <div class="header-icon">
				            <svg viewBox="0 0 24 24" fill="none" class="icon">
				                <g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <path d="M12 15H12.01M12 12V9M4.98207 19H19.0179C20.5615 19 21.5233 17.3256 20.7455 15.9923L13.7276 3.96153C12.9558 2.63852 11.0442 2.63852 10.2724 3.96153L3.25452 15.9923C2.47675 17.3256 3.43849 19 4.98207 19Z" stroke="rgb(220 38 38)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> </g>
				            </svg>
				        </div>
				        <h3 class="header-title">Confirmation de suppression</h3>
				        <button type="button" class="header-icon closeBtn btn-outline" popovertarget="confirmation{{.Employee.Id}}"  popovertargetaction="hide">
				            <svg viewBox="0 0 24 24" fill="none" class="icon"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g><g id="SVGRepo_iconCarrier"> <g id="Menu / Close_SM"> <path id="Vector" d="M16 16L12 12M12 12L8 8M12 12L16 8M12 12L8 16" stroke="var(--text-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"></path> </g> </g></svg>
				        </button>
				    </div>
				    <div class="confirmation-body">
				        <p>Vous êtes sur le point de supprimer définitivement l'employé suivant :</p>
				        <div class="confirmation-employee">
				            <div class="shortname">{{.Employee.ShortName}}</div>
				            <div class="employee-data">
				                <h4>{{.Employee.Name}}</h4>
				                <p>Barber</p>
				                <p>Ancienneté: {{.Employee.Joined}}</p>
				            </div>
				        </div>
				    </div>
				    <div class="confirmation-footer">
				        <button type="button" class="btn btn-danger" popovertarget="confirmation{{.Employee.Id}}" popovertargetaction="hide">Cancel</button>
						<button type="button" class="btn btn-primary" hx-delete="/etablissement/employee" hx-vals='{"id": "{{.Employee.Id}}", "active": "false"}' hx-swap="delete" 
						hx-target="closest .employee-card">Confirmer</button>
				    </div>
				</div>
            </div>
            <form class="employee-schedule" id="popover{{.Employee.Id}}" hx-patch="/etablissement/employee" hx-swap="none" hx-ext="json-enc-custom" hx-vals='{"id": "{{.Employee.Id}}"}' popover>
                {{range $index, $element := $week}}
                    <div class="day">
                        <h1 class="label">{{$element}}</h1>
                        <input type="time" name="from" class="input"/>
                        <p>Au</p>
                        <input type="time" name="to" class="input" />
                    </div>
                {{end}}
                <button type="submit" class="btn">Enregistrer</button>
            </form>
        </div>
    `)
    if err != nil{
        log.Printf("error parsing the template: %s", err)
        DisplayNotification(Notitification{"Error", "Requete echoué", "error"}, w)
        return
    }
    if err = temp.Execute(w, newEmployee); err != nil{
        log.Printf("error executing the template: %s", err)
        DisplayNotification(Notitification{"Error", "Requete echoué", "error"}, w)
        return
    }
    DisplayNotification(Notitification{"Reussi", "Employée Ajouté", "success"}, w)
}

func (e EtablishmentEmployeeRoute) Delete(w http.ResponseWriter, r *http.Request){
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		log.Printf("error in the token")
		return
	}
	valMap := make(map[string]string)
    body, _ := io.ReadAll(r.Body)
	for v := range strings.SplitSeq(string(body), "&"){
		keyVal := strings.Split(v, "=")
		valMap[keyVal[0]] = keyVal[1]
	}
	id, _ := strconv.Atoi(valMap["id"])
    employee := model.Employe{Id: int64(id), EtablishmentId: user.Etablishment, UserId: user.Id, IsActive: valMap["active"]=="true"}
    if employee.Id == 0{
        DisplayNotification(Notitification{"Error", "requete echoué", "error"}, w) 
        return
    }
    if err := employee.Delete(); err != nil{
         DisplayNotification(Notitification{"Error", "requete echoué", "error"}, w)
        return
	}
	DisplayNotification(Notitification{"Reussi", "employee supprimer", "success"}, w)
}
