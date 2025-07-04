package route

import (
	"log"
	"net/http"
	"planify/model"
)

type ProParametreRoute struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Etablishment model.Etablishment
    Category []model.KeyValue
}

var ParametreHandler http.Handler = ProParametreRoute{}

func (p ProParametreRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){
    switch r.Method{
        default:
            p.Get(w, r)
    }
}

func (p ProParametreRoute) Get(w http.ResponseWriter, r *http.Request){
    if err := VerifyToken(r, w, &p.User); err != nil{
        w.Header().Add("Location", "/connexion")
        http.Error(w, "unauthorized", http.StatusTemporaryRedirect)
        return
    }
    conn := model.GetDBPoolConn()
    defer conn.Close()

	p.Navbar = model.GetNavbarFromCache(conn, p.User)
    etablishment := model.Etablishment{Id: p.User.Etablishment, UserId: p.User.Id}
    p.Category = model.Categorys(conn)
    if err := etablishment.Parametre(conn); err != nil{
        w.Header().Add("Location", "/etablissement")
        http.Error(w, "no content", http.StatusNoContent)
        return
    }
    p.Etablishment = etablishment
    if err := CreatePage(p, w, "view/page.html", "view/parametre.tmpl"); err != nil{
        log.Printf("error creating the page: %s", err)
    }
}
