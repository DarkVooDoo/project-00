package route

import (
	"net/http"
	"planify/model"
)

type Store struct{
    User model.UserClaim
    Etablishment model.Etablishment
    DayIndex int
}

const (
    EntrepriseCookie = "eid"
)

var StoreHandler http.Handler = &Store{}

func (s Store) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
        default:
            s.Get(w, r)
    }
}

func (s Store) Get(w http.ResponseWriter, r *http.Request){
    VerifyToken(r, w, &s.User)
    s.Etablishment.Id = r.PathValue("id")
    
    weekDay, err := s.Etablishment.Public()
    if err != nil{
        w.Header().Add("Location", "/")
        w.WriteHeader(http.StatusTemporaryRedirect)
        return
    }
    s.DayIndex = weekDay
    CreatePage(s, w, "view/page.html", "view/etablissement.tmpl")
}

