package route

import (
	"net/http"
	"planify/model"
	"strconv"
)

type Store struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Etablishment model.Etablishment
    Week []string
}

var StoreHandler http.Handler = &Store{}

func (s Store) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
        default:
            s.Get(w, r)
    }
}

func (s Store) Get(w http.ResponseWriter, r *http.Request){
	conn := model.GetDBPoolConn()
	defer conn.Close()
	if err := VerifyToken(r, w, &s.User); err == nil{
		s.Navbar = model.GetNavbarFromCache(conn, s.User)
	}
    
	etablishmentId, _ := strconv.Atoi(r.PathValue("id"))
    s.Etablishment.Id = int64(etablishmentId)

    _, err := s.Etablishment.Public(conn)
    if err != nil{
        w.Header().Add("Location", "/")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    s.Week = model.Week
	CreatePage(s, w, "view/page.html", "view/etablissement.tmpl")
}

