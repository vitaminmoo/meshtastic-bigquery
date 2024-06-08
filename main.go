package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/vitaminmoo/meshtastic-bigquery/internal/zaplog"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	ctx := context.Background() // upstream mqtt package doesn't do context
	zl := zaplog.FromContext(ctx).With(
		zap.String("topic", msg.Topic()),
		// zap.Uint16("messageid", msg.MessageID()),
		// zap.Any("qos", msg.Qos()),
	)
	ctx = zaplog.WithLogger(ctx, zl)

	envelope := &meshtastic.ServiceEnvelope{}
	err := proto.Unmarshal(msg.Payload(), envelope)
	if err != nil {
		zl.Error("unmarshalling protobuf", zap.Error(err))
		return
	}

	j, err := json.Marshal(envelope)
	if err != nil {
		zl.Error("marshaling to json", zap.Error(err))
		return
	}
	fmt.Println(string(j))

	// fmt.Printf("encrypted: %x\n", envelope.Packet.GetDecoded().GetPayload())
	message, err := decrypt(envelope.Packet.From, envelope.Packet.Id, envelope.Packet.GetEncrypted())
	if err != nil {
		zl.Error("decrypting", zap.Error(err))
		return
	}

	fmt.Printf("decrypted: %s\n\n", message)

	data := &meshtastic.Data{}
	err = proto.Unmarshal(message, data)
	if err != nil {
		zl.Error("unmarshalling decrypted data", zap.Error(err))
		return
	}
	fmt.Printf("data: %0x\n", data)
}

func decrypt(from, id uint32, enc []byte) ([]byte, error) {
	//b64key := "AQ=="
	b64key := "1PG7OiApB1nwvP+rz05pAQ=="
	/*
	l := base64.StdEncoding.DecodedLen(len(b64key))
	var keyLen int
	if l <= 16 {
		keyLen = 16
	} else if l <= 32 {
		keyLen = 32
	} else {
		return nil, errors.New("invalid key length")
	}
	*/
	keyLen := 16
	key := make([]byte, keyLen)
	var err error
	_, err = base64.StdEncoding.Decode(key, []byte(b64key))
	if err != nil {
		return nil, fmt.Errorf("decoding key: %w", err)
	}

	//fmt.Printf("key: %0x\n", key)

	nonce_packet_id := make([]byte, 8)
	nonce_from_node := make([]byte, 8)
	binary.LittleEndian.PutUint32(nonce_from_node, from)
	binary.LittleEndian.PutUint32(nonce_packet_id, id)
	iv := append(nonce_packet_id, nonce_from_node...)
	//fmt.Printf("iv: %s\n", hex.EncodeToString(iv))

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}
	mode := cipher.NewCTR(block, iv)
	dec := make([]byte, 256)
	mode.XORKeyStream(dec, enc)
	return dec, nil
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	zap.L().Info("connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	zap.L().Info("connection lost", zap.Error(err))
}

func main() {
	zl := zaplog.ConfigureZapLogger()
	defer zl.Sync()

	broker := "homeassistant"
	port := 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername("meshtastic")
	// this isn't exposed to the internet, good luck
	opts.SetPassword("FDQ9flAlvybb0a")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sub(client)

	time.Sleep(3600 * time.Second)

	client.Disconnect(250)
}

func sub(client mqtt.Client) {
	topic := "meshtastic/2/e/LongFast/!da5ec1b8"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	zap.L().Debug("subscribed", zap.String("topic", topic))
}
