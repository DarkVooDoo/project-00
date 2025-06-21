package route

import (
	"encoding/json"
	"log"
	"net/http"
	"planify/model"
	"strconv"
)

type SearchPayload struct{
	id string `json:"id"`
	Latitude float64 `json:"lat,string"`
	Longitude float64 `json:"lon,string"`
	Query string `json:"query"`
	Postal string `json:"postal"`
	Radius int `json:"radius,string"`
}

type Search struct{
    User model.UserClaim
	Navbar model.CacheNavbar
    Etablishment []model.Etablishment
	NavbarData SearchPayload
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
    //location := r.URL.Query().Get("location")
    lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 32)
	if err != nil{
		log.Printf("error no geolocation")
		return
	}
    lon, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 32)
	if err != nil{
		log.Printf("error no geolocation")
		return
	}
    radius, err := strconv.Atoi(r.URL.Query().Get("radius"))
	if err != nil{
		radius = 2
	}
	s.NavbarData.Query = query
	s.NavbarData.Latitude = lat
	s.NavbarData.Longitude = lon
    s.Etablishment = model.SearchEtablishment(query, lat, lon, radius)

    CreatePage(s, w, "view/page.html", "view/search.tmpl", "view/component/EtablishmentCard.tmpl")
}

func (s Search) Post(w http.ResponseWriter, r *http.Request){
	var searchPayload SearchPayload
	if err := ReadJsonBody(r.Body, &searchPayload); err != nil{
		log.Printf("error decoding the json: %s", err)
		return
	}
	s.Etablishment = model.SearchEtablishment(searchPayload.Query, searchPayload.Latitude, searchPayload.Longitude, searchPayload.Radius)
	data, err := json.Marshal(s.Etablishment)
	if err != nil{
		log.Printf("error marshal etablishments: %s", err)
		return
	}
	w.Write(data)
}
