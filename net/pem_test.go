package net

import (
	"crypto/tls"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	configutil"github.com/stuphlabs/pullcord/config/util"
)

// example key and cert stolen from
// https://github.com/freelan-developers/freelan/tree/d248b/samples/fscp/client
const exampleKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`
const exampleCert = `-----BEGIN CERTIFICATE-----
MIIF0DCCA7igAwIBAgIBATANBgkqhkiG9w0BAQUFADB2MQswCQYDVQQGEwJGUjEP
MA0GA1UECAwGQWxzYWNlMRMwEQYDVQQHDApTdHJhc2JvdXJnMRAwDgYDVQQKDAdG
cmVlbGFuMQswCQYDVQQDDAJjYTEiMCAGCSqGSIb3DQEJARYTY29udGFjdEBmcmVl
bGFuLm9yZzAeFw0xMjA1MDUxMjQ3MTFaFw0yMjA1MDMxMjQ3MTFaMGQxCzAJBgNV
BAYTAkZSMQ8wDQYDVQQIDAZBbHNhY2UxEDAOBgNVBAoMB0ZyZWVsYW4xDjAMBgNV
BAMMBWFsaWNlMSIwIAYJKoZIhvcNAQkBFhNjb250YWN0QGZyZWVsYW4ub3JnMIIC
IjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA47QNFNFb5Z0BZx66+RxhSlWr
nIDHmzXaT/UNJuJFgj5cAPNHtNq9s/g7Id5qeL2OV96be0AH03P1pL3ahXSRxubR
6G6lIenU9vn/j50JZlrevewE/rJP6u6AMOsoO8xKqWF2m1qWdSQc2LVZgdBkWbnw
PfS0i5kU9LyiSxzWzBREUV/Gtjw6+6+y6yzrpWzi4QL9Ojf4vuhvmZLpZG7qRXQv
h/qcCMYM6fgannZkUkhnoL5ljrEi5wI7mzvqDbuUGYOKp8rYTG5Qu+lpZku76Ih/
u5A2DPRQ6JPJmnf569O4of+XjmcxzSmzkpoY5NE9aB3ZJQ1fl8t3ClMzbMJ79Oik
vVmKyMPf97hvdBPcmnrzzgF3ePCV7UlvTC5JOELWDDEdhoLIQR6Rnh40V+gzvP0d
Boyvs6fyrcthDanbxOEWg0hBQxdj58E4YvFu09B3WdlasuNImB3kH/UhB4E3dSDM
kL4+soD2aTs11ZiTMZkpxuTyVb7l6CZC6aAOeJax7sBnrte0FuDTWm/6xThjlfL6
elW1k37IioVaOIn8Qw50EIsKpmbpWNXnYZj3h+qEumtjzXsvkU1pKHtGQqrSgfbO
n0x1edgLt/ByJFLOF1fdgLrgiUXTOw+Q0sQNy+46dTfG2a6I1FmST1PX/Fk4qkzI
eBOFHACJGReqlFwSw6sCAwEAAaN7MHkwCQYDVR0TBAIwADAsBglghkgBhvhCAQ0E
HxYdT3BlblNTTCBHZW5lcmF0ZWQgQ2VydGlmaWNhdGUwHQYDVR0OBBYEFFg1MTTD
/IfnxNdVLyiphgEsQBxnMB8GA1UdIwQYMBaAFEKcNr3LL/TZXqsrGCSLAS1CwuFH
MA0GCSqGSIb3DQEBBQUAA4ICAQABWb/1QUiu1uGkL1/VvYvQWdMTkZ8jMOb8AWUw
RlCnLmrFb5uqZQzhNsXJMKZNDIFdqhXNwJxXmq58a7lF30HUFhezVGLIpVTb3Eqs
dn4ixW/8Zpy4S8Z1SvKo8ROo5YNpsJJLYp4SDOFfg290jlzIqwy7Prv8VTzXiNXU
4y7pVFKgFZOwVctvjk9/f9vKvNhvljttgmECVTIwuP/qV9xJSNE+R/GAkkEB25WV
yFi5slnJ4jcHhM/zvqvtOwPVNmGBMkbLOy0oXqsqEoRj8WUobqxoCkdkrdXjgPv2
/iT3CHx1H8StYyaSK+j5SkVnqoRDIkQkgjeHyeHm6jngGiEosZT1PV4eECQcxiLH
zB1MrMk63a/cDsNDOMPTj6y+xSsrMOTY7TxNvjVyQU4Zu6QVXhYh19zEunfvvvG7
wfacwKv6JZ4TlxVFtb0DP9oRUkRZaSlNpRTKqiNjF+e2OkW0IS9OfoBnnhU0cf3R
SUbY6R96mqRQnbRgcpdOUvgDVcIzckizXDd5rleen6vD/OB2W+yEFj2+X4rVVbYQ
mtiMRaeQnHpWbIH2YnQikWKffZlsZGnW0x99+6P1bEp1NdV7tajjxXECFlnOW+cS
77OBJOhvu4hOMkn8qZY5cHGjVKprS6qm5oBzorAO1iw8TMOYFNAMxCVWuC/67z0b
1BOqnA==
-----END CERTIFICATE-----`

func TestPem(t *testing.T) {
	testCases := map[string]struct{
		inputCert []byte
		inputKey []byte
		inputHello *tls.ClientHelloInfo
		errorExpected bool
	} {
		"nil values": {
			inputCert: nil,
			inputKey: nil,
			inputHello: nil,
			errorExpected: true,
		},
		"example cert": {
			inputKey: []byte(exampleKey),
			inputCert: []byte(exampleCert),
			inputHello: nil,
			errorExpected: false,
		},
	}

	for explanation, test := range testCases {
		pemConfig := PemConfig{
			test.inputCert,
			test.inputKey,
		}
		_, actualError := pemConfig.GetCertificate(test.inputHello)

		if test.errorExpected {
			assert.Errorf(
				t,
				actualError,
				"Expected an error from"+
					" net.PemConfig.GetCertificate for:"+
					" %s",
				explanation,
			)
		} else {
			assert.NoErrorf(
				t,
				actualError,
				"Expected no error from"+
					" net.PemConfig.GetCertificate for:"+
					" %s",
				explanation,
			)
		}
	}
}

func escnl(s string) string {
	return strings.Replace(s, "\n", `\n`, -1)
}

func TestPemConfig(t *testing.T) {
	test := configutil.ConfigTest{
		ResourceType: "pem",
		ListenerTest: true,
		SyntacticallyBad: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: ``,
				Explanation: "empty config",
			},
			configutil.ConfigTestData{
				Data: `42`,
				Explanation: "numeric config",
			},
			configutil.ConfigTestData{
				Data: `"test"`,
				Explanation: "string config",
			},
			configutil.ConfigTestData{
				Data: `{
					"key": 7,
					"cert": "`+escnl(exampleCert)+`"
				}`,
				Explanation: "bad key",
			},
			configutil.ConfigTestData{
				Data: `{
					"key": "`+escnl(exampleKey)+`",
					"cert": {
						"val": "`+escnl(exampleCert)+`"
					}
				}`,
				Explanation: "bad cert",
			},
		},
		Good: []configutil.ConfigTestData{
			configutil.ConfigTestData{
				Data: `{
				}`,
				Explanation: "empty object",
			},
			configutil.ConfigTestData{
				Data: `{
					"key": "`+escnl(exampleKey)+`",
					"cert": "`+escnl(exampleCert)+`"
				}`,
				Explanation: "good config",
			},
			configutil.ConfigTestData{
				Data: `{
					"key": "`+escnl(exampleKey)+`"
				}`,
				Explanation: "missing cert",
			},
			configutil.ConfigTestData{
				Data: `{
					"cert": "`+escnl(exampleCert)+`"
				}`,
				Explanation: "missing key",
			},
		},
	}

	test.Run(t)
}


