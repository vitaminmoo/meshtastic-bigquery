package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

/*
{"packet":{"from":4201362404,"to":4294967295,"channel":8,"PayloadVariant":{"Encrypted":"FaZAY0XkAYAGPrSlwLcPMnX6A2W0Yg=="},"id":1354433964,"rx_time":1717798516,"rx_snr":-17.25,"hop_limit":2,"rx_rssi":-128},"channel_id":"LongFast","gateway_id":"!da5ec1b8"}
{"packet":{"from":994490600,"to":4294967295,"channel":8,"PayloadVariant":{"Encrypted":"qUuilAxOUFPCc3k="},"id":2091761419,"rx_time":1717798520,"rx_snr":-16.75,"hop_limit":1,"rx_rssi":-127},"channel_id":"LongFast","gateway_id":"!da5ec1b8"}
{"packet":{"from":862220940,"to":4294967295,"channel":8,"PayloadVariant":{"Encrypted":"6oi5p0m7nv0Wao32nQZEakoZZjwyxfuvX2KfFL38WLI="},"id":1622496756,"rx_time":1717798529,"rx_snr":-17.75,"hop_limit":3,"rx_rssi":-128},"channel_id":"LongFast","gateway_id":"!da5ec1b8"}
{"packet":{"from":621114304,"to":146330472,"channel":8,"PayloadVariant":{"Encrypted":"ycQWiqWKJWKqLJ1LA46pZwgu0AW+/SpzOIm7Kwr3wT2M2G5GBVV3ps78DA29R1j9aWp+XET+QsHT"},"id":1274020173,"rx_time":1717798534,"rx_snr":-15.5,"hop_limit":1,"rx_rssi":-126},"channel_id":"LongFast","gateway_id":"!da5ec1b8"}
{"packet":{"from":3663889336,"to":146330472,"channel":8,"PayloadVariant":{"Encrypted":"+0pNJBLV4WTyhkTESZBhqSAxBte32dYk7wUGhBXzx1RG8pqujnlckQzAhozeIAVJ"},"id":308197114,"rx_time":1717798541,"rx_snr":-14,"hop_limit":4,"rx_rssi":-125},"channel_id":"LongFast","gateway_id":"!da5ec1b8"}
*/

func main() {
	b64enc := "FaZAY0XkAYAGPrSlwLcPMnX6A2W0Yg=="
	enc, err := base64.StdEncoding.DecodeString(b64enc)
	if err != nil {
		panic(err)
	}
	dec, err := decrypt(uint32(4201362404), uint32(1354433964), enc)
	if err != nil {
		panic(err)
	}
	fmt.Println(dec)
}

func strToKey(key string) []byte {
	buf := make([]byte, 16)
	for i := len(buf) - 1; seq != 0; i-- {
		buf[i] = byte(seq & 0xff)
		seq >>= 8
	}
	return buf
}

func decrypt(from, id uint32, enc []byte) ([]byte, error) {
	key := []byte("AQ==")
	fmt.Printf("%0x\n", key)

	nonce_packet_id := make([]byte, 8)
	nonce_from_node := make([]byte, 8)
	binary.LittleEndian.PutUint32(nonce_from_node, from)
	binary.LittleEndian.PutUint32(nonce_packet_id, id)
	iv := append(nonce_packet_id, nonce_from_node...)
	fmt.Printf("iv: %s\n", hex.EncodeToString(iv))

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}
	mode := cipher.NewCTR(block, iv)
	dec := make([]byte, 256)
	mode.XORKeyStream(dec, enc)
	return dec, nil
}
