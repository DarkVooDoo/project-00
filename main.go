package main

import (
	"log"
	"net/http"
	"planify/model"
	"planify/route"
	"time"
)

const (
    Addr = ":8000"
)

func main(){
    
    if err := model.InitDB(); err != nil{
        log.Fatalf("error db init\n %v", err)
    }   
    mux := http.NewServeMux()
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
    mux.HandleFunc("/", route.LandpageHandler.ServeHTTP)
    mux.HandleFunc("/connexion",  route.SigninHandler.ServeHTTP)
    mux.HandleFunc("/compte", route.AccountHandler.ServeHTTP)
    mux.HandleFunc("/recherche", route.SearchHandler.ServeHTTP)
    mux.HandleFunc("/rendez-vous", route.AppointmentHandler.ServeHTTP)
    mux.HandleFunc("/rendez-vous/{id}", route.ViewAppointmentHandler.ServeHTTP)
    mux.HandleFunc("/etablissement/creer", route.NewEtablishmentHandler.ServeHTTP)
    mux.HandleFunc("/etablissement/{id}", route.StoreHandler.ServeHTTP)
    mux.HandleFunc("/etablissement/{id}/rendez-vous", route.EtablishmentAppointmentHandler.ServeHTTP)
    mux.HandleFunc("/etablissement/{id}/rendez-vous/nouveau", route.NewAppointmentHandler.ServeHTTP)
    mux.HandleFunc("/planning", route.PlanningHandler.ServeHTTP)
    mux.HandleFunc("/schedule", route.ScheduleHandler.ServeHTTP)
    mux.HandleFunc("/etablissement", route.MyEtablishmentHandler.ServeHTTP)
    mux.HandleFunc("/etablissement/parametre", route.ParametreHandler.ServeHTTP)
    mux.HandleFunc("/etablissement/service", route.ServiceHandler.ServeHTTP)
    mux.HandleFunc("/etablissement/employee", route.EtablishmentEmployeHandler.ServeHTTP)
    mux.HandleFunc("/employee/{id}", route.EmployeHandler.ServeHTTP)

    server := http.Server{
        Addr: Addr,
        ReadTimeout: time.Second * 5,
        Handler: mux,
    }

    if err := server.ListenAndServe(); err != nil{
        log.Fatalf("error spinning server: %s", Addr)
    }
}
