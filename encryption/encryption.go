package encryption

import (
	"crypto/sha256"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/ldsec/lattigo/v2/rlwe"
)

type Client struct {
	bfv.Encoder
	bfv.Encryptor
	bfv.Decryptor

	Params bfv.Parameters
}

type Server struct {
	Participants []string
	bfv.Evaluator
	Responses []*bfv.Ciphertext
	Result    *bfv.Ciphertext
}

func NewClient() Client {

	client := Client{}

	p := bfv.DefaultParams[1]
	params, _ := bfv.NewParametersFromLiteral(p)
	client.Params = params
	//keyGenerator := bfv.NewKeyGenerator(params)
	client.Encoder = bfv.NewEncoder(params)

	sk := bfv.NewSecretKey(params)
	client.Encryptor = bfv.NewEncryptor(params, sk)
	client.Decryptor = bfv.NewDecryptor(params, sk)

	return client
}

func NewServer() Server {

	p := bfv.DefaultParams[1]
	params, _ := bfv.NewParametersFromLiteral(p)
	evaluationKey := rlwe.EvaluationKey{
		Rlk: bfv.NewRelinearizationKey(params, 2),
	}
	return Server{
		Evaluator:    bfv.NewEvaluator(params, evaluationKey),
		Responses:    make([]*bfv.Ciphertext, 0),
		Participants: make([]string, 0),
	}
}

func (client *Client) Encrypt(values ...int64) (string, error) {
	text := bfv.NewPlaintext(client.Params)

	coeffs := make([]int64, 0)
	for _, v := range values {
		coeffs = append(coeffs, v)
	}
	client.EncodeInt(coeffs, text)
	encrypted := client.EncryptNew(text)
	output := MarshalToBase64String(encrypted)

	return output, nil
}

func (client *Client) Decrypt(input string) ([]int64, error) {
	cipher := bfv.NewCiphertext(client.Params, 1)
	UnmarshalFromBase64(cipher, input)

	text := client.DecryptNew(cipher)
	coeffs := make([]int64, 2)
	for i, v := range client.DecodeIntNew(text)[:2] {
		coeffs[i] = v
	}
	return coeffs, nil
}

// ------------ UTILS BELOW ------------

// MarshalToBase64String returns serialization of a marshallable type as a base-64-encoded string
func MarshalToBase64String(bm encoding.BinaryMarshaler) string {
	if bm == nil || reflect.ValueOf(bm).IsNil() {
		return "nil"
	}
	b, err := bm.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// UnmarshalFromBase64 reads a base-64 string into a unmarshallable type
func UnmarshalFromBase64(bum encoding.BinaryUnmarshaler, b64string string) error {
	b, err := base64.StdEncoding.DecodeString(b64string)
	if err != nil {
		return err
	}
	return bum.UnmarshalBinary(b)
}

// GetSha256Hex returns an hexadecimal string representation of the Sha256 hash of marshallable type
func GetSha256Hex(bm encoding.BinaryMarshaler) string {
	b, _ := bm.MarshalBinary()
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

// PublicDataJSON returns the public state of the poll struct as a JSON encoding Lattigo objects in base64.
func (s *Server) PublicDataJSON() string {
	b, _ := json.Marshal(map[string]interface{}{
		"result": MarshalToBase64String(s.Result),
	})
	return string(b)
}

func (s *Server) AddParticipants(ps ...string) {
	for _, p := range ps {
		s.Participants = append(s.Participants, p)
	}
}
