package util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

func EncryptString(str, keyID, region string) ([]byte, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := kms.New(sess, aws.NewConfig().WithRegion(region))
	input := &kms.EncryptInput{
		KeyId:     aws.String(keyID),
		Plaintext: []byte(str),
	}
	res, err := svc.Encrypt(input)
	if err != nil {
		return nil, err
	}
	return res.CiphertextBlob, nil
}
