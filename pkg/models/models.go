package models

type WebsocketMessage struct {
	UserId    string `json:"uid,omitempty"`
	RequestId string `json:"rid",omitempty`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	Active    string `json:"active,omitempty"`
}

type AccessRequest struct {
	Username   string `json:"username,omitempty"`
	Email      string `json:"email,omitempty"`
	Role       string `json:"role,omitempty"`
	TimePeriod string `json:"time_period,omitempty"`
	Cluster    string `json:"cluster,omitempty"`
}
