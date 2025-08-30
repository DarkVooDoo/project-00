package route

import (
	"context"
	"log"
	"net/http"
	"planify/model"
	"slices"
	"strings"
	"unicode"

	"google.golang.org/api/idtoken"
)

type Sign struct{
    User model.UserClaim
}

var SigninHandler http.Handler = Sign{}

func (s Sign) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        case http.MethodPut:
            s.Put(w, r)
        case http.MethodPost:
            s.Post(w, r)
        case http.MethodDelete:
            s.Delete(w, r)
         default:
            s.Get(w, r)
    }

}

func (s Sign) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &s.User); err == nil{
        w.Header().Add("Location", "/")
        w.WriteHeader(http.StatusSeeOther)
    }
    CreatePage(nil, w, "view/page.html", "view/sign.tmpl")
}

func (s Sign) Post(w http.ResponseWriter, r *http.Request){
	googleToken := r.FormValue("cred")
	if googleToken != ""{
		payload, err := idtoken.Validate(context.Background(), googleToken, "432757696898-pbn4r01ut5ejpnrs342foham08ger5rp.apps.googleusercontent.com")
		if err != nil{
			log.Printf("error validating the token: %s", err)
		}
		//TODO: Create a new JWT token with his google info cuz google token duration is 1H
		log.Println(payload.Claims["email"])
		return
	}
    email := r.FormValue("email")
    password := r.FormValue("password")
    
    var user model.User
	err := user.Sign(email, password)
	if err != nil{
        w.WriteHeader(http.StatusForbidden)
        return
    }
    err = model.CreateAccessToken(user.Id, user.ShortName, user.Picture,  user.EtablishmentId, user.EmployeeId, w)
    if err != nil{
        w.WriteHeader(http.StatusForbidden)
        return
    }
    w.Header().Add("HX-Redirect", "/")
}

func (s Sign) Put(w http.ResponseWriter, r *http.Request){
    var user model.User
    if err := ReadJsonBody(r.Body, &user); err != nil{
		DisplayNotification(Notitification{"Error", "Requete echoué payload malformed", "error"}, w)
        return
    }
    if user.Confirmation != user.Password || !isStrongPassword(user.Password){
		DisplayNotification(Notitification{"Error", "Requete echoué", "error"}, w)
        return
    }
    if err := user.Create(); err != nil{
		DisplayNotification(Notitification{"Reussi", "Requete echoué", "error"}, w)
        return
    }
	DisplayNotification(Notitification{"Reussi", "Utilisateur creé", "success"}, w)
    w.WriteHeader(http.StatusOK)
}

func (s Sign) Delete(w http.ResponseWriter, r *http.Request){
    deleteCookie := http.Cookie{
        Name: "access-token",
        MaxAge: -1,
    }
    deleteEtablishment := http.Cookie{
        Name: "e_id",
        MaxAge: -1,
    }
    http.SetCookie(w, &deleteCookie)
    http.SetCookie(w, &deleteEtablishment)
    w.Header().Add("HX-Redirect", "/connexion")
    w.WriteHeader(http.StatusSeeOther)
}

func isStrongPassword(password string)bool{
    pass := []int{0, 0, 0, 0, 0}
    valide := []int{1, 1, 1, 1, 1}
    for _, v := range password{
        switch {
            case unicode.IsUpper(v):
                pass[0] = 1
            case unicode.IsLower(v):
                pass[1] = 1
            case unicode.IsNumber(v):
                pass[2] = 1

        }
        if strings.Contains("!%?$#", string(v)){
            pass[4] = 1
        }
    }
    if len(password) > 7{
        pass[3] = 1
    }
    return slices.Equal(pass, valide)
}
