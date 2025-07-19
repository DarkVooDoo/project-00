package route

import (
	"log"
	"net/http"
	"planify/model"
)

type ProParametreRoute struct{
    User model.UserClaim
    Etablishment model.Etablishment
    Category []model.KeyValue
}

var ParametreHandler http.Handler = ProParametreRoute{}

func (p ProParametreRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
		case http.MethodPost:
			p.Post(w, r)
        default:
            p.Get(w, r)
    }
}

func (p ProParametreRoute) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &p.User); err != nil{
		w.Header().Add("HX-Redirect", "/connexion")
        return
    }
    conn := model.GetDBPoolConn()
    defer conn.Close()

    etablishment := model.Etablishment{Id: p.User.Etablishment, UserId: p.User.Id}
    p.Category = model.Categorys(conn)
    if err := etablishment.Parametre(conn); err != nil{
        w.Header().Add("Location", "/etablissement")
        http.Error(w, "no content", http.StatusNoContent)
        return
    }
    p.Etablishment = etablishment
    if err := CreatePage(p, w, "view/parametre.tmpl"); err != nil{
        log.Printf("error creating the page: %s", err)
    }
}

func(p ProParametreRoute) Post(w http.ResponseWriter, r *http.Request){
	if err := VerifyToken(r, w, &p.User); err != nil{
		log.Printf("error verifying the token: %s", err)
		DisplayNotification(Notitification{"Error", "Vous n'etes pas connecté", "error"}, w)
		return
	}
	etablishment := model.Etablishment{Id: p.User.Etablishment}
	if err := ReadJsonBody(r.Body, &etablishment); err != nil{
		log.Printf("error decoding the json")
		DisplayNotification(Notitification{"Error", "Requete echoué", "error"}, w)
		return
	}
	conn := model.GetDBPoolConn()
	defer conn.Close()
	if err := etablishment.UpdateParametre(conn); err != nil{
		log.Printf("error updating the etablish parametres: %s", err)
		DisplayNotification(Notitification{"Error", "Impossible de mettre a jours les parametres", "error"}, w)
		return
	}
	DisplayNotification(Notitification{"Reussi", "Parametre mise a jour", "success"}, w)
}
