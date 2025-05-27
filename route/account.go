package route

import (
	"html/template"
	"log"
	"net/http"
	"planify/model"
)

type Account struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Profile model.User
}

var AccountHandler http.Handler = &Account{}

func (a Account) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
        case http.MethodPost:
            a.Post(w, r)
        case http.MethodPatch:
            a.Patch(w, r)
        default:
            a.Get(w, r)
    }
}

func (a Account) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &a.User); err != nil{
        w.Header().Add("Location", "/")
        w.WriteHeader(http.StatusTemporaryRedirect)
    }
	conn := model.GetDBPoolConn()
	defer conn.Close()
	a.Navbar = model.GetNavbarFromCache(conn, a.User)
    var user = model.User{Id: a.User.Id}
    if err := user.Profile(conn); err != nil{
        w.Header().Add("Location", "/")
        w.WriteHeader(http.StatusTemporaryRedirect)
    }
    a.Profile = user
    if err := CreatePage(a, w, "view/page.html", "view/account.tmpl"); err != nil{
        log.Printf("error creating page")
        return
    }
}

func (a Account) Patch(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &a.User); err != nil{
        w.Header().Add("HX-Redirect", "/")
        return
    }
    file, fileHeader, err:= r.FormFile("picture")
    if err != nil{
        DisplayNotification(Notitification{"Error", "Requete echoué", "error"}, w)
        return
    }

    if fileHeader.Size > 30000{
        DisplayNotification(Notitification{"Error", "Photo trop grand", "error"}, w)
        return
    }
    user := model.User{Id: a.User.Id}
    if err = user.UploadPhoto(file, fileHeader.Header.Get("Content-Type")); err != nil{
        DisplayNotification(Notitification{"Error", "Requete echoué impossible de enregistrer", "error"}, w)
    }
    temp, err := template.New("picture").Parse(`<img src="{{.}}" class="display" />`)
    if err != nil{
        log.Printf("error creating the template: %s", err)
        return
    }
    if err = temp.Execute(w, user.Picture); err != nil{
        log.Printf("error executing the template: %s", err)
        return
    }
    DisplayNotification(Notitification{"Reussi", "Photo enregistrer", "success"}, w)

    
}

func (a Account) Post(w http.ResponseWriter, r *http.Request){
    VerifyToken(r, w, &a.User)
    var u model.User = model.User{Id: a.User.Id}
    if err := ReadJsonBody(r.Body, &u); err != nil{
        DisplayNotification(Notitification{Title: "Error", Message: "Payload error", Type: "error"}, w)
        return
    }
    if err := u.Modify(); err !=  nil{
        DisplayNotification(Notitification{Title: "Error", Message: "Error dans la mise a jour", Type: "error"}, w)
        return
    }
    DisplayNotification(Notitification{Title: "Success", Message: "Mise a jour", Type: "success"}, w)
}
