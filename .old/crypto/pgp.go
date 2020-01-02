//package crypto
//
//import (
//	"bytes"
//	"encoding/hex"
//	"fmt"
//	"github.com/arnisoph/postisto/pkg/log"
//	"github.com/jcmdev0/gpgagent"
//	"golang.org/x/crypto/openpgp"
//	"golang.org/x/crypto/openpgp/armor"
//	"io/ioutil"
//	"os"
//	"strings"
//)
//
//func Decrypt(encryptedText string, secretKeyPath string, secretKeyPassphrase string) (string, error) {
//	var err error
//
//	// Prepare message
//	decbuf := bytes.NewBuffer([]byte(encryptedText))
//	armorBlock, err := armor.Decode(decbuf)
//	if err != nil {
//		log.Errorw("Failed to decode armored PGP block", err, "secret-key", secretKeyPath)
//		return "", err
//	}
//
//	// Load secret key
//	keyringFileBuffer, err := os.Open(secretKeyPath)
//	if err != nil {
//		return "", err
//	}
//	defer keyringFileBuffer.Close()
//
//	var entityList openpgp.EntityList
//	if entityList, err = openpgp.ReadKeyRing(keyringFileBuffer); err != nil {
//		log.Errorw("Failed to load secret key (keyring)", err, "secret-key", secretKeyPath)
//		return "", err
//	}
//
//
//	// Decrypt secret key
//	//entity := entityList[0]
//	////entity.PrivateKey.Decrypt(passphraseByte)
//	//for _, subkey := range entity.Subkeys {
//	//	if err = subkey.PrivateKey.Decrypt([]byte(secretKeyPassphrase)); err != nil {
//	//		log.Errorw("Failed to decrypt secret key using user-defined key passphrase", err, "secret-key", secretKeyPath)
//	//		return "", err
//	//	}
//	//}
//
//	md, err := openpgp.ReadMessage(armorBlock.Body, entityList, promptFunction, nil)
//	if err != nil {
//		return "", err
//	}
//
//	//dec, err := base64.StdEncoding.DecodeString(string(ciphertext))
//	//if err != nil {
//	//	return "", err
//	//}
//	//md, err := openpgp.ReadMessage(bytes.NewBuffer(dec), entityList, prompt, nil)
//	//if err != nil {
//	//	return "", err
//	//}
//
//	if plainText, err := ioutil.ReadAll(md.UnverifiedBody); err != nil {
//		log.Errorw("Failed to parse decrypted message", err, "secret-key", secretKeyPath)
//		return "", err
//	} else {
//		return string(plainText), nil
//	}
//}
//
//func promptFunction(keys []openpgp.Key, symmetric bool) ([]byte, error) {
//	conn, err := gpgagent.NewGpgAgentConn()
//	if err != nil {
//		return nil, err
//	}
//	defer conn.Close()
//
//	for _, key := range keys {
//		cacheId := strings.ToUpper(hex.EncodeToString(key.PublicKey.Fingerprint[:]))
//		// TODO: Add prompt, etc.
//		request := gpgagent.PassphraseRequest{CacheKey: cacheId}
//		passphrase, err := conn.GetPassphrase(&request)
//		if err != nil {
//			return nil, err
//		}
//		err = key.PrivateKey.Decrypt([]byte(passphrase))
//		if err != nil {
//			return nil, err
//		}
//		return []byte(passphrase), nil
//	}
//	return nil, fmt.Errorf("unable to find key")
//}