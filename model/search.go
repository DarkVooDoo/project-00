package model

import (
	"context"
	"log"
)

func SearchEtablishment(query string, lat float64, lon float64, radius int)[]Etablishment{

    conn := GetDBPoolConn()
    defer conn.Close()

    var searchList []Etablishment
    var etablishment Etablishment
    resultRows, err := conn.QueryContext(context.Background(), `SELECT e.id, e.name, adresse, postal, COALESCE(e.phone, 'N/A'), c.name, 
    CONCAT(e.schedule->'from'->>EXTRACT(ISODOW FROM NOW())::int - 1, ' - ', e.schedule->'to'->>EXTRACT(ISODOW FROM NOW())::int - 1), e.geolocation[0], e.geolocation[1] 
    FROM etablishment AS e LEFT JOIN category AS c ON e.category_id=c.id WHERE search_vector @@ websearch_to_tsquery('french', $1) OR 
	c.category_vector @@ websearch_to_tsquery('french', $1) AND GeolocationDistance(POINT($2,$3),e.geolocation) < $4 LIMIT 5`, query, lat, lon, radius)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return searchList
    }
    for resultRows.Next(){
        if err =resultRows.Scan(&etablishment.Id, &etablishment.Name, &etablishment.Adresse, &etablishment.Postal, &etablishment.Phone, &etablishment.Category, &etablishment.TodaySchedule,
    &etablishment.Lat, &etablishment.Lon); err != nil{
            log.Printf("error scanning the row: %s", err)
        }
        if len(etablishment.TodaySchedule) < 4 {etablishment.TodaySchedule = "Fermé"}
        searchList = append(searchList, etablishment)
    }
    return searchList
}
