package domain

import "time"

type User struct {
	Id       int64
	Username string
	Account  string
	Password string
	Phone    string
	Nickname string
	WeightKG int
	HeightCM int
	Birthday time.Time
	AboutMe  string
	Ctime    time.Time
	Utime    time.Time
}
