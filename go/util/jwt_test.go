package util

import (
	"crypto/rsa"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

func Test_PubKeyJwtAuthorizer_decode(t *testing.T) {
	verifyBytes := []byte(cert)

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}
	token2049 := jwt2049

	tokenExpired := jwtExpired

	type fields struct {
		pubKey *rsa.PublicKey
	}
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *JwtData
		wantErr bool
	}{
		{
			name:    "2049",
			fields:  fields{verifyKey},
			args:    args{string(token2049)},
			want:    &JwtData{Sub: "massimiliano.muzi", Iss: "goa/ad/corp"},
			wantErr: false,
		},
		{
			name:    "expired",
			fields:  fields{verifyKey},
			args:    args{string(tokenExpired)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &PubKeyJwtAuthorizer{
				pubKey: tt.fields.pubKey,
			}
			got, err := v.decode(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("jwtDecoder.decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && *got != *tt.want {
				t.Errorf("jwtDecoder.decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

var cert = `
-----BEGIN CERTIFICATE-----
MIIEwTCCA6mgAwIBAgIJAOxc7QrlayIpMA0GCSqGSIb3DQEBCwUAMIGbMQswCQYD
VQQGEwJJVDEOMAwGA1UECBMFTWlsYW4xDjAMBgNVBAcTBU1pbGFuMRAwDgYDVQQK
EwdJcmlkZW9zMRAwDgYDVQQLEwdJcmlkZW9zMRAwDgYDVQQDEwdJcmlkZW9zMQ8w
DQYDVQQpEwZTYWJhdG8xJTAjBgkqhkiG9w0BCQEWFmtzc28tc2FiYXRvQGRldi5r
cWkuaXQwHhcNMTkwOTE3MTUzMTA4WhcNMTkxMDE3MTUzMTA4WjCBmzELMAkGA1UE
BhMCSVQxDjAMBgNVBAgTBU1pbGFuMQ4wDAYDVQQHEwVNaWxhbjEQMA4GA1UEChMH
SXJpZGVvczEQMA4GA1UECxMHSXJpZGVvczEQMA4GA1UEAxMHSXJpZGVvczEPMA0G
A1UEKRMGU2FiYXRvMSUwIwYJKoZIhvcNAQkBFhZrc3NvLXNhYmF0b0BkZXYua3Fp
Lml0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyL602GFDCtO+MhBL
YkQJ4GHPkV0ukx5yhEO6Hhjcu2AqwJ3LAJCwOsxg9lw/bgy5UqYI5BZyvj6/CkVw
axP8tm5NzLEuAQeWMpBIi5xOsYv471cMhUHootvE4l93eLFqKA9taLEcaBcwlTAD
dzJNzPHqc5sCYvLUOZKKYR3i2/8uGV4MsEDjPNMOlHGCeMobZa5MVz0H55XplINT
j2lFzCa2C8B6MezEIsjf343FC1L1NwuONG/4RiQUZ/EEkSSYJG9k7G7tFTYz2ECX
4GwS0eZOv6fCknXaHBwYCA63Jd8M6zCvKzJdUq0iNl/e5pjaNonHb2J+Q0gI7zxw
3nAn4wIDAQABo4IBBDCCAQAwHQYDVR0OBBYEFAwa0z0xI30lT8lmmWbeSxhgEbn+
MIHQBgNVHSMEgcgwgcWAFAwa0z0xI30lT8lmmWbeSxhgEbn+oYGhpIGeMIGbMQsw
CQYDVQQGEwJJVDEOMAwGA1UECBMFTWlsYW4xDjAMBgNVBAcTBU1pbGFuMRAwDgYD
VQQKEwdJcmlkZW9zMRAwDgYDVQQLEwdJcmlkZW9zMRAwDgYDVQQDEwdJcmlkZW9z
MQ8wDQYDVQQpEwZTYWJhdG8xJTAjBgkqhkiG9w0BCQEWFmtzc28tc2FiYXRvQGRl
di5rcWkuaXSCCQDsXO0K5WsiKTAMBgNVHRMEBTADAQH/MA0GCSqGSIb3DQEBCwUA
A4IBAQB1ubkbXDaBTVVA/B9VoKSAdwdG4xyNb+P7HBqm1yt4w6C0ZLQrehEmlJqt
fB8PhRaFtypAn6f9k+F64LrP+Vdc+PnJ5jdWbkPaM6kt9K4fGHSHSVlAFCYIK3yX
UJl+mDqf8nkeRSFDFbGjdcSGDCpEfoMaZs/ZoJ5E0qoaOu0xvYKm0k/3k6N5+j8v
wjQ4U1A7X9qrfyVNRQVqSVHi4ARhCViBvCUoZgRo9YPkNtrmH41GVHPvF4bkuWEm
ESu72pTZpuaIKoWaojMionP1MX2byzaOLkiI9rrFkLRc7TWLm62UzPrBeke7+iJ+
Yz6UN0rSYwDSFX667znY1sgVfm1B
-----END CERTIFICATE-----
`

const jwt2049 = `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiTXV6aSBNYXNzaW1pbGlhbm8iLCJzdWIiOiJtYXNzaW1pbGlhbm8ubXV6aSIsInNpZCI6ImFMcHFWbHVKajBTQmJIcmIzZ2NRdHciLCJpc3MiOiJnb2EvYWQvY29ycCIsImV4cCI6MjUxNTUwNjM5NC4wfQ.CHVy4a1o471tmZj0Qf3rc2mMEhKZbDzVHfemlr2gKZPkGJAt--HgdglBiR3qMkaCYgsn82KJnKwufi5K-mR3n_bAtsoH9l2wyvrtsLkg_EYdqcyS7GU0dYJH3F9hwkIfPIxEz2Or2FNsr19C7oPQTcoBWsYdKVFo-EJrs9wvAuwZEwf_XTO118LFtt7EW_jjvtI9RaOlo4FrcRT8CoqtTA9-RUuWgOC_bao2cjXJgeczTE4HvpgMlAAxObUxLcD68SPE2hk_Gb4JXVvwT-bPU-cqUfNgoY5rTDbeVy5bMdGTBr__U4g5UQCI47WQqU3y2yBfzkAcpxX1mgnUeH5pOg`
const jwtExpired = `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiTXV6aSBNYXNzaW1pbGlhbm8iLCJzdWIiOiJtYXNzaW1pbGlhbm8ubXV6aSIsInNpZCI6ImFMcHFWbHVKajBTQmJIcmIzZ2NRdHciLCJpc3MiOiJnb2EvYWQvY29ycCIsImV4cCI6MTU2ODczNDk1My4wfQ.JdxFt7rFlnuc0PBUHwcv6O76zMAkTPntAlgE68qCYUyBMu3Czx-FvzTlUns6AsOWQS_xtg_kKnEPRbpG-DoCQiN3CN2kvHSbiCDHAx5v_ac3pt5FYhnAMKN7zEWfs8Yec92wQj0bESrm6Wr0XJAEXoo9cdIgeakMUrnGXbSy9mhK1T7_o9LzpNArKmSNzogPSmKa4N_TZ7Vvczses-OxXMO8-71uVk4jJa70bRPyf2W99b5xPW3sqn0Xy35JSXXLK-suseNBxnO5KoqhaVioZ0nz4dhrqeOsVpW54kH6HV_nfzEiehnxhhTGu5eHrxJC0hSmvdIhythLOG9BMgCaDQ`
