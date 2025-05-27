package route

import (
	"log"
	"net/http"
	"planify/model"
)

type Search struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Etablishment []model.Etablishment
}

var SearchHandler http.Handler = &Search{}

func (s Search) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
        case http.MethodPost:
            s.Post(w, r)
        default:
            s.Get(w, r)
    }
}

func (s Search) Get(w http.ResponseWriter, r *http.Request){
	if err := VerifyToken(r, w, &s.User); err == nil{
		conn := model.GetDBPoolConn()
		defer conn.Close()
		s.Navbar = model.GetNavbarFromCache(conn, s.User)
	}
    query := r.URL.Query().Get("query")
    location := r.URL.Query().Get("location")
    lat := r.URL.Query().Get("lon")
    lon := r.URL.Query().Get("lat")

    log.Printf("Text: %s\tLocation: %s\tLatitude: %s\tLongitude: %s", query, location, lat, lon)
    s.Etablishment = model.SearchEtablishment(query)

    CreatePage(s, w, "view/page.html", "view/search.tmpl", "view/component/EtablishmentCard.tmpl")
}

func (s Search) Post(w http.ResponseWriter, r *http.Request){


}
