package db

type User struct {
	ID               uint    `json:"id" gorm:"primary_key"`
	Username         string  `json:"Username" gorm:"type:varchar(200);unique_index"`
	Email            string  `json:"email"`
	Password         string  `json:"password"`
	Mac              *string `json:"mac" gorm:"-"`
	RegistrationCode *string `json:"registrationCode"`
	SystemName       string  `json:"systemName" gorm:"-"`
}

type Query struct {
	Command    string  `json:"command"`
	Path       string  `json:"path"`
	Created    int64   `json:"created"`
	Uuid       string  `json:"uuid"`
	ExitStatus int     `json:"exitStatus"`
	Username   string  `json:"username"`
	SystemName string  `gorm:"-"  json:"systemName"`
	SessionID  *string `json:"sessionId"`
}

type Command struct {
	ProcessId        int    `json:"processId"`
	ProcessStartTime int64  `json:"processStartTime"`
	Uuid             string `json:"uuid"`
	Command          string `json:"command"`
	Created          int64  `json:"created"`
	Path             string `json:"path"`
	SystemName       string `json:"systemName"`
	ExitStatus       int    `json:"exitStatus"`
	User             User   `gorm:"association_foreignkey:ID"`
	UserId           uint
	Limit            int    `gorm:"-"`
	Unique           bool   `gorm:"-"`
	Query            string `gorm:"-"`
	SessionID        string `json:"sessionId"`
}

type System struct {
	ID            uint `json:"id" gorm:"primary_key"`
	Created       int64
	Updated       int64
	Mac           string  `json:"mac" gorm:"default:null"`
	Hostname      *string `json:"hostname"`
	Name          *string `json:"name"`
	ClientVersion *string `json:"clientVersion"`
	User          User    `gorm:"association_foreignkey:ID"`
	UserId        uint    `json:"userId"`
}

type Status struct {
	User                 `json:"-"`
	ProcessID            int    `json:"-"`
	Username             string `json:"username"`
	TotalCommands        int    `json:"totalCommands"`
	TotalSessions        int    `json:"totalSessions"`
	TotalSystems         int    `json:"totalSystems"`
	TotalCommandsToday   int    `json:"totalCommandsToday"`
	SessionName          string `json:"sessionName"`
	SessionStartTime     int64  `json:"sessionStartTime"`
	SessionTotalCommands int    `json:"sessionTotalCommands"`
}

type Import Query
