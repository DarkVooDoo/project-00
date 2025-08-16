package route

import (
	"html/template"
	"log"
	"net/http"
	"planify/model"
)

type ServiceRoute struct{

}

var ServiceHandler http.Handler = ServiceRoute{}

func (s ServiceRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
        case http.MethodPost:
            s.Post(w, r)
        case http.MethodPut:
            s.Put(w, r)
        case http.MethodDelete:
            s.Delete(w, r)
        default:
            s.Get(w, r)
    }
}

func (s *ServiceRoute) Get(w http.ResponseWriter, r *http.Request){

    var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		w.Header().Add("HX-Redirect", "/connexion")
		return
	}
    conn := model.GetDBPoolConn()
    defer conn.Close()
    etablishmentList := model.Service{EtablishmentId: user.Etablishment}
    list, _ := etablishmentList.GetList(conn)
    serviceTemplate, err := template.ParseGlob("view/Service.tmpl")

    if err != nil{
        log.Printf("error parsing service: %s", err)
        return
    }
    if err := serviceTemplate.Execute(w, list); err != nil{
        log.Println(err)
    }

}

func (s *ServiceRoute) Post(w http.ResponseWriter, r *http.Request){
    var user model.UserClaim
    if err := VerifyToken(r, w, &user); err != nil{
        DisplayNotification(Notitification{"Error", "Requete echouée", "error"}, w)
        return
    }

    service := model.Service{EtablishmentId: user.Etablishment}
    ReadJsonBody(r.Body, &service)
    if err := service.Create(); err != nil{
        DisplayNotification(Notitification{"Error", "Requete echouée", "error"}, w)
        return
    }
    DisplayNotification(Notitification{"Reussi", "Service crée", "success"}, w)
    temp, err := template.New("service").Parse(`
		<form class="service" hx-put="/etablissement/service" hx-swap="none" hx-ext="json-enc-custom" hx-vals='{"id": "{{.Id}}"}'>
		  <div class="service-header">
		    <h1 class="header-title">Service</h1>
            <div class="command">
                <button type="button" class="btn btn-danger" hx-delete="/etablissement/service" hx-vals='{"id": "{{.Id}}"}' hx-swap="delete" hx-target="closest .service">
                <svg class="icon" viewBox="0 0 24 24" fill="none">
                    <path d="M4 6H20M16 6L15.7294 5.18807C15.4671 4.40125 15.3359 4.00784 15.0927 3.71698C14.8779 3.46013 14.6021 3.26132 14.2905 3.13878C13.9376 3 13.523 3 12.6936 3H11.3064C10.477 3 10.0624 3 9.70951 3.13878C9.39792 3.26132 9.12208 3.46013 8.90729 3.71698C8.66405 4.00784 8.53292 4.40125 8.27064 5.18807L8 6M18 6V16.2C18 17.8802 18 18.7202 17.673 19.362C17.3854 19.9265 16.9265 20.3854 16.362 20.673C15.7202 21 14.8802 21 13.2 21H10.8C9.11984 21 8.27976 21 7.63803 20.673C7.07354 20.3854 6.6146 19.9265 6.32698 19.362C6 18.7202 6 17.8802 6 16.2V6M14 10V17M10 10V17" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                </button>
                <button type="submit" class="btn btn-primary">
                <svg class="icon" viewBox="0 0 24 24" fill="none">
                    <path d="M21 3V8M21 8H16M21 8L18 5.29168C16.4077 3.86656 14.3051 3 12 3C7.02944 3 3 7.02944 3 12C3 16.9706 7.02944 21 12 21C16.2832 21 19.8675 18.008 20.777 14" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                </button>
            </div>
		  </div>
		  <div class="body">
			<div class="form-group">
		      <div class="field">
		        <label for="name" class="form-label">Nom</label>
		        <input type="text" class="form-input" id="name" value="{{.Name}}" autocomplete="off" />
		      </div>
		      <div class="field">
		        <label for="duration" class="form-label">Duration (min)</label>
		        <input type="number" class="form-input" id="duration" value="{{.Duration}}" autocomplete="off" />
		      </div>
		      <div class="doubleField">
		        <div class="field">
		          <label for="price" class="form-label">Prix</label>
		          <input type="number" class="form-input" id="price" value="{{.Price}}" autocomplete="off" />
		        </div>
		        <div class="field">
		          <label for="solde" class="form-label">Solde (%)</label>
		          <input type="number" class="form-input" id="solde" min="O" max="100" value="{{.Discount}}" />
		        </div>
		      </div>
		    </div>
		    <div class="form-group">
		      <label for="description" class="form-label">Description</label>
		      <textarea id="description" maxlength="150">{{.Description}}</textarea>
		    </div>
		  </div>
		</form>
    `)
    if err != nil{
        log.Printf("error parsing the template: %s", err)
        return
    }
    temp.Execute(w, service)
}

func (s *ServiceRoute) Put(w http.ResponseWriter, r *http.Request){

	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		log.Printf("error user unauthorized: %s", err)
		DisplayNotification(Notitification{"Echoué", "Vous n'etes pas connecté", "error"}, w)
		return
	}

    service := model.Service{EtablishmentId: user.Etablishment}
    if err := ReadJsonBody(r.Body, &service); err != nil{
        log.Println("error reading the payload")
        DisplayNotification(Notitification{"Error", "Mise a jour echouée", "error"}, w)
        return
    }
    if err := service.Update(); err != nil{
        log.Println("error updating")
        DisplayNotification(Notitification{"Error", "Mise a jour echouée", "error"}, w)
        return
    }
    DisplayNotification(Notitification{"Reussi", "Mise a jour reussi", "success"}, w)
}


func (s *ServiceRoute) Delete(w http.ResponseWriter, r *http.Request){
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		log.Printf("error unauthorized: %s", err)
		w.Header().Add("HX-Redirect", "/connexion")
		return
	}
	service := model.Service{EtablishmentId: user.Etablishment}
	ReadJsonBody(r.Body, &service)
    if err := service.Delete(); err != nil{
        DisplayNotification(Notitification{"Error", "requete echoée", "error"}, w)
        return
    }
    DisplayNotification(Notitification{"Reussi", "Service crée", "success"}, w)
}
