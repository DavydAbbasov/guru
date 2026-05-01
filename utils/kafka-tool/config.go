package kafka

type Config struct {
	Brokers  []string
	ClientID string
	Topic    string
	GroupID  string
}
