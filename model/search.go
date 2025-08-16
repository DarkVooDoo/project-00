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
    resultRows, err := conn.QueryContext(context.Background(), `
		WITH derniers_etabs AS (
		    SELECT e.id, e.name, e.adresse, e.postal, e.phone, e.category_id, e.geolocation FROM etablishment AS e LEFT JOIN category AS c 
			ON c.id=e.category_id WHERE GeolocationDistance(POINT($2,$3), e.geolocation) < $4 AND search_vector @@ websearch_to_tsquery('french', $1) 
			OR GeolocationDistance(POINT($2,$3), e.geolocation) < $4 AND c.category_vector @@ websearch_to_tsquery('french', $1)  LIMIT 5
		),
		maintenant AS (
		    SELECT EXTRACT(ISODOW FROM NOW()) - 1 AS jour_semaine, NOW()::time AS heure_actuelle
		),
		horaires AS (
		    SELECT 
		        e.id,
		        e.name,
		        e.adresse,
				e.postal,
				e.phone,
				c.name AS category,
				e.geolocation[0] AS lat,
				e.geolocation[1] AS lon,
		        oh.day,
		        oh.open_time,
		        oh.close_time,
				-- détermine si l'établissement est ouvert maintenant
		        CASE 
		            WHEN oh.day = EXTRACT(ISODOW FROM NOW()) - 1
		                 AND NOW()::time BETWEEN oh.open_time AND oh.close_time 
		            THEN TRUE 
		            ELSE FALSE 
		        END AS est_ouvert,
		        -- calcule le prochain horaire d'ouverture après maintenant
		        ((oh.day - EXTRACT(DOW FROM NOW()) + 7) % 7) AS jours_d_attente
		    FROM derniers_etabs e
		    LEFT JOIN schedule oh ON e.id = oh.etablishment_id
			LEFT JOIN category AS c ON c.id = e.category_id
		),
		prochains_horaires AS (
		    SELECT *,
		        CASE 
					WHEN est_ouvert THEN CONCAT('Ferme à ', TO_CHAR(close_time, 'HH24:MI'))
					WHEN open_time IS NULL THEN 'Fermé Termporairement'
					ELSE CONCAT('Ouvre ', ('{Lun,Mar,Mer,Jeu,Ven,Sam,Dim}'::TEXT[])[day + 1], ' ' , TO_CHAR(open_time, 'HH24:MI'))
				END AS prochain_horaire,
		        ROW_NUMBER() OVER (PARTITION BY id ORDER BY jours_d_attente) AS rn
		    FROM horaires
		)
		SELECT id, name, adresse, postal, COALESCE(phone, ''), category, lat, lon, CASE WHEN est_ouvert THEN 'Ouvert' ELSE 'Fermé' END, prochain_horaire FROM prochains_horaires
		WHERE rn = 1 
		GROUP BY id, name, adresse, postal, phone, category, lat, lon, prochain_horaire, est_ouvert`, query, lat, lon, radius)

    if err != nil{
        log.Printf("error in the query: %s", err)
        return searchList
    }
    for resultRows.Next(){
        if err := resultRows.Scan(&etablishment.Id, &etablishment.Name, &etablishment.Adresse, &etablishment.Postal, &etablishment.Phone, &etablishment.Category, 
		&etablishment.Lat, &etablishment.Lon, &etablishment.IsOpen, &etablishment.NextShift); err != nil{
            log.Printf("error scanning rows: %s", err)
        }
        searchList = append(searchList, etablishment)
    }
    return searchList
}
