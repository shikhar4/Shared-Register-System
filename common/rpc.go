package common

type SetRequest struct {
	Key       string
	Val       string
	Timestamp int
	ClientID  int
}

type SetResponse struct {
	Ok bool
}

type GetRequest struct {
	Key string
}

type GetResponse struct {
	Exist     bool
	Val       string
	Timestamp int
}
