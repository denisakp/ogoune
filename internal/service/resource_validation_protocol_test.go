package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(i int) *int { return &i }

func TestValidateProtocolFields_AllValidTypes(t *testing.T) {
	for _, pt := range []string{"redis", "mongodb", "ftp", "ssh", "mysql", "postgres", "rabbitmq"} {
		err := validateProtocolFields(strPtr(pt), intPtr(1234), "broker.local")
		require.NoError(t, err, "protocol %s should validate", pt)
	}
	// kafka requires bootstrap CSV in target and no protocol_port
	err := validateProtocolFields(strPtr("kafka"), nil, "k1:9092,k2:9092")
	require.NoError(t, err)
}

func TestValidateProtocolFields_RejectsUnknown(t *testing.T) {
	err := validateProtocolFields(strPtr("amqp"), nil, "h:5672")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrValidationFailed))
}

func TestValidateProtocolFields_RabbitMQ_RejectsCommas(t *testing.T) {
	err := validateProtocolFields(strPtr("rabbitmq"), nil, "h1,h2")
	require.Error(t, err)
}

func TestValidateProtocolFields_Kafka_RejectsPort(t *testing.T) {
	err := validateProtocolFields(strPtr("kafka"), intPtr(9092), "k1:9092")
	require.Error(t, err)
}

func TestValidateProtocolFields_Kafka_RejectsEmptyBootstrap(t *testing.T) {
	err := validateProtocolFields(strPtr("kafka"), nil, "")
	require.Error(t, err)
}

func TestValidateProtocolFields_Kafka_RejectsMalformedEntry(t *testing.T) {
	err := validateProtocolFields(strPtr("kafka"), nil, "k1:9092,nojustport")
	require.Error(t, err)
}

func TestParseKafkaBootstrapTargets_Normalizes(t *testing.T) {
	got, err := parseKafkaBootstrapTargets(" k1:9092 , k2:9092 ")
	require.NoError(t, err)
	assert.Equal(t, []string{"k1:9092", "k2:9092"}, got)
}
