//package crypto_test
//
//import (
//	"github.com/arnisoph/postisto/pkg/crypto"
//	"github.com/stretchr/testify/require"
//	"testing"
//)
//
//func TestDecrypt(t *testing.T) {
//	require := require.New(t)
//
//	// ACTUAL TESTS BELOW
//
//	// gpg --no-default-keyring --keyring ./ring.gpg --trustdb ./trustdb.gpg --fingerprint
//	// gpg --no-default-keyring --keyring ./ring.gpg --trustdb ./trustdb.gpg --full-generate-key
//	// gpg --no-default-keyring --keyring ./ring.gpg --trustdb ./trustdb.gpg -k
//	// gpg --no-default-keyring --keyring ./ring.gpg --trustdb ./trustdb.gpg  -K
//
//	// echo -n secret-password | gpg --no-default-keyring --keyring ./ring.gpg --trustdb ./trustdb.gpg --encrypt --recipient F4012A8814AD95226FC278D5CC6ADC1CA5C79217 --armor --output secret-password.asc
//
//	//  gpg --no-default-keyring --keyring ./ring.gpg --export-secret-keys > fuckring.gpg
//
//	encryptedText := `
//-----BEGIN PGP MESSAGE-----
//
//hQEMA7GdgqeHCnC5AQgAoXYv5mTQw+Jd4zhNoD7KdLPxTETPSORUHVxxjLyrKS2b
//hAsxXixUOyfsBdbC9eLiUPAvtvHIoiXgSMt0Fb11p86NSfTwzoKGaRRuXgLqPgEq
//bX0mRHjiWEie4uViOfyipPK7qUJ9UvpyoxFvcW1n3zfRUIgEId4mHfOJiBG3gUJy
//Q9mjEo7dRCboUGKyCQwkt/qyQZSJRSfvNjhCn2NjCySqNpBRSiTYFdocZ/0bOmq/
//vHJ9ZsZLfMQZCJ6C6uxyTqzMw5cVh+HiM3wnjrHdjkUXimk+PipMmlel32E+btOS
//yEWZ87PvDHDxQZ3uwVaKqIo4xcznN3EcM0ba7RW1htJKAasRwH0Kd30vjJyMZ0Dp
///esGckhTMY+e3euoZ5u+x95gKngLMSzGES1tOI6MNbC7PyJ1zT9sOFhJVH81ys1V
//Wf4tvDkbfcrsOSM=
//=Xbun
//-----END PGP MESSAGE-----
//`
//
//	plainText, err := crypto.Decrypt(encryptedText, "../../test/data/pgp/secret-key.gpg", "pass123")
//	require.NoError(err)
//	require.Equal("secret-password", plainText)
//}
