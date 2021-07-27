package config

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
)

type Config struct {
	Host            string
	Port            uint16
	SenderKeyPair   *ecdsa.PrivateKey
	ReceiverKeyPair *ecdsa.PublicKey
}

var Conf Config

func MkConfig(host string, port uint16, senderKey *ecdsa.PrivateKey, receiverKey *ecdsa.PublicKey) Config {
	return Config{
		Host:            host,
		Port:            port,
		SenderKeyPair:   senderKey,
		ReceiverKeyPair: receiverKey,
	}
}

var config_path string

func init() {
	flag.String("host", "127.0.0.1", "host ip to bind")
	flag.Uint16P("port", "p", 1336, "port used to bind")
	flag.String("sender_key", "", "private key used to send cmds to agent")
	flag.String("sender_key_file", "", "private key file used to send cmds to agent")
	flag.String("receiver_key", "", "public key used to validate incoming reqs")
	flag.String("receiver_key_file", "", "public key file used to validate incoming reqs")
	flag.StringVarP(&config_path, "config", "c", "nessaj_proxy.yaml", "config file path")
}

func chkKeyOpt(key string, keyName string, keyFile string, keyFileName string) error {
	if key == "" && keyFile == "" {
		return errors.New(fmt.Sprintf("%s and %s can not both be empty", keyName, keyFileName))
	}
	if key != "" && keyFile != "" {
		return errors.New(fmt.Sprintf("only one of %s and %s can be specified", keyName, keyFileName))
	}
	return nil
}

func getKeyContent(key string, keyFile string, ty string) (*pem.Block, error) {
	var keyContent []byte
	if key != "" {
		keyContent = []byte(key)
	} else {
		keyBytes, err := ioutil.ReadFile(keyFile)
		if err != nil {
			return nil, err
		}
		keyContent = keyBytes
	}
	parsedTy := ""
	var block *pem.Block
	for parsedTy != ty {
		block, keyContent = pem.Decode(keyContent)
		if block == nil {
			return nil, errors.New(fmt.Sprintf("can not find target PEM block %s", ty))
		}
		parsedTy = block.Type
	}
	return block, nil
}

func chkPrivateKey(key string, keyName string, keyFile string, keyFileName string) (*ecdsa.PrivateKey, error) {
	err := chkKeyOpt(key, keyName, keyFile, keyFileName)
	if err != nil {
		return nil, err
	}
	block, err := getKeyContent(key, keyFile, "EC PRIVATE KEY")
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

func chkPublicKey(key string, keyName string, keyFile string, keyFileName string) (*ecdsa.PublicKey, error) {
	err := chkKeyOpt(key, keyName, keyFile, keyFileName)
	if err != nil {
		return nil, err
	}
	block, err := getKeyContent(key, keyFile, "PUBLIC KEY")
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pub := pub.(type) {
	case *ecdsa.PublicKey:
		break
	default:
		return nil, errors.New(fmt.Sprintf("unexpected public key type %T", pub))
	}
	return pub.(*ecdsa.PublicKey), nil
}
func Parse() (*Config, error) {
	// env
	viper.SetEnvPrefix("nessaj_proxy")
	viper.BindEnv("host")
	viper.BindEnv("port")

	// flag
	flag.Parse()
	err := viper.BindPFlags(flag.CommandLine)
	if err != nil {
		return nil, err
	}

	// config file
	viper.SetConfigFile(config_path)
	viper.ReadInConfig() // ignore error

	senderKey := viper.GetString("sender_key")
	senderKeyFile := viper.GetString("sender_key_file")
	privKey, err := chkPrivateKey(senderKey, "sender_key", senderKeyFile, "sender_key_file")
	if err != nil {
		return nil, err
	}

	receiverKey := viper.GetString("receiver_key")
	receiverKeyFile := viper.GetString("receiver_key_file")
	pubKey, err := chkPublicKey(receiverKey, "receiver_key", receiverKeyFile, "receiver_key_file")
	if err != nil {
		return nil, err
	}

	config := MkConfig(viper.GetString("host"), uint16(viper.GetUint("port")),
		privKey, pubKey)

	Conf = config
	return &config, nil
}
