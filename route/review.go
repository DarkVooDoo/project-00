package route

import (
	"html/template"
	"log"
	"net/http"
	"planify/model"
	"strconv"
)

type ReviewRoute struct{

}

var ReviewHandler http.Handler = &ReviewRoute{}

func (rr ReviewRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){

	switch r.Method{
		case http.MethodPost:
			rr.Post(w, r)
		case http.MethodDelete:
			rr.Delete(w, r)
		default:
			log.Println("Deleted")
	}
}

func(rr ReviewRoute) Post(w http.ResponseWriter, r *http.Request){
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err !=  nil{
		log.Printf("unauthorized")
		return
	}
	conn := model.GetDBPoolConn()
	defer conn.Close()
	id, _ := strconv.Atoi(r.PathValue("id"))
	review := model.Review{Id: int64(id), UserId: user.Id}
	if err := ReadJsonBody(r.Body, &review); err != nil{
		log.Printf("error reading the payload: %s", err)
		return
	}
	if err := review.Update(conn); err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	temp, err := template.New("success").Parse(`
        <div class="success-message" id="successMessage">
            <svg class="success-icon" viewBox="0 0 24 24" fill="none" stroke="var(--success-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                <polyline points="22 4 12 14.01 9 11.01"></polyline>
            </svg>
            <h3 class="success-title">Merci pour votre avis !</h3>
            <p class="success-text">Votre retour nous aide à améliorer nos services et à offrir la meilleure expérience possible.</p>
            <button class="btn btn-primary" onclick="closeReviewModal()">Fermer</button>
        </div>
	`)
	if err != nil{
		log.Printf("error parsing the template: %s", err)
		return
	}
	if err := temp.Execute(w, nil); err != nil{
		log.Printf("error executing the template: %s", err)
	}
	
}

func (rr ReviewRoute) Delete(w http.ResponseWriter, r *http.Request){
	var user model.UserClaim
	if err := VerifyToken(r, w, &user); err != nil{
		w.Header().Add("HX-Redirect", "/connexion")
		return
	}
	conn := model.GetDBPoolConn()
	defer conn.Close()
	id, _ := strconv.Atoi(r.PathValue("id"))
	review := model.Review{Id: int64(id), UserId: user.Id}
	if err := review.Delete(conn); err != nil{
		log.Printf("error deleting the review: %s", err)
	}
}
