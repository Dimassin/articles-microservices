package config

type Config struct {
	Kafka KafkaConfig
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}
