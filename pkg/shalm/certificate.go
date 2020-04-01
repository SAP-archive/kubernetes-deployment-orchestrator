package shalm

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/rickb777/date/period"

	"github.com/pkg/errors"
	"go.starlark.net/starlark"
	corev1 "k8s.io/api/core/v1"
)

type certificateBackend struct {
	privateKeyKey  string
	caKey          string
	certificateKey string
	validityPeriod period.Period
	isCa           bool
	signer         *vault
	domains        *starlark.List
}

var _ VaultBackend = (*certificateBackend)(nil)

func (c *certificateBackend) Name() string {
	return "certificate"
}

func (c *certificateBackend) Keys() map[string]string {
	return map[string]string{
		"certificate": c.certificateKey,
		"private_key": c.privateKeyKey,
		"ca":          c.caKey,
	}
}

func (c *certificateBackend) Apply(m map[string][]byte) (map[string][]byte, error) {
	if m[c.certificateKey] != nil {
		return m, nil
	}
	if c.isCa {
		return c.createCA()
	}
	return c.createCertificate()
}

func (c *certificateBackend) createCA() (map[string][]byte, error) {
	domains := listToStringArray(c.domains)
	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(1653),
		NotBefore:             time.Now(),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	ca.NotAfter, _ = c.validityPeriod.AddTo(ca.NotBefore)
	if len(domains) > 0 {
		ca.Subject.CommonName = domains[0]
	}

	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := &priv.PublicKey
	caCreated, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{
		c.certificateKey: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCreated}),
		c.privateKeyKey:  pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}),
	}, nil
}

func (c *certificateBackend) getSignerPEM(name string) ([]byte, error) {
	value, err := c.signer.Attr(name)
	if err != nil {
		return nil, err
	}
	stringValue, ok := value.(starlark.String)
	if !ok {
		return nil, errors.Errorf("Invalid type for signer attribute %s", name)
	}
	p, _ := pem.Decode([]byte(stringValue.GoString()))
	return p.Bytes, nil

}

func (c *certificateBackend) createCertificate() (map[string][]byte, error) {
	if c.signer == nil {
		return nil, errors.Errorf("Parameter signer required")
	}
	domains := listToStringArray(c.domains)
	if len(domains) == 0 {
		return nil, errors.Errorf("No domains given for certificates")
	}
	var caCert *x509.Certificate
	certificate, err := c.getSignerPEM("certificate")
	if err != nil {
		return nil, err
	}
	if caCert, err = x509.ParseCertificate(certificate); err != nil {
		return nil, err
	}
	var privateKey *rsa.PrivateKey
	privKey, err := c.getSignerPEM("private_key")
	if err != nil {
		return nil, err
	}
	if privateKey, err = x509.ParsePKCS1PrivateKey(privKey); err != nil {
		return nil, err
	}
	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: domains[0],
		},
		Issuer:       caCert.Subject,
		SerialNumber: big.NewInt(1658),
		NotBefore:    time.Now(),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		DNSNames:     domains,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	cert.NotAfter, _ = c.validityPeriod.AddTo(cert.NotBefore)
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := &priv.PublicKey

	// Sign the certificate
	certCreated, err := x509.CreateCertificate(rand.Reader, cert, caCert, pub, privateKey)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{
		c.certificateKey: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certCreated}),
		c.privateKeyKey:  pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}),
		c.caKey:          pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate}),
	}, nil

}

func makeCertificate(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
	c := &certificateBackend{
		privateKeyKey:  corev1.TLSPrivateKeyKey,
		caKey:          "ca.crt",
		certificateKey: corev1.TLSCertKey,
		isCa:           false,
	}
	var name string
	var err error
	var validity string = "P3M"
	if err = starlark.UnpackArgs("certificate", args, kwargs, "name", &name, "signer?", &c.signer, "is_ca?", &c.isCa, "domains?", &c.domains,
		"private_key_key?", &c.privateKeyKey, "ca_key?", &c.caKey, "certificate_key?", c.certificateKey, "validity?", &validity); err != nil {
		return nil, err
	}
	if c.validityPeriod, err = period.Parse(validity); err != nil {
		return nil, err
	}
	return NewVault(c, name)
}

func listToStringArray(list *starlark.List) []string {
	if list == nil {
		return nil
	}
	a := make([]string, starlark.Len(list))
	for i := 0; i < starlark.Len(list); i++ {
		a[i] = list.Index(i).(starlark.String).GoString()
	}
	return a

}
