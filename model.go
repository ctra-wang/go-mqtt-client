package gm_client

type MqttClient struct {
	Broker   string `json:"broker"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	ClientID string `json:"clientID,optional"`
	Qos      int    `json:"qos,optional"`
	CrtPem   string `json:"crtPem"`
	KeyPem   string `json:"keyPem"`
	Ca       string `json:"ca,optional"`
}
