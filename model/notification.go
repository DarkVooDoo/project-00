package model

import "database/sql"

type Notification struct{
	Id string
	Message string
	Timestamp string
	Path string
	UserPhoto string
	UserId int64
	EtablishmentId int64
	EmployeeId int64
}

func (n Notification) Create(conn *sql.Conn)error{
	
	return nil
}
