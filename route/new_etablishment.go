package route

import (
	"log"
	"net/http"
	"planify/model"
)

type NewEtablishmentRoute struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Category []model.KeyValue

}

var NewEtablishmentHandler http.Handler = &NewEtablishmentRoute{}

func (e NewEtablishmentRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        case http.MethodPut:
            e.Put(w, r)
        default:
            e.Get(w, r)
    }
}

func (e NewEtablishmentRoute) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &e.User); err != nil{
        w.Header().Add("Location", "/connexion")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }

    conn := model.GetDBPoolConn()
    defer conn.Close()

	e.Navbar = model.GetNavbarFromCache(conn, e.User)
    e.Category = model.Categorys(conn)
    if err := CreatePage(e, w, "view/page.html", "view/new_etablishment.tmpl"); err != nil{
        log.Println(err)
    }
}

func (e NewEtablishmentRoute) Put(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &e.User); err != nil{
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    var newEtablishment model.Etablishment = model.Etablishment{UserId: e.User.Id}
    if err := ReadJsonBody(r.Body, &newEtablishment); err != nil{
        log.Printf("error reading the body: %s", err)
        DisplayNotification(Notitification{Title: "Error", Message: "Payload error", Type: "error"}, w)
        return
    }
    if err := newEtablishment.Create(); err != nil{
        DisplayNotification(Notitification{Title: "Error", Message: "On a pa reussi", Type: "error"}, w)
        return
    }
    w.Header().Add("HX-Redirect", "/etablissement")
}
